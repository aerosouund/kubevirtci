package podman

import (
	"context"

	"github.com/containers/podman/v5/pkg/bindings"
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
