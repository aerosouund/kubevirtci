package bindvfio

import (
	"fmt"
	"strings"
	"time"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type BindVfioOpt struct {
	sshPort   uint16
	nodeIdx   int
	pciID     string
	sshClient utils.SSHClient
}

func NewBindVfioOpt(sshClient utils.SSHClient, sshPort uint16, nodeIdx int, id string) *BindVfioOpt {
	return &BindVfioOpt{
		sshPort:   sshPort,
		nodeIdx:   nodeIdx,
		pciID:     id,
		sshClient: sshClient,
	}
}

func (o *BindVfioOpt) Exec() error {
	addr, err := o.sshClient.SSH("lspci -D -d "+o.pciID, false)
	if err != nil {
		return err
	}

	pciDevId := strings.Split(addr, " ")[0]

	devSysfsPath := "/sys/bus/pci/devices/" + pciDevId
	driverPath := devSysfsPath + "/driver"
	driverOverride := devSysfsPath + "/driver_override"

	driver, err := o.sshClient.SSH("readlink "+driverPath+" | awk -F'/' '{print $NF}'", false)
	if err != nil {
		return err
	}
	driver = strings.TrimSuffix(driver, "\n")

	if _, err := o.sshClient.SSH("modprobe -i vfio-pci", false); err != nil {
		return fmt.Errorf("Error loading vfio-pci module: %v", err)
	}

	for i := 0; i < 10; i++ {
		if _, err := o.sshClient.SSH("ls /sys/bus/pci/drivers/vfio-pci", false); err != nil {
			fmt.Println("module not loaded properly, sleeping 1 second and trying again")
			time.Sleep(time.Second * 1)
			o.sshClient.SSH("modprobe -i vfio-pci", false)
		} else {
			break
		}
	}

	cmds := []string{
		"if [[ ! -d /sys/bus/pci/devices/" + pciDevId + " ]]; then echo 'Error: PCI address does not exist!' && exit 1; fi",
		"if [[ ! -d /sys/bus/pci/devices/" + pciDevId + "/iommu/ ]]; then echo 'Error: No vIOMMU found in the VM' && exit 1; fi",
		"[[ '" + driver + "' != 'vfio-pci' ]] && echo " + pciDevId + " > " + driverPath + "/unbind && echo 'vfio-pci' > " + driverOverride + " && echo " + pciDevId + " > /sys/bus/pci/drivers/vfio-pci/bind",
	}

	for _, cmd := range cmds {
		if _, err := o.sshClient.SSH(cmd, true); err != nil {
			return err
		}
	}

	return nil
}
