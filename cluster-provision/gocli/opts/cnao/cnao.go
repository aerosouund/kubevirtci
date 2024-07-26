package cnao

import (
	"embed"
	"io/fs"
	"path/filepath"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/*
var f embed.FS

type CnaoOpt struct {
	client    k8s.K8sDynamicClient
	sshClient libssh.Client
}

func NewCnaoOpt(c k8s.K8sDynamicClient, sshClient libssh.Client) *CnaoOpt {
	return &CnaoOpt{
		client:    c,
		sshClient: sshClient,
	}
}

func (o *CnaoOpt) Exec() error {
	err := fs.WalkDir(f, "manifests", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".yaml" {
			yamlData, err := f.ReadFile(path)
			if err != nil {
				return err
			}
			if err := o.client.Apply(yamlData); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	if err := o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s"); err != nil {
		return err
	}
	return nil
}
