package nodeprovisioner

import (
	"embed"
	"fmt"

	"github.com/sirupsen/logrus"
	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

//go:embed conf/*
var f embed.FS

type NodesProvisioner struct {
	sshPort uint16
	nodeIdx int
}

func NewNodesProvisioner(sshPort uint16, nodeIdx int) *NodesProvisioner {
	return &NodesProvisioner{
		sshPort: sshPort,
		nodeIdx: nodeIdx,
	}
}

func (n *NodesProvisioner) Exec() error {
	cgroupv2, err := f.ReadFile("conf/00-cgroupv2.conf")
	if err != nil {
		return err
	}

	cmds := []string{
		"source /var/lib/kubevirtci/shared_vars.sh",
		`timeout=30; interval=5; while ! hostnamectl | grep Transient; do echo "Waiting for dhclient to set the hostname from dnsmasq"; sleep $interval; timeout=$((timeout - interval)); [ $timeout -le 0 ] && exit 1; done`,
		`[ -f /sys/fs/cgroup/cgroup.controllers ] && mkdir -p /etc/crio/crio.conf.d && echo '` + string(cgroupv2) + `' | sudo tee /etc/crio/crio.conf.d/00-cgroupv2.conf > /dev/null && sudo sed -i 's/--cgroup-driver=systemd/--cgroup-driver=cgroupfs/' /etc/sysconfig/kubelet && sudo systemctl stop kubelet && sudo systemctl restart crio`,
		"while [[ $(systemctl status crio | grep -c active) -eq 0 ]]; do sleep 2; done",
		`echo "KUBELET_EXTRA_ARGS=${KUBELET_CGROUP_ARGS} --fail-swap-on=false ${nodeip} --feature-gates=CPUManager=true,NodeSwap=true --cpu-manager-policy=static --kube-reserved=cpu=250m --system-reserved=cpu=250m" | sudo tee /etc/sysconfig/kubelet > /dev/null`, // todo: add the condition
		"sudo systemctl daemon-reload && sudo service kubelet restart",
		"kubelet_rc=$?; [[ $kubelet_rc -ne 0 ]] && rm -rf /var/lib/kubelet/cpu_manager_state && service kubelet restart",
		"sudo swapoff -a",
		"until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done",
		"sudo kubeadm join --token abcdef.1234567890123456 192.168.66.101:6443 --ignore-preflight-errors=all --discovery-token-unsafe-skip-ca-verification=true",
		"sudo mkdir -p /var/lib/rook",
		"sudo chcon -t container_file_t /var/lib/rook",
	}

	for _, cmd := range cmds {
		logrus.Info("executing: ", cmd)
		_, err := utils.JumpSSH(n.sshPort, n.nodeIdx, cmd, true)
		if err != nil {
			return fmt.Errorf("error executing %s: %s", cmd, err)
		}
	}
	return nil
}
