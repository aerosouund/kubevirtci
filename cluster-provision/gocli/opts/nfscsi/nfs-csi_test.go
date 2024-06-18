package nfscsi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func TestNfsCsiOpt(t *testing.T) {
	client := k8s.NewTestClient([]runtime.Object{})
	opt := NewNfsCsiOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
