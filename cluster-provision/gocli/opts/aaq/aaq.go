package aaq

import (
	_ "embed"
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/operator.yaml
var operator []byte

//go:embed manifests/cr.yaml
var cr []byte

type AaqOpt struct {
	client        k8s.K8sDynamicClient
	sshClient     libssh.Client
	customVersion string
}

func NewAaqOpt(c k8s.K8sDynamicClient, sshClient libssh.Client, customVersion string) *AaqOpt {
	return &AaqOpt{
		client:        c,
		sshClient:     sshClient,
		customVersion: customVersion,
	}
}

func (o *AaqOpt) Exec() error {
	if o.customVersion != "" {
		pattern := `v[0-9]+\.[0-9]+\.[0-9]+(.*)?$`
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		operatorNewVersion := regex.ReplaceAllString(string(operator), o.customVersion)
		operator = []byte(operatorNewVersion)
	}

	for i, manifest := range [][]byte{operator, cr} {
		if err := o.client.Apply(manifest); err != nil {
			return fmt.Errorf("error applying manifest at index %d, %s", i, err)
		}
	}

	if err := o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pod --timeout=180s --all --namespace aaq"); err != nil {
		return err
	}
	logrus.Info("AAQ Operator is ready!")
	return nil
}
