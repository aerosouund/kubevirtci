package podman

import (
	"context"

	"github.com/containers/common/libnetwork/types"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
)

type Podman struct {
	// podman adapter
	Conn context.Context
}

func NewPodman() (*Podman, error) {
	conn, err := bindings.NewConnection(context.Background(), "unix:///run/podman/podman.sock")
	if err != nil {
		return nil, err
	}
	return &Podman{
		Conn: conn,
	}, nil
}

func (p *Podman) ImagePull(image string) error {
	if _, err := images.Pull(p.Conn, image, nil); err != nil {
		return err
	}

	return nil
}

func (p *Podman) Create(image string, co *cri.CreateOpts) (string, error) {
	s := specgen.NewSpecGenerator(image, false)
	s = p.createOptsToSpec(s, co)
	createResponse, err := containers.CreateWithSpec(p.Conn, s, &containers.CreateOptions{})
	if err != nil {
		return "", err
	}
	return createResponse.ID, nil
}

func (p *Podman) Run(containerID string) error {
	if err := containers.Start(p.Conn, containerID, &containers.StartOptions{}); err != nil {
		return err
	}
	return nil
}

func (p *Podman) Inspect(containerID string) ([]byte, error) {
	_, err := containers.Inspect(p.Conn, containerID, nil)
	if err != nil {
		return nil, err
	}

	return []byte{}, nil
}

func (p *Podman) Remove(containerID string) error {
	if _, err := containers.Remove(p.Conn, containerID, nil); err != nil {
		return err
	}
	return nil
}

func (p *Podman) createOptsToSpec(s *specgen.SpecGenerator, co *cri.CreateOpts) *specgen.SpecGenerator {
	s.CapAdd = co.Capabilities
	s.Privileged = &co.Privileged
	s.RestartPolicy = co.RestartPolicy
	s.Command = co.Command
	s.Networks = map[string]types.PerNetworkOptions{
		co.Network: {},
	}
	s.Remove = &co.Remove
	return s
}
