package etcdinmemory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRealTimeOpt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewEtcdInMemOpt(sshClient, 2020, 1, "512M")

	sshClient.EXPECT().JumpSSH(opt.sshPort, 1, "mkdir -p /var/lib/etcd", true, true)
	sshClient.EXPECT().JumpSSH(opt.sshPort, 1, "test -d /var/lib/etcd", true, true)
	sshClient.EXPECT().JumpSSH(opt.sshPort, 1, fmt.Sprintf("mount -t tmpfs -o size=%s tmpfs /var/lib/etcd", opt.etcdSize), true, true)
	err := opt.Exec()
	assert.NoError(t, err)
}
