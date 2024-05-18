package utils

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func K8sApply(config *rest.Config, manifestPath string) error {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Error creating dynamic client: %v", err)
	}

	yamlData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("Error reading YAML file: %v", err)
	}

	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		return fmt.Errorf("Error converting YAML to JSON: %v", err)
	}

	obj := &unstructured.Unstructured{}
	dec := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	_, _, err = dec.Decode(jsonData, nil, obj)
	if err != nil {
		return fmt.Errorf("Error decoding JSON to Unstructured object: %v", err)
	}

	gvk := obj.GroupVersionKind()
	restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gvk.GroupVersion()})
	restMapper.Add(gvk, meta.RESTScopeNamespace)
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("Error getting REST mapping: %v", err)
	}

	resourceClient := dynamicClient.Resource(mapping.Resource).Namespace("default")
	_, err = resourceClient.Create(context.TODO(), obj, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Error applying manifest: %v", err)
	}

	fmt.Printf("Manifest %v applied successfully!\n", manifestPath)
	return nil
}
