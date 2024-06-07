package rookceph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestCephOpt(t *testing.T) {
	mockK8sClient := kubevirtcimocks.NewMockK8sDynamicClient(gomock.NewController(t))

	// obj := unstructured.Unstructured{
	// 	Object: make(map[string]interface{}),
	// }

	opt := NewCephOpt(mockK8sClient)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/snapshot.storage.k8s.io_volumesnapshots.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/snapshot.storage.k8s.io_volumesnapshotcontents.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/snapshot.storage.k8s.io_volumesnapshotclasses.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/rbac-snapshot-controller.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/setup-snapshot-controller.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/common.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/crds.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/operator.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/cluster-test.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/pool-test.yaml").Return(nil)

	err := opt.Exec()
	assert.NoError(t, err)
}
