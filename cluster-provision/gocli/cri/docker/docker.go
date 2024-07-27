package docker

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
)

type DockerClient struct{}

func NewDockerClient() *DockerClient {
	return &DockerClient{}
}

func IsAvailable() bool {
	cmd := exec.Command("docker", "-v")
	out, err := cmd.Output()
	if err != nil {
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
		"start",
		containerID)

	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func (dc *DockerClient) Create(image string, createOpts *cri.CreateOpts) (string, error) {
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

	cmd := exec.Command("docker",
		fullArgs...,
	)

	containerID, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	logrus.Info("created registry container with id: ", string(containerID))
	return strings.TrimSuffix(string(containerID), "\n"), nil
}

func (dc *DockerClient) Remove(containerID string) error {
	cmd := exec.Command("docker", "rm", "-f", containerID)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}