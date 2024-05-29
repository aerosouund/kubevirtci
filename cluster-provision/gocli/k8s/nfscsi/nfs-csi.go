package nfscsi

import (
	"embed"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/k8s/common"
)

//go:embed manifests/*
var f embed.FS

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
		"nfs-service.yaml",
		"nfs-server.yaml",
		"csi-nfs-controller-rbac.yaml",
		"csi-nfs-driverinfo.yaml",
		"csi-nfs-controller.yaml",
		"csi-nfs-node.yaml",
		"csi-nfs-sc.yaml",
		"csi-nfs-test-pvc.yaml",
	}

	for _, manifest := range manifests {
		err := o.client.Apply(f, manifest)
		if err != nil {
			return err
		}
	}
	pvc := &corev1.PersistentVolumeClaim{}
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		obj, err := o.client.Get(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "PersistentVolumeClaim"},
			"pvc-nfs-dynamic",
			"nfs-csi")
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pvc)
		if err != nil {
			return err
		}

		if pvc.Status.Phase == "Bound" {
			break
		}
		fmt.Println("PVC didn't move to Bound phase, sleeping for 10 seconds")
		time.Sleep(10 * time.Second)
	}

	if pvc.Status.Phase != "Bound" {
		return fmt.Errorf("PVC failed to transition to Bound!")
	}

	err := o.client.Delete(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "PersistentVolumeClaim",
	}, "pvc-nfs-dynamic", "nfs-csi")
	if err != nil {
		return err
	}

	return nil
}
