package podman

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/containers/common/libnetwork/types"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
)

type Podman struct {
	// podman adapter
	Conn context.Context
}

func NewPodman() (*Podman, error) {
	// conn, err := bindings.NewConnection(context.Background(), "unix:///run/podman/podman.sock")
	// if err != nil {
	// 	return nil, err
	// }
	return &Podman{}, nil
}

func (p *Podman) ImagePull(image string) error {
	if _, err := images.Pull(p.Conn, image, nil); err != nil {
		return err
	}

	return nil
}

func (p *Podman) Create(image string, createOpts *cri.CreateOpts) (string, error) {
	ports := ""
	for containerPort, hostPort := range createOpts.Ports {
		ports += "-p " + containerPort + ":" + hostPort
	}

	args := []string{
		"--name=" + createOpts.Name,
		"--privileged=" + strconv.FormatBool(createOpts.Privileged),
		"--rm=" + strconv.FormatBool(createOpts.Remove),
		"--restart=" + createOpts.RestartPolicy,
		"--network=" + createOpts.Network,
	}

	for containerPort, hostPort := range createOpts.Ports {
		args = append(args, "-p", containerPort+":"+hostPort)
	}

	if len(createOpts.Capabilities) > 0 {
		args = append(args, "--cap-add="+strings.Join(createOpts.Capabilities, ","))
	}

	fullArgs := append([]string{"create"}, args...)
	fullArgs = append(fullArgs, image)
	fullArgs = append(fullArgs, createOpts.Command...)

	cmd := exec.Command("podman",
		fullArgs...,
	)
	fmt.Println(cmd.String())

	containerID, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	logrus.Info("created registry container with id: ", string(containerID))
	return strings.TrimSuffix(string(containerID), "\n"), nil
}

func (p *Podman) Start(containerID string) error {
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
