package nodes

import (
	"embed"
	"fmt"

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
		`[ -f /sys/fs/cgroup/cgroup.controllers ] && mkdir -p /etc/crio/crio.conf.d && echo '` + string(cgroupv2) + `' |  tee /etc/crio/crio.conf.d/00-cgroupv2.conf > /dev/null && source /var/lib/kubevirtci/shared_vars.sh && systemctl stop kubelet && systemctl restart crio`,
		"while [[ $(systemctl status crio | grep -c active) -eq 0 ]]; do sleep 2; done",
		`echo "KUBELET_EXTRA_ARGS=${KUBELET_CGROUP_ARGS} --fail-swap-on=false ${nodeip} --feature-gates=CPUManager=true,NodeSwap=true --cpu-manager-policy=static --kube-reserved=cpu=250m --system-reserved=cpu=250m" |  tee /etc/sysconfig/kubelet > /dev/null`, // todo: add the condition
		" systemctl daemon-reload &&  service kubelet restart",
		// "if [[ $? -ne 0 ]]; then && rm -rf /var/lib/kubelet/cpu_manager_state && service kubelet restart; fi",
		"swapoff -a",
		"until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done",
		"kubeadm join --token abcdef.1234567890123456 192.168.66.101:6443 --ignore-preflight-errors=all --discovery-token-unsafe-skip-ca-verification=true",
		"mkdir -p /var/lib/rook",
		"chcon -t container_file_t /var/lib/rook",
	}

	for _, cmd := range cmds {
		_, err := utils.JumpSSH(n.sshPort, n.nodeIdx, cmd, true, true)
		if err != nil {
			return fmt.Errorf("error executing %s: %s", cmd, err)
		}
	}
	return nil
}
