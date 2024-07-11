package swap

import (
	"fmt"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type SwapOpt struct {
	sshClient     utils.SSHClient
	swapiness     int
	unlimitedSwap bool
	size          int
}

func NewSwapOpt(sc utils.SSHClient, swapiness int, us bool, size int) *SwapOpt {
	return &SwapOpt{
		sshClient:     sc,
		swapiness:     swapiness,
		unlimitedSwap: us,
		size:          size,
	}
}

func (o *SwapOpt) Exec() error {
	if o.size != 0 {
		if _, err := o.sshClient.SSH("dd if=/dev/zero of=/swapfile count="+fmt.Sprintf("%d", o.size)+" bs=1G", true); err != nil {
			return err
		}
		if _, err := o.sshClient.SSH("mkswap /swapfile", true); err != nil {
			return err
		}
	}
	if _, err := o.sshClient.SSH("swapon -a", true); err != nil {
		return err
	}

	if o.swapiness != 0 {
		cmds := []string{
			"/bin/su -c \"echo vm.swappiness = " + fmt.Sprintf("%d", o.swapiness) + " >> /etc/sysctl.conf\"",
			"sysctl vm.swappiness=" + fmt.Sprintf("%d", o.swapiness),
		}
		for _, cmd := range cmds {
			if _, err := o.sshClient.SSH(cmd, true); err != nil {
				return err
			}
		}
	}

	if o.unlimitedSwap {
		cmds := []string{
			`sed -i 's/memorySwap: {}/memorySwap:\n  swapBehavior: UnlimitedSwap/g' /var/lib/kubelet/config.yaml`,
			"systemctl restart kubelet",
		}
		for _, cmd := range cmds {
			if _, err := o.sshClient.SSH(cmd, true); err != nil {
				return err
			}
		}
	}

	return nil
}
