package cdi

import (
	_ "embed"
	"regexp"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/operator.yaml
var operator []byte

//go:embed manifests/cr.yaml
var cr []byte

type CdiOpt struct {
	client        k8s.K8sDynamicClient
	sshClient     libssh.Client
	customVersion string
}

func NewCdiOpt(c k8s.K8sDynamicClient, sshClient libssh.Client, cv string) *CdiOpt {
	return &CdiOpt{
		client:        c,
		sshClient:     sshClient,
		customVersion: cv,
	}
}

func (o *CdiOpt) Exec() error {
	if o.customVersion != "" {
		pattern := `v[0-9]+\.[0-9]+\.[0-9]+(.*)?$`
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		operatorNewVersion := regex.ReplaceAllString(string(operator), o.customVersion)
		operator = []byte(operatorNewVersion)
	}

	for _, manifest := range [][]byte{operator, cr} {
		if err := o.client.Apply(manifest); err != nil {
			return err
		}
	}

	if err := o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pod --timeout=180s --all --namespace cdi"); err != nil {
		return err
	}
	return nil
}
