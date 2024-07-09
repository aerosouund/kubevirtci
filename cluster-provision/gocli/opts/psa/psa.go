package psa

import (
	"embed"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

//go:embed manifests/*
var f embed.FS

type PsaOpt struct {
	sshPort   uint16
	sshClient utils.SSHClient
}

func NewPsaOpt(sc utils.SSHClient, p uint16) *PsaOpt {
	return &PsaOpt{
		sshPort:   p,
		sshClient: sc,
	}
}

func (o *PsaOpt) Exec() error {
	psa, err := f.ReadFile("manifests/psa.yaml")
	if err != nil {
		return err
	}
	cmds := []string{
		"rm /etc/kubernetes/psa.yaml",
		"echo '" + string(psa) + "' | sudo tee /etc/kubernetes/psa.yaml > /dev/null",
	}
	for _, cmd := range cmds {
		if _, err := o.sshClient.SSH(cmd, true); err != nil {
			return err
		}
	}

	return nil
}
