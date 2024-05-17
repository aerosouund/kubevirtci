package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

func k8sApply(manifestPath string) error {
	kubeconfig := filepath.Join(
		".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	yamlData, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		log.Fatalf("Error converting YAML to JSON: %v", err)
	}

	obj := &unstructured.Unstructured{}
	dec := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	_, _, err = dec.Decode(jsonData, nil, obj)
	if err != nil {
		log.Fatalf("Error decoding JSON to Unstructured object: %v", err)
	}

	gvk := obj.GroupVersionKind()
	restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gvk.GroupVersion()})
	restMapper.Add(gvk, meta.RESTScopeNamespace)
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		log.Fatalf("Error getting REST mapping: %v", err)
	}

	resourceClient := dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	_, err = resourceClient.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Error applying manifest: %v", err)
	}

	fmt.Println("Manifest applied successfully!")
	return nil
}
