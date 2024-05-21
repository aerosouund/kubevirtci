package nfscsi

import k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/k8s/common"

type NfsCsiOpt struct {
	client *k8s.K8sDynamicClient
}

func NewNfsCsiOpt(c *k8s.K8sDynamicClient) *NfsCsiOpt {
	return &NfsCsiOpt{
		client: c,
	}
}

func (o *NfsCsiOpt) Exec() error {
	manifests := []string{
		"/workdir/manifests/nfs-csi/nfs-service.yaml",
		"/workdir/manifests/nfs-csi/nfs-server.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-controller-rbac.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-driverinfo.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-controller.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-node.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-sc.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-test-pvc.yaml",
	}

	for _, manifest := range manifests {
		err := o.client.Apply(manifest)
		if err != nil {
			return err
		}
	}

	return nil
}
