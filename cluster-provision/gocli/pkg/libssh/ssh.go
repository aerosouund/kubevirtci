package libssh

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"net"
	"os"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

//go:embed key.pem
var sshKey []byte

// Represents an interface to run a command on a node in the kubevirt cluster, the interface assumes only the bare command or script
// is going to be passed. any leading ways to configure the script like /bin/bash or anything is left to the caller to account for as an implementation detail
type Client interface {
	Command(cmd string) error
	CommandWithNoStdOut(cmd string) (string, error)
	CopyRemoteFile(remotePath, localPath string) error
	SCP(destPath string, contents fs.File) error
}

// Represents an interface to run a command on a node in the kubevirt cluster
// Implementation to the SSHClient interface based on native golang libraries
type SSHClientImpl struct {
	sshPort uint16
	nodeIdx int
	config  *ssh.ClientConfig
	client  *ssh.Client
}

func NewSSHClient(port uint16, idx int, root bool) (*SSHClientImpl, error) {
	signer, err := ssh.ParsePrivateKey(sshKey)
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

// SSH performs two ssh connections, one to the forwarded port by dnsmasq to the local which is the ssh port of the control plane node
// then a hop to the designated host where the command is desired to be ran
func (s *SSHClientImpl) Command(cmd string) error {
	if s.client == nil {
		err := s.initClient()
		if err != nil {
			return err
		}
	}

	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if len(cmd) > 0 {
		firstCmdChar := cmd[0]
		// indicates the command is a script or a script with params
		if string(firstCmdChar) == "/" || string(firstCmdChar) == "-" {
			cmd = "sudo /bin/bash " + cmd
		}
	}
	logrus.Infof("[node %d]: %s", s.nodeIdx, cmd)

	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("Failed to execute command: %v, %v", cmd, err)
	}
	return nil
}

func (s *SSHClientImpl) CommandWithNoStdOut(cmd string) (string, error) {
	if s.client == nil {
		err := s.initClient()
		if err != nil {
			return "", err
		}
	}
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if len(cmd) > 0 {
		firstCmdChar := cmd[0]
		// indicates the command is a script or a script with params
		if string(firstCmdChar) == "/" || string(firstCmdChar) == "-" {
			cmd = "sudo /bin/bash " + cmd
		}
	}
	logrus.Infof("[node %d]: %s", s.nodeIdx, cmd)

	err = session.Run(cmd)
	if err != nil {
		err = fmt.Errorf(stderr.String())
		return "", fmt.Errorf("Failed to execute command: %v, %v", cmd, err)
	}
	return stdout.String(), nil
}

func (s *SSHClientImpl) SCP(fileName string, contents fs.File) error {
	if s.client == nil {
		err := s.initClient()
		if err != nil {
			return err
		}
	}

	scpClient, err := scp.NewClientBySSH(s.client)
	if err != nil {
		return err
	}

	err = scpClient.Connect()
	if err != nil {
		return err
	}

	err = scpClient.CopyFile(context.Background(), contents, fileName, "0775")
	if err != nil {
		return err
	}

	return nil
}

func (s *SSHClientImpl) CopyRemoteFile(remotePath, localPath string) error {
	if s.client == nil {
		err := s.initClient()
		if err != nil {
			return err
		}
	}

	scpClient, err := scp.NewClientBySSH(s.client)
	if err != nil {
		return err
	}

	err = scpClient.Connect()
	if err != nil {
		return err
	}

	destFile, err := os.Open(localPath)
	if err != nil {
		return err
	}

	err = scpClient.CopyFromRemote(context.Background(), destFile, remotePath)
	if err != nil {
		return err
	}

	return nil
}

func (s *SSHClientImpl) initClient() error {
	client, err := ssh.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(s.sshPort)), s.config)
	if err != nil {
		return fmt.Errorf("Failed to connect to SSH server: %v", err)
	}

	conn, err := client.Dial("tcp", fmt.Sprintf("192.168.66.10%d:22", s.nodeIdx))
	if err != nil {
		return fmt.Errorf("Error establishing connection to the next hop host: %s", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("192.168.66.10%d:22", s.nodeIdx), s.config)
	if err != nil {
		return fmt.Errorf("Error creating forwarded ssh connection: %s", err)
	}

	jumpHost := ssh.NewClient(ncc, chans, reqs)
	s.client = jumpHost
	return nil
}
