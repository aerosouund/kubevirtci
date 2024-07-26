package psa

import (
	_ "embed"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/psa.yaml
var psa []byte

type PsaOpt struct {
	sshClient libssh.Client
}

func NewPsaOpt(sc libssh.Client) *PsaOpt {
	return &PsaOpt{
		sshClient: sc,
	}
}

func (o *PsaOpt) Exec() error {
	cmds := []string{
		"rm /etc/kubernetes/psa.yaml",
		"echo '" + string(psa) + "' | sudo tee /etc/kubernetes/psa.yaml > /dev/null",
	}
	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
