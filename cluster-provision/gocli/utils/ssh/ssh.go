package utils

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

type SSHClient interface {
	SSH(cmd string, stdout bool) (string, error)
}

type SSHClientImpl struct {
	sshPort uint16
	nodeIdx int
	config  *ssh.ClientConfig
}

func NewSSHClient(port uint16, idx int, root bool) (*SSHClientImpl, error) {
	signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return nil, err
	}
	u := "vagrant"
	if root {
		u = "root"
	}

	c := &ssh.ClientConfig{
		User: u,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return &SSHClientImpl{
		config:  c,
		sshPort: port,
		nodeIdx: idx,
	}, nil
}

// Jump performs two ssh connections, one to the forwarded port by dnsmasq to the local which is the ssh port of the control plane node
// then a hop to the designated host where the command is desired to be ran
func (s *SSHClientImpl) SSH(cmd string, stdOut bool) (string, error) {
	client, err := ssh.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(s.sshPort)), s.config)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	conn, err := client.Dial("tcp", fmt.Sprintf("192.168.66.10%d:22", s.nodeIdx))
	if err != nil {
		return "", fmt.Errorf("Error establishing connection to the next hop host: %s", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("192.168.66.10%d:22", s.nodeIdx), s.config)
	if err != nil {
		return "", fmt.Errorf("Error creating forwarded ssh connection: %s", err)
	}
	jumpHost := ssh.NewClient(ncc, chans, reqs)
	session, err := jumpHost.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	if !stdOut {
		session.Stdout = &stdout
		session.Stderr = &stderr
	}

	err = session.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("Failed to execute command: %v, %v", err, stderr.String())
	}
	return stdout.String(), nil
}
