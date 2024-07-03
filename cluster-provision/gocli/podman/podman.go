package podman

import (
	"context"

	"github.com/containers/podman/v5/libpod/define"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
)

type Podman struct {
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

func (p *Podman) Create(image string) (string, error) {
	s := specgen.NewSpecGenerator(image, false)
	createResponse, err := containers.CreateWithSpec(p.Conn, s, nil)
	if err != nil {
		return "", err
	}
	return createResponse.ID, nil
}

func (p *Podman) Run(containerID string) error {
	if err := containers.Start(p.Conn, containerID, nil); err != nil {
		return err
	}
	return nil
}

func (p *Podman) Inspect(containerID string) (*define.InspectContainerData, error) {
	inspectData, err := containers.Inspect(p.Conn, "foobar", new(containers.InspectOptions).WithSize(true))
	if err != nil {
		return nil, err
	}
	// Print the container ID
	return inspectData, nil

}

func (p *Podman) Remove(containerID string) error {
	if _, err := containers.Remove(p.Conn, containerID, nil); err != nil {
		return err
	}
	return nil
}
