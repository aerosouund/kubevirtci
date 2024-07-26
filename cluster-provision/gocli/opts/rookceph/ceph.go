package rookceph

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	cephv1 "github.com/aerosouund/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type CephOpt struct {
	client k8s.K8sDynamicClient
}

func NewCephOpt(c k8s.K8sDynamicClient) *CephOpt {
	return &CephOpt{
		client: c,
	}
}

func (o *CephOpt) Exec() error {
	err := fs.WalkDir(f, "manifests", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".yaml" {
			yamlData, err := f.ReadFile(path)
			if err != nil {
				return err
			}
			if err := o.client.Apply(yamlData); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	blockpool := &cephv1.CephBlockPool{}

	operation := func() error {
		obj, err := o.client.Get(schema.GroupVersionKind{
			Group:   "ceph.rook.io",
			Version: "v1",
			Kind:    "CephBlockPool",
		}, "replicapool", "rook-ceph")

		if err != nil {
			logrus.Errorf("Attempt failed: %v", err)
			return err
		}

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, blockpool)
		if err != nil {
			logrus.Errorf("Attempt failed: %v", err)
			return err
		}

		if blockpool.Status == nil || blockpool.Status.Phase != "Ready" {
			err := fmt.Errorf("Ceph pool block didn't move to ready status")
			logrus.Info(err)
			return err
		}

		return nil
	}

	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.InitialInterval = 30 * time.Second
	backoffStrategy.MaxElapsedTime = 6 * time.Minute

	err = backoff.Retry(operation, backoffStrategy)
	if err != nil {
		return fmt.Errorf("Operation failed after maximum retries: %v", err)
	}

	return nil
}
