package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	ssh1 "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

// NewSSHCommand returns command to SSH to the cluster node
func NewSSHCommand() *cobra.Command {

	ssh := &cobra.Command{
		Use:   "ssh",
		Short: "ssh into a node",
		RunE:  ssh,
		Args:  cobra.MinimumNArgs(1),
	}
	return ssh
}

func ssh(cmd *cobra.Command, args []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	node := args[0]

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	// TODO we can do the ssh session with the native golang client
	container := prefix + "-" + node
	ssh_command := append([]string{"ssh.sh"}, args[1:]...)
	file := os.Stdout
	if terminal.IsTerminal(int(file.Fd())) {
		exitCode, err := docker.Terminal(cli, container, ssh_command, file)
		if err != nil {
			return err
		}
		os.Exit(exitCode)
	} else {
		execExitCodeIsZero, err := docker.Exec(cli, container, ssh_command, file)
		if err != nil {
			return err
		}
		exitCode := 0
		if !execExitCodeIsZero {
			exitCode = 1
		}
		os.Exit(exitCode)
	}
	return nil
}

func hostSSH(nodeIdx int, dnsmasqID string, sshPort uint16, cmd string) (string, error) {
	success, err := docker.Exec(cli, dnsmasqID, []string{
		"yum", "install", "socat", "-y",
	}, os.Stdout)
	if err != nil {
		return "", err
	}
	if !success {
		return "", fmt.Errorf("failed to install socat")
	}
	// docker exec socat on the dnsmasq
	go func() {
		success, err = docker.Exec(cli, dnsmasqID, []string{
			"/usr/bin/socat",
			"TCP-LISTEN:2222,fork,reuseaddr",
			fmt.Sprintf("TCP:192.168.66.10%d:22", nodeIdx),
		}, os.Stdout)
		if err != nil {
			// Handle error if needed
			fmt.Printf("Error executing socat command: %v\n", err)
		}
		if !success {
			// Handle failure if needed
			fmt.Printf("Failed to execute socat command\n")
		}
	}()

	defer func() {
		docker.Exec(cli, dnsmasqID, []string{
			"sh", "-c", "pkill socat",
		}, os.NewFile(0, os.DevNull))
	}()

	signer, err := ssh1.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return "", err
	}

	config := &ssh1.ClientConfig{
		User: "vagrant",
		Auth: []ssh1.AuthMethod{
			ssh1.PublicKeys(signer),
		},
		HostKeyCallback: ssh1.InsecureIgnoreHostKey(),
	}

	client, err := ssh1.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(sshPort)), config)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	session.Stderr = &stderr
	session.Stdout = &stdout

	err = session.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("Failed to execute command: %v", err)
	}
	return stdout.String(), nil
}
