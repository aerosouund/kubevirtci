package sriov

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/kindbase"
)

type KindSriov struct {
	pfs            []string
	pfCountPerNode int
	vfsCount       int

	*kind.KindBaseProvider
}

func NewKindSriovProvider(kindConfig *kind.KindConfig) (*KindSriov, error) {
	kindBase, err := kind.NewKindBaseProvider(kindConfig)
	if err != nil {
		return nil, err
	}
	return &KindSriov{
		KindBaseProvider: kindBase,
	}, nil
}

func (ks *KindSriov) Start(ctx context.Context, cancel context.CancelFunc) error {
	devs, err := ks.discoverHostPFs()
	if err != nil {
		return err
	}
	ks.pfs = devs

	if ks.Nodes*ks.pfCountPerNode > len(devs) {
		return fmt.Errorf("Not enough virtual functions available, there are %d functions on the host", len(devs))
	}

	if err = ks.KindBaseProvider.Start(ctx, cancel); err != nil {
		return err
	}

	nodes, err := ks.Provider.ListNodes(ks.Version)
	if err != nil {
		return err
	}

	// fix this by adding the ssh interface to the cri interface
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	controlPlaneAdapter := docker.NewDockerAdapter(cli, ks.Version+"-control-plane")

	pfOffset := 0
	for _, node := range nodes {
		nodeName := node.String()
		da := docker.NewDockerAdapter(cli, nodeName)
		nodeJson := []types.ContainerJSON{}
		resp, err := ks.CRI.Inspect(nodeName)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resp, &nodeJson)
		if err != nil {
			return err
		}

		pid := nodeJson[0].State.Pid
		if err = ks.linkNetNS(pid, nodeName); err != nil {
			return err
		}

		pfsForNode := ks.pfs[pfOffset : pfOffset+ks.pfCountPerNode]
		if err = ks.assignPfsToNode(pfsForNode, nodeName); err != nil {
			return err
		}

		pfOffset += ks.pfCountPerNode
		cmds := []string{
			"mount -o remount,rw /sys",
			"ls -la -Z /dev/vfio",
			"chmod 0666 /dev/vfio/vfio",
		}

		for _, cmd := range cmds {
			if _, err := da.SSH(cmd, true); err != nil {
				return err
			}
		}

		if _, err = controlPlaneAdapter.SSH("kubectl label node "+nodeName+" sriov_capable=true", true); err != nil {
			return err
		}

	}
	return nil
}

func (ks *KindSriov) createVfsOnNode(da docker.DockerAdapter) error {
	sysfs, err := da.SSH(`grep -Po 'sysfs.*\K(ro|rw)' /proc/mounts`, false)
	if err != nil {
		return nil
	}
	if sysfs != "rw" {
		return fmt.Errorf("FATAL: sysfs is read-only, try to remount as RW")
	}

	mod, err := da.SSH(`grep vfio_pci /proc/modules`, false)
	if err != nil {
		return nil
	}

	if _, err = da.SSH("modprobe -i vfio_pci", true); err != nil {
		return err
	}

	if mod == "" {
		return fmt.Errorf("System doesn't have the vfio_pci module, provisioning failed")
	}

	pfsString, err := da.SSH(`find /sys/class/net/*/device/sriov_numvfs`, false)
	if err != nil {
		return nil
	}
	pfs := strings.Split(pfsString, " ")
	if len(pfs) == 0 {
		return fmt.Errorf("No physical functions found on node, exiting")
	}

	for _, pf := range pfs {
		pfDevice, err := da.SSH("dirname "+pf, false)
		if err != nil {
			return err
		}

		vfsSysFsDevices, err := ks.createVFsforPF(da, pfDevice)
		for _, vfDevice := range vfsSysFsDevices {
			err = ks.bindToVfio(da, vfDevice)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ks *KindSriov) createVFsforPF(sshClient docker.DockerAdapter, id string) ([]string, error) {
	pfName, err := sshClient.SSH("basename "+id, false)
	if err != nil {
		return nil, err
	}

	pfSysFsDevice, err := sshClient.SSH("readlink -e "+id, false)
	if err != nil {
		return nil, err
	}
	totalVfs, err := sshClient.SSH("cat "+pfSysFsDevice+"/sriov_totalvfs", false)
	if err != nil {
		return nil, err
	}

	totalVfsCount, err := strconv.Atoi(totalVfs)
	if err != nil {
		return nil, err
	}

	if totalVfsCount < ks.vfsCount {
		return nil, fmt.Errorf("FATAL: PF %s, VF's count should be up to sriov_totalvfs: %d", pfName, totalVfsCount)
	}

	cmds := []string{
		"echo 0 >> " + pfSysFsDevice + "/sriov_numvfs",
		"echo " + totalVfs + " >> " + pfSysFsDevice + "/sriov_numvfs",
	}

	for _, cmd := range cmds {
		if _, err := sshClient.SSH(cmd, true); err != nil {
			return nil, err
		}
	}

	vfsString, err := sshClient.SSH(`readlink -e `+pfName+`/virtfn*`, false)
	if err != nil {
		return nil, err
	}

	return strings.Split(vfsString, " "), nil
}

func (ks *KindSriov) bindToVfio(sshClient docker.DockerAdapter, sysFsDevice string) error {
	devSysfsPath, err := sshClient.SSH("basename "+sysFsDevice, false)
	if err != nil {
		return err
	}

	driverPath := devSysfsPath + "/driver"
	driverOverride := devSysfsPath + "/driver_override"

	vfBusPciDeviceDriver, err := sshClient.SSH("readlink "+driverPath+" | awk -F'/' '{print $NF}'", false)
	if err != nil {
		return err
	}
	vfBusPciDeviceDriver = strings.TrimSuffix(vfBusPciDeviceDriver, "\n")
	vfDriverName, err := sshClient.SSH("basename "+vfBusPciDeviceDriver, false)
	if err != nil {
		return err
	}

	if _, err := sshClient.SSH("modprobe -i vfio-pci", false); err != nil {
		return fmt.Errorf("Error loading vfio-pci module: %v", err)
	}

	for i := 0; i < 10; i++ {
		if _, err := sshClient.SSH("ls /sys/bus/pci/drivers/vfio-pci", false); err != nil {
			fmt.Println("module not loaded properly, sleeping 1 second and trying again")
			time.Sleep(time.Second * 1)
			sshClient.SSH("modprobe -i vfio-pci", false)
		} else {
			break
		}
	}

	cmds := []string{
		"[[ '" + vfDriverName + "' != 'vfio-pci' ]] && echo " + devSysfsPath + " > " + driverPath + "/unbind && echo 'vfio-pci' > " + driverOverride + " && echo " + devSysfsPath + " > /sys/bus/pci/drivers/vfio-pci/bind",
	}

	for _, cmd := range cmds {
		if _, err := sshClient.SSH(cmd, true); err != nil {
			return err
		}
	}

	return nil
}

func (ks *KindSriov) assignPfsToNode(pfs []string, nodeName string) error {
	for _, pf := range pfs {
		cmds := []string{
			"link set " + pf + " netns " + nodeName,
			"netns exec " + nodeName + " ip link set up dev " + pf,
			"netns exec " + nodeName + " ip link show",
		}
		for _, cmd := range cmds {
			cmd := exec.Command("ip", cmd)
			if _, err := cmd.Output(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ks *KindSriov) linkNetNS(pid int, nodeName string) error {
	cmd := exec.Command("ln", "-sf", "/proc/"+fmt.Sprintf("%d", pid)+"/ns/net", "/var/run/netns/"+nodeName)
	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func (ks *KindSriov) discoverHostPFs() ([]string, error) {
	files, err := filepath.Glob("/sys/class/net/*/device/sriov_numvfs")
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errors.New("FATAL: Could not find available sriov PFs on host")
	}

	pfNames := make([]string, 0)
	for _, file := range files {
		pfName := filepath.Base(filepath.Dir(filepath.Dir(file)))
		pfNames = append(pfNames, pfName)
	}

	return pfNames, nil
}
