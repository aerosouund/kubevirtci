package kindcommon

import (
	"context"

	kindmain "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

type KindCommonProvider struct {
	Version         string
	Nodes           int
	Client          k8s.K8sDynamicClient `json:"-"`
	ContainerClient kindmain.ContainerClient
}

func (k *KindCommonProvider) Start(ctx context.Context, cancel context.CancelFunc) {
	// download the kind cli or a library or whatever -- _fetch_kind
	kind := kind.Provider{}

}
