package cdi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestCdiOpt(t *testing.T) {
	client := k8s.NewTestClient()
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))

	opt := NewCdiOpt(client, sshClient, "")
	sshClient.EXPECT().SSH("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pod --timeout=180s --all --namespace cdi", true)

	err := opt.Exec()
	assert.NoError(t, err)
}
