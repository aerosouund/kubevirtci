package utils

import (
	"fmt"
	"os/exec"
)

func DetectContainerRuntime() (string, error) {
	podmanCmd := exec.Command("podman", "ps")
	dockerCmd := exec.Command("docker", "ps")
	if err := podmanCmd.Run(); err != nil {
		if err := dockerCmd.Run(); err != nil {
			return "", fmt.Errorf("No valid CRI is running")
		}
		return "docker", nil
	}
	return "podman", nil
}
