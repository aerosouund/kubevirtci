package dockerproxy

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestNodeProvisionerOpt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewDockerProxyOpt(sshClient, "test-proxy")
	override, err := f.ReadFile("conf/override.conf")
	script := strings.ReplaceAll(string(override), "$PROXY", opt.proxy)

	cmds := []string{
		"curl " + opt.proxy + "/ca.crt > /etc/pki/ca-trust/source/anchors/docker_registry_proxy.crt",
		"update-ca-trust",
		"mkdir -p /etc/systemd/system/crio.service.d",
		"echo '" + script + "' | sudo tee /etc/systemd/system/crio.service.d/override.conf > /dev/null",
		"systemctl daemon-reload",
		"systemctl restart crio.service",
	}
	for _, cmd := range cmds {
		sshClient.EXPECT().SSH(cmd, true)
	}

	err = opt.Exec()
	assert.NoError(t, err)
}
