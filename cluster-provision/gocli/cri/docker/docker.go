package docker

import (
	"os/exec"
	"strconv"
	"strings"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
)

type DockerClient struct{}

func NewDockerClient() *DockerClient {
	return &DockerClient{}
}

func IsAvailable() bool {
	cmd := exec.Command("docker", "-v")
	out, err := cmd.Output()
	if err != nil || len(out) != 1 {
		return false
	}
	return strings.HasPrefix(string(out), "Docker version")
}

func (dc *DockerClient) ImagePull(image string) error {
	cmd := exec.Command("docker", "pull", image)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (dc *DockerClient) Inspect(containerID string) ([]byte, error) {
	cmd := exec.Command("docker", "inspect", containerID)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (dc *DockerClient) Start(containerID string) error {
	cmd := exec.Command("docker",
		"run",
		containerID)

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (dc *DockerClient) Create(image string, createOpts *cri.CreateOpts) (string, error) {
	ports := ""
	for containerPort, hostPort := range createOpts.Ports {
		ports += "-p " + containerPort + ":" + hostPort
	}
	cmd := exec.Command("docker",
		"create",
		image,
		"--name="+createOpts.Name,
		"--priviliged="+strconv.FormatBool(createOpts.Privileged),
		"--rm="+strconv.FormatBool(createOpts.Remove),
		ports,
		"--restart="+createOpts.RestartPolicy,
		"--network="+createOpts.Network,
		"--cap-add="+strings.Join(createOpts.Capabilities, ","),
		strings.Join(createOpts.Command, " "),
	)

	containerID, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(containerID), nil
}

func (dc *DockerClient) Remove(containerID string) error {
	cmd := exec.Command("docker", "rm", "-f", containerID)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
