package rookceph

import (
	"kubevirt.io/kubevirtci/cluster-provision/gocli/k8s/common"
)

type CephOpt struct {
	client *common.K8sDynamicClient
}

func NewCephOpt(c *common.K8sDynamicClient) *CephOpt {
	return &CephOpt{
		client: c,
	}
}

func (co *CephOpt) Exec() error {
	manifests := []string{
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshots.yaml",
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshotcontents.yaml",
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshotclasses.yaml",
		"/workdir/manifests/ceph/rbac-snapshot-controller.yaml",
		"/workdir/manifests/ceph/setup-snapshot-controller.yaml",
	}

	for _, manifest := range manifests {
		err := co.client.Apply(manifest)
		if err != nil {
			return err
		}
	}
	return nil
}
