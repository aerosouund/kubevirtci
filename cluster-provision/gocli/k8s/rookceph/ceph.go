package rookceph

import (
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
	}

	for _, manifest := range manifests {
		err := o.client.Apply(manifest)
		if err != nil {
			return err
		}
	}
	return nil
}
