package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// this will be used to compile go code to a target os then scp it to the vm to be executed
func CompileToTargetOS(location string) error {
	time.Sleep(time.Second * 5000)
	cmd := exec.Command(
		"go", "build", "-o", fmt.Sprintf("./bin/%s", location), fmt.Sprintf("./scripts/%s", location),
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("error:", err)
		return fmt.Errorf("Error executing build: %s", stderr.String())
	}

	return nil
}
