package multus

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

//go:embed manifests/*
var f embed.FS

type MultusOpt struct {
	client    k8s.K8sDynamicClient
	sshClient utils.SSHClient
}

func NewMultusOpt(c k8s.K8sDynamicClient, sshClient utils.SSHClient) *MultusOpt {
	return &MultusOpt{
		client:    c,
		sshClient: sshClient,
	}
}

func (o *MultusOpt) Exec() error {
	yamlData, err := f.ReadFile("manifests/multus.yaml")
	if err != nil {
		return err
	}
	if err := o.client.Apply(yamlData); err != nil {
		return err
	}

	if _, err = o.sshClient.SSH("kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system ds/kube-multus-ds --timeout=200s", true); err != nil {
		return err
	}
	return nil
}
