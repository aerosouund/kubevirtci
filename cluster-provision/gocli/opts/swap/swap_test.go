package swap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestSwapOpt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	o := NewSwapOpt(sshClient, 2020, 1, 10, true, "1G")

	cmds := []string{
		"dd if=/dev/zero of=/swapfile count=" + o.size + " bs=1G",
		"swapon -a",
		"/bin/su -c \"echo vm.swappiness = " + fmt.Sprintf("%d", o.swapiness) + " >> /etc/sysctl.conf\"",
		"sysctl vm.swappiness=" + fmt.Sprintf("%d", o.swapiness),
		`sed -i ':a;N;\$!ba;s/memorySwap: {}/memorySwap:\n  swapBehavior: UnlimitedSwap/g'  /var/lib/kubelet/config.yaml`,
		"systemctl restart kubelet",
	}

	for _, cmd := range cmds {
		sshClient.EXPECT().SSH(cmd, true)
	}
	err := o.Exec()
	assert.NoError(t, err)
}
