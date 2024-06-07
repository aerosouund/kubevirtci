package rookceph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func TestCephOpt(t *testing.T) {
	client := k8s.NewTestClient()
	opt := NewCephOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
