package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

// this will be used to compile go code to a target os then scp it to the vm to be executed
func compileToTargetOS(location string) error {
	cmd := exec.Command("CGO_ENABLED=0",
		"GOOS=linux", "GOARCH=amd64",
		"go", "build", "-o", fmt.Sprintf("./bin/%s", location), fmt.Sprintf("./bin/%s", location),
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error executing build: %s", stderr.String())
	}

	return nil
}
