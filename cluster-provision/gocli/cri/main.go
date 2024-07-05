package cri

import utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"

// maybe just create wrappers around bash after all
type ContainerClient interface {
	ImagePull(image string) error
	Create(image string, co *CreateOpts) (string, error)
	Start(containerID string) error
	Remove(containerID string) error
	Inspect(containerID string) ([]byte, error)
	utils.SSHClient
}

type CreateOpts struct {
	Privileged    bool
	Name          string
	Ports         map[string]string
	RestartPolicy string
	Network       string
	Command       []string
	Remove        bool
	Capabilities  []string
}
