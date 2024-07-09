package sriov

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/kindbase"
)

type KindSriov struct {
	pfs            []string
	pfCountPerNode int

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
	pfOffset := 0

	controlPlaneAdapter := docker.NewDockerAdapter(cli, ks.Version+"-control-plane")

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

// func (ks *KindSriov) createVfsOnNode(sshClient utils.SSHClient) error {
// 	cmds := []string{}
// 	return nil
// }

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
