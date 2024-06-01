package istio

import (
	"embed"
	"fmt"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

//go:embed manifests/*
var f embed.FS

type IstioOpt struct {
	sshPort     uint16
	cnaoEnabled bool
}

func NewIstioOpt(sshPort uint16, cnaoEnabled bool) *IstioOpt {
	return &IstioOpt{
		sshPort:     sshPort,
		cnaoEnabled: cnaoEnabled,
	}
}

func (o *IstioOpt) Exec() error {
	istioCnao, err := f.ReadFile("manifests/istio-operator-with-cnao.yaml")
	if err != nil {
		return err
	}
	istioWithoutCnao, err := f.ReadFile("manifests/istio-operator-with-cnao.yaml")
	if err != nil {
		return err
	}
	cmds := []string{
		"/bin/bash -c /var/lib/kubevirtci/shared_vars.sh",
		"sudo kubectl --kubeconfig /etc/kubernetes/admin.conf create ns istio-system",
		"sudo istioctl --kubeconfig /etc/kubernetes/admin.conf --hub quay.io/kubevirtci operator init",
		fmt.Sprintf("cat <<EOF > /opt/istio/istio-operator-with-cnao.cr.yaml\n%s\nEOF", string(istioCnao)),
		fmt.Sprintf("cat <<EOF > /opt/istio/istio-operator.cr.yaml\n%s\nEOF", string(istioWithoutCnao)),
	}
	for _, cmd := range cmds {
		_, err := utils.JumpSSH(o.sshPort, 1, cmd, true)
		if err != nil {
			return err
		}
	}
	confFile := "/opt/istio/istio-operator.cr.yaml"
	if o.cnaoEnabled {
		confFile = "/opt/istio/istio-operator-with-cnao.cr.yaml"
	}

	_, err = utils.JumpSSH(o.sshPort, 1,
		"sudo kubectl --kubeconfig /etc/kubernetes/admin.conf create -f "+confFile,
		true)
	if err != nil {
		return err
	}

	return nil
}
