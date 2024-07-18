package vgpu

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/docker/docker/client"
	dockercri "kubevirt.io/kubevirtci/cluster-provision/gocli/cri/docker"
	podmancri "kubevirt.io/kubevirtci/cluster-provision/gocli/cri/podman"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/remountsysfs"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/kindbase"
)

const kindVGPUImage = "kindest/node:v1.30.0@sha256:047357ac0cfea04663786a612ba1eaba9702bef25227a794b52890dd8bcd692e"

type KindVGPU struct {
	*kind.KindBaseProvider
}

func NewKindVGPU(kindConfig *kind.KindConfig) (*KindVGPU, error) {
	kindBase, err := kind.NewKindBaseProvider(kindConfig)
	if err != nil {
		return nil, err
	}

	kindBase.Image = kindVGPUImage
	cluster, err := kindBase.PrepareClusterYaml(true, true)
	if err != nil {
		return nil, err
	}

	kindBase.Cluster = cluster

	return &KindVGPU{
		KindBaseProvider: kindBase,
	}, nil
}

func (kv *KindVGPU) Start(ctx context.Context, cancel context.CancelFunc) error {
	err := kv.KindBaseProvider.Start(ctx, cancel)
	if err != nil {
		return err
	}

	nodes, err := kv.Provider.ListNodes(kv.Version)
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	var sshClient libssh.Client
	for _, node := range nodes {
		switch kv.CRI.(type) {
		case *dockercri.DockerClient:
			sshClient = docker.NewDockerAdapter(cli, node.String())
		case *podmancri.Podman:
			sshClient = podmancri.NewPodmanSSHClient(node.String())
		}

		rsf := remountsysfs.NewRemountSysFSOpt(sshClient)
		if err := rsf.Exec(); err != nil {
			return err
		}

		// what are we doing with the vgpus discovered ??
		if _, err = kv.discoverHostVGPUs(); err != nil {
			return err
		}
	}

	return nil
}

func (kv *KindVGPU) discoverHostVGPUs() ([]string, error) {
	files, err := filepath.Glob("/sys/class/mdev_bus/*/mdev_supported_types")
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errors.New("FATAL: Could not find available GPUs on host")
	}

	vpgus := make([]string, 0)
	for _, file := range files {
		vgpuName := filepath.Base(filepath.Dir(filepath.Dir(file)))
		vpgus = append(vpgus, vgpuName)
	}

	return vpgus, nil
}
