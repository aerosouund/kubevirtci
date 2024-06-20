package bindvfio

import (
	"fmt"
	"strings"
	"time"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type BindVfioOpt struct {
	sshPort uint16
	nodeIdx int
	pciID   string
}

func NewBindVfioOpt(sshPort uint16, nodeIdx int, id string) *BindVfioOpt {
	return &BindVfioOpt{
		sshPort: sshPort,
		nodeIdx: nodeIdx,
		pciID:   id,
	}
}

func (o *BindVfioOpt) Exec() error {
	addr, err := utils.JumpSSH(o.sshPort, o.nodeIdx, "lspci -D -d "+o.pciID, true, false)
	if err != nil {
		return err
	}

	pciDevId := strings.Split(addr, " ")[0]
	fmt.Println("pci addr :", addr, pciDevId)

	devSysfsPath := "/sys/bus/pci/devices/" + pciDevId
	driverPath := devSysfsPath + "/driver"
	driverOverride := devSysfsPath + "/driver_override"

	driver, err := utils.JumpSSH(o.sshPort, o.nodeIdx, "readlink "+driverPath+" | awk -F'/' '{print $NF}'", true, false)
	if err != nil {
		return err
	}
	driver = strings.TrimSuffix(driver, "\n")

	if _, err := utils.JumpSSH(o.sshPort, o.nodeIdx, "modprobe -i vfio-pci", true, false); err != nil {
		return fmt.Errorf("Error loading vfio-pci module: %v", err)
	}

	for i := 0; i < 10; i++ {
		if _, err := utils.JumpSSH(o.sshPort, o.nodeIdx, "ls /sys/bus/pci/drivers/vfio-pci", true, false); err != nil {
			fmt.Println("module not loaded properly, sleeping 1 second and trying again")
			time.Sleep(time.Second * 1)
			utils.JumpSSH(o.sshPort, o.nodeIdx, "modprobe -i vfio-pci", true, false)
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
		if _, err := utils.JumpSSH(o.sshPort, 1, cmd, true, true); err != nil {
			return err
		}
	}

	return nil
}
