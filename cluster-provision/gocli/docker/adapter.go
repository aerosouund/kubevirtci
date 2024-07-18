package docker

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

// DockerAdapter is a wrapper around client.Client to conform it to the SSH interface
type DockerAdapter struct {
	nodeName     string
	dockerClient *client.Client
}

func NewDockerAdapter(cli *client.Client, nodeName string) *DockerAdapter {
	return &DockerAdapter{
		nodeName:     nodeName,
		dockerClient: cli,
	}
}

func (d *DockerAdapter) Command(cmd string, stdOut bool) (string, error) {
	logrus.Infof("[node %s]: %s\n", d.nodeName, cmd)
	success, err := Exec(d.dockerClient, d.nodeName, []string{"/bin/sh", "-c", cmd}, os.Stdout)
	if err != nil {
		return "", err
	}

	if !success {
		return "", fmt.Errorf("Error executing %s on node %s", cmd, d.nodeName)
	}
	return "", nil
}

// maybe add ctx ?
func (d *DockerAdapter) SCP(destPath string, contents fs.File) error {
	return d.dockerClient.CopyToContainer(context.Background(), d.nodeName, destPath, contents, types.CopyToContainerOptions{})
}

// todo
func (d *DockerAdapter) CopyRemoteFile(remotePath, localPath string) error {
	return nil
}
