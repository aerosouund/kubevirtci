package provision

import (
	"embed"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

//go:embed conf/*
var f embed.FS

type LinuxProvisioner struct {
	sshPort uint16
}

func NewLinuxProvisioner() *LinuxProvisioner {
	return &LinuxProvisioner{}
}

func (l *LinuxProvisioner) Exec() error {
	sharedVars, err := f.ReadFile("conf/shared_vars.sh")
	if err != nil {
		return nil
	}

	cmds := []string{
		`echo '` + string(sharedVars) + `' | sudo tee /var/lib/kubevirtci/shared_vars.sh > /dev/null`,
		`sudo dnf install -y "kernel-modules-$(uname -r)"`,
		"sudo dnf install -y cloud-utils-growpart",
		`if growpart /dev/vda 1; then sudo resize2fs /dev/vda1; fi`,
		"sudo dnf install -y patch",
		"sudo systemctl stop firewalld || :",
		"systemctl disable firewalld || :",
		"dnf -y remove firewalld",
		"dnf -y install iscsi-initiator-utils",
		"dnf -y install nftables",
		"dnf -y install lvm2",
		`echo 'ACTION=="add|change", SUBSYSTEM=="block", KERNEL=="vd[a-z]", ATTR{queue/rotational}="0"' > /etc/udev/rules.d/60-force-ssd-rotational.rules`,
		"sudo dnf install -y iproute-tc",
		"mkdir -p /opt/istio-1.15.0/bin",
		`curl "https://storage.googleapis.com/kubevirtci-istioctl-mirror/istio-1.15.0/bin/istioctl" -o "/opt/istio-1.15.0/bin/istioctl"`,
		`chmod +x /opt/istio-1.15.0/bin/istioctl`,
		"sudo dnf install -y container-selinux",
		"sudo dnf install -y libseccomp-devel",
		"sudo dnf install -y centos-release-nfv-openvswitch",
		"sudo dnf install -y openvswitch2.16",
		"sudo dnf install -y NetworkManager NetworkManager-ovs NetworkManager-config-server",
	}
	for _, cmd := range cmds {
		if _, err := utils.JumpSSH(l.sshPort, 1, cmd, true); err != nil {
			return err
		}
	}
	return nil
}
