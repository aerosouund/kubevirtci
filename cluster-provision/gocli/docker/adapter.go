package docker

import (
	"bytes"
	"fmt"
	"os"

	"github.com/docker/docker/client"
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

func (d *DockerAdapter) SSH(cmd string, stdOut bool) (string, error) {
	if len(cmd) > 0 {
		firstCmdChar := cmd[0]
		switch string(firstCmdChar) {
		// directly runnable script
		case "/":
			cmd = "ssh.sh sudo /bin/bash < " + cmd
		// script with parameters
		case "-":
			cmd = "ssh.sh sudo /bin/bash " + cmd
		// ordinary command
		default:
			cmd = "ssh.sh " + cmd
		}
	}
	var err error
	var success bool
	var stdout bytes.Buffer

	if stdOut {
		success, err = Exec(d.dockerClient, d.nodeName, []string{"/bin/bash", "-c", cmd}, os.Stdout)
	} else {
		success, err = Exec(d.dockerClient, d.nodeName, []string{"/bin/bash", "-c", cmd}, &stdout)
	}

	if err != nil {
		return "", err
	}

	if !success {
		return "", fmt.Errorf("Error executing %s on node %s", cmd, d.nodeName)
	}
	return stdout.String(), nil
}
