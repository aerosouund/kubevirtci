package k8sprovision

import (
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed conf/*
var f embed.FS

//go:embed patches/*
var patchFs embed.FS

type K8sProvisioner struct {
	version   string
	slim      bool
	sshClient libssh.Client
}

func NewK8sProvisioner(sshClient libssh.Client, version string, slim bool) *K8sProvisioner {
	return &K8sProvisioner{
		version:   version,
		slim:      slim,
		sshClient: sshClient,
	}
}

func (k *K8sProvisioner) Exec() error {
	registries, err := f.ReadFile("conf/registries.conf")
	if err != nil {
		return err
	}

	storage, err := f.ReadFile("conf/storage.conf")
	if err != nil {
		return err
	}

	cniPatch, err := f.ReadFile("conf/cni.diff")
	if err != nil {
		return err
	}

	cniV6Patch, err := f.ReadFile("conf/cni_ipv6.diff")
	if err != nil {
		return err
	}

	k8sConf, err := f.ReadFile("conf/k8s.conf")
	if err != nil {
		return err
	}

	calico, err := f.ReadFile("conf/001-calico.conf")
	if err != nil {
		return err
	}

	dhclient, err := f.ReadFile("conf/002-dhclient.conf")
	if err != nil {
		return err
	}

	secContextPatch, err := patchFs.ReadFile("patches/add-security-context-deployment-patch.yaml")
	if err != nil {
		return err
	}

	advAudit, err := f.ReadFile("conf/adv-audit.yaml")
	if err != nil {
		return err
	}

	psa, err := f.ReadFile("conf/psa.yaml")
	if err != nil {
		return err
	}

	kubeAdm, err := f.ReadFile("conf/kubeadm.conf")
	if err != nil {
		return err
	}

	kubeAdm6, err := f.ReadFile("conf/kubeadm_ipv6.conf")
	if err != nil {
		return err
	}

	kubeAdmConf := strings.Replace(string(kubeAdm), "VERSION", k.version, 1)
	kubeAdm6Conf := strings.Replace(string(kubeAdm6), "VERSION", k.version, 1)

	cmds := []string{
		"echo '" + string(registries) + "' | tee /etc/containers/registries.conf >> /dev/null",
		"echo '" + string(storage) + "' | tee /etc/containers/storage.conf >> /dev/null",
		"systemctl enable --now crio",
		"kubeadm config images pull --kubernetes-version " + k.version,
		`image_regex='([a-z0-9\_\.]+[/-]?)+(@sha256)?:[a-z0-9\_\.\-]+' image_regex_w_double_quotes='"?'"${image_regex}"'"?' find /tmp -type f -name '*.yaml' -print0 | xargs -0 grep -iE '(image|value): '"${image_regex_w_double_quotes}" > /tmp/images`,
	}

	for _, cmd := range cmds {
		if _, err := k.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}

	images, err := k.sshClient.Command(`image_regex='([a-z0-9\_\.]+[/-]?)+(@sha256)?:[a-z0-9\_\.\-]+' && image_regex_w_double_quotes='"?'"${image_regex}"'"?' && grep -ioE "${image_regex_w_double_quotes}" /tmp/images`, false)
	if err != nil {
		return err
	}

	if !k.slim {
		imagesList := strings.Split(images, "\n")
		for _, image := range imagesList {
			err := k.pullImageRetry(image)
			if err != nil {
				logrus.Infof("Failed to pull image: %s, it will not be available offline", image)
			}
		}

		extraImg, err := f.ReadFile("conf/extra-images")
		if err != nil {
			return err
		}

		imagesList = strings.Split(string(extraImg), "\n")
		for _, image := range imagesList {
			err := k.pullImageRetry(image)
			if err != nil {
				logrus.Infof("Failed to pull image: %s, it will not be available offline", image)
			}
		}
	}

	cmds = []string{
		"mkdir -p /etc/provision",
		"cp /tmp/cni.do-not-change.yaml /etc/provision/cni.yaml",
		"mv /tmp/cni.do-not-change.yaml /etc/provision/cni_ipv6.yaml",
		"echo '" + string(cniPatch) + "' | tee /tmp/cni_patch.diff >> /dev/null",
		"echo '" + string(cniV6Patch) + "' | tee /tmp/cni_v6_patch.diff >> /dev/null",
		"patch /etc/provision/cni.yaml /tmp/cni_patch.diff",
		"patch /etc/provision/cni_ipv6.yaml /tmp/cni_v6_patch.diff",
		"cp /tmp/local-volume.yaml /etc/provision/local-volume.yaml",
		`echo "vm.unprivileged_userfaultfd = 1" > /etc/sysctl.d/enable-userfaultfd.conf`,
		"modprobe bridge",
		"modprobe overlay",
		"modprobe br_netfilter",
		"echo '" + string(k8sConf) + "' | tee /etc/sysctl.d/k8s.conf >> /dev/null",
		"sysctl --system",
		"echo bridge >> /etc/modules-load.d/k8s.conf",
		"echo br_netfilter >> /etc/modules-load.d/k8s.conf",
		"echo overlay >> /etc/modules-load.d/k8s.conf",
		"rm -f /etc/cni/net.d/*",
		"systemctl daemon-reload",
		"systemctl enable crio kubelet --now",
		"echo '" + string(calico) + "' | tee /etc/NetworkManager/conf.d/001-calico.conf >> /dev/null",
		"echo '" + string(dhclient) + "' | tee /etc/NetworkManager/conf.d/002-dhclient.conf >> /dev/null",
		`echo "net.netfilter.nf_conntrack_max=1000000" >> /etc/sysctl.conf`,
		"sysctl --system",
		"systemctl restart NetworkManager",
		"echo bridge >> /etc/modules-load.d/k8s.conf",
		"echo br_netfilter >> /etc/modules-load.d/k8s.conf",
		"echo overlay >> /etc/modules-load.d/k8s.conf",
		"mkdir -p /etc/provision/kubeadm-patches",
		"mkdir /etc/kubernetes/audit",
		"echo '" + string(advAudit) + "' | tee /etc/kubernetes/audit/adv-audit.yaml >> /dev/null",
		"echo '" + string(psa) + "' | tee /etc/kubernetes/psa.yaml >> /dev/null",
		"echo '" + kubeAdmConf + "' | tee /etc/kubernetes/kubeadm.conf >> /dev/null",
		"echo '" + kubeAdm6Conf + "' | tee /etc/kubernetes/kubeadm_ipv6.conf >> /dev/null",
		"swapoff -a",
		"systemctl restart kubelet",
		"kubeadm init --config /etc/kubernetes/kubeadm.conf -v5 || true",
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p '" + string(secContextPatch) + "'",
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /etc/provision/cni.yaml",
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pods --all -n kube-system --timeout=300s",
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system",
		"kubeadm reset --force",
		// "mkdir -p /var/provision/kubevirt.io/tests",
		// "chcon -t container_file_t /var/provision/kubevirt.io/tests",
		// `echo "tmpfs /var/provision/kubevirt.io/tests tmpfs rw,context=system_u:object_r:container_file_t:s0 0 1" >> /etc/fstab`,
		// "rm -f /etc/sysconfig/network-scripts/ifcfg-*",
		// "nmcli connection add con-name eth0 ifname eth0 type ethernet",
		// "rm -f /etc/machine-id ; touch /etc/machine-id",
	}

	for _, cmd := range cmds {
		if _, err := k.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}

	return nil
}

func (k *K8sProvisioner) pullImageRetry(image string) error {
	maxRetries := 5
	downloaded := false

	for i := 0; i < maxRetries; i++ {
		if _, err := k.sshClient.Command("crictl pull "+image, true); err != nil {
			logrus.Infof("Attempt [%d]: Failed to download image %s: %s, sleeping 3 seconds and trying again", i+1, image, err.Error())
			time.Sleep(time.Second * 3)
		} else {
			downloaded = true
			break
		}
	}

	if !downloaded {
		return fmt.Errorf("reached max retries to download for %s", image)
	}
	return nil
}
