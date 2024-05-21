package rookceph

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/k8s/common"
)

type CephOpt struct {
	client *k8s.K8sDynamicClient
}

func NewCephOpt(c *k8s.K8sDynamicClient) *CephOpt {
	return &CephOpt{
		client: c,
	}
}

func (o *CephOpt) Exec() error {
	manifests := []string{
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshots.yaml",
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshotcontents.yaml",
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshotclasses.yaml",
		"/workdir/manifests/ceph/rbac-snapshot-controller.yaml",
		"/workdir/manifests/ceph/setup-snapshot-controller.yaml",
		"/workdir/manifests/ceph/common.yaml",
		"/workdir/manifests/ceph/crds.yaml",
		"/workdir/manifests/ceph/operator.yaml",
		"/workdir/manifests/ceph/cluster-test.yaml",
		"/workdir/manifests/ceph/pool-test.yaml",
	}

	for _, manifest := range manifests {
		err := o.client.Apply(manifest)
		if err != nil {
			return err
		}
	}
	blockpools, err := o.client.List(schema.GroupVersionKind{
		Group:   "ceph.rook.io",
		Version: "v1",
		Kind:    "cephblockpools"},
		"rook-ceph")
	fmt.Println("blockpools: ", blockpools)
	// for _, bp := range blockpools.Items {
	// 	fmt.Println("found block pool", bp.GetName())
	// }

	if err != nil {
		return err
	}
	return nil
}
