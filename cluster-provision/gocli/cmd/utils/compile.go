package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// this will be used to compile go code to a target os then scp it to the vm to be executed
func CompileToTargetOS(location string) error {
	err := os.Mkdir("bin", 0755) // 0755 sets permissions for the directory
	if err != nil {
		return fmt.Errorf("error creating bin directory")
	}
	os.Chdir("scripts/" + location)
	cmd := exec.Command(
		"go", "build", "-o", fmt.Sprintf("../../bin/%s", location), ".",
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("error:", err)
		return fmt.Errorf("Error executing build: %s", stderr.String())
	}
	os.Chdir("/workdir" + location)

	return nil
}
