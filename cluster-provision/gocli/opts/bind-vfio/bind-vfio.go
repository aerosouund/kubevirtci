package bindvfio

import (
	"fmt"
	"strings"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type BindVfioOpt struct {
	pciID     string
	sshClient libssh.Client
}

func NewBindVfioOpt(sshClient libssh.Client, id string) *BindVfioOpt {
	return &BindVfioOpt{
		pciID:     id,
		sshClient: sshClient,
	}
}

func (o *BindVfioOpt) Exec() error {
	addr, err := o.sshClient.CommandWithNoStdOut("lspci -D -d " + o.pciID)
	if err != nil {
		return err
	}

	pciDevId := strings.Split(addr, " ")[0]

	devSysfsPath := "/sys/bus/pci/devices/" + pciDevId
	driverPath := devSysfsPath + "/driver"
	driverOverride := devSysfsPath + "/driver_override"

	driver, err := o.sshClient.CommandWithNoStdOut("readlink " + driverPath + " | awk -F'/' '{print $NF}'")
	if err != nil {
		return err
	}
	driver = strings.TrimSuffix(driver, "\n")

	if err := o.sshClient.Command("modprobe -i vfio-pci"); err != nil {
		return fmt.Errorf("Error loading vfio-pci module: %v", err)
	}

	for i := 0; i < 10; i++ {
		if err := o.sshClient.Command("ls /sys/bus/pci/drivers/vfio-pci"); err != nil {
			fmt.Println("module not loaded properly, sleeping 1 second and trying again")
			time.Sleep(time.Second * 1)
			o.sshClient.Command("modprobe -i vfio-pci")
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
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
