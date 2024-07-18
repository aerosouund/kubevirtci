package ovn

import (
	"context"
	"os/exec"
)

type KindOvn struct{}

const (
	ovnRepo     = "https://github.com/ovn-org/ovn-kubernetes.git"
	ovnCommit   = "c77ee8c38c6a6d9e55131a1272db5fad5b606e44"
	clusterPath = "/tmp/ovn"
)

func NewKindOvnProvider() *KindOvn {
	return &KindOvn{}
}

func (ko *KindOvn) Start(ctx context.Context, cancel context.CancelFunc) error {
	err := ko.cloneOvnRepo()
	if err != nil {
		return err
	}
	return nil
}

func (ko *KindOvn) cloneOvnRepo() error {
	cmds := []string{
		"rm -rf " + clusterPath + " || true",
		"git clone " + ovnRepo + " " + clusterPath,
		"cd " + clusterPath,
		"git checkout " + ovnCommit,
	}

	for _, cmd := range cmds {
		command := exec.Command("/bin/sh", "-c", cmd)
		_, err := command.CombinedOutput()
		if err != nil {
			return err
		}
	}
	return nil
}
