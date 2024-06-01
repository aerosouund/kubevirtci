package node01

import (
	"fmt"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type Node01Provisioner struct {
	sshPort uint16
}

func NewNode01Provisioner(sshPort uint16) *Node01Provisioner {
	return &Node01Provisioner{
		sshPort: sshPort,
	}
}

func (n *Node01Provisioner) Exec() error {
	cmds := []string{
		`[ -f /home/vagrant/single_stack ] && kubeadm_conf="/etc/kubernetes/kubeadm_ipv6.conf" && cni_manifest="/provision/cni_ipv6.yaml" || { kubeadm_conf="/etc/kubernetes/kubeadm.conf"; cni_manifest="/provision/cni.yaml"; }`,
		// `[ -f /home/vagrant/enable_audit ] && apiVer=$(head -1 /etc/kubernetes/audit/adv-audit.yaml) && echo "$apiVer" > /etc/kubernetes/audit/adv-audit.yaml && echo -e "kind: Policy\nrules:\n- level: Metadata" >> /etc/kubernetes/audit/adv-audit.yaml`,
		`timeout=30; interval=5; while ! hostnamectl | grep Transient; do echo "Waiting for dhclient to set the hostname from dnsmasq"; sleep $interval; timeout=$((timeout - interval)); [ $timeout -le 0 ] && exit 1; done`,
		// `[ -f /sys/fs/cgroup/cgroup.controllers ] && echo "Configuring cgroup v2" && CRIO_CONF_DIR=/etc/crio/crio.conf.d && mkdir -p ${CRIO_CONF_DIR} && sudo echo '[crio.runtime]\nconmon_cgroup = "pod"\ncgroup_manager = "cgroupfs"' > ${CRIO_CONF_DIR}/00-cgroupv2.conf && sudo sed -i 's/--cgroup-driver=systemd/--cgroup-driver=cgroupfs/' /etc/sysconfig/kubelet && sudo systemctl stop kubelet && sudo systemctl restart crio && sudo systemctl start kubelet`,
		"while [[ $(systemctl status crio | grep -c active) -eq 0 ]]; do sleep 2; done",
		"sudo swapoff -a",
		"until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done",
		`sudo kubeadm init --config "$kubeadm_conf" -v5`,
		`sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p "$(cat /provision/kubeadm-patches/add-security-context-deployment-patch.yaml)"`, // todo: dont make this depend on the node container
		`sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$cni_manifest"`,
		`sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/control-plane:NoSchedule-`,
		`sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers; kubectl_rc=$?; retry_counter=0; while [[ $retry_counter -lt 20 && $kubectl_rc -ne 0 ]]; do sleep 10; echo "Waiting for api server to be available..."; kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers; kubectl_rc=$?; retry_counter=$((retry_counter + 1)); done`,
		"sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf version",
		`local_volume_manifest="/provision/local-volume.yaml"; kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$local_volume_manifest"`,
		"mkdir -p /var/lib/rook",
		"chcon -t container_file_t /var/lib/rook",
	}
	for _, cmd := range cmds {
		_, err := utils.JumpSSH(n.sshPort, 1, cmd, true)
		if err != nil {
			return fmt.Errorf("error executing %s: %s", cmd, err)
		}
	}
	return nil
}
