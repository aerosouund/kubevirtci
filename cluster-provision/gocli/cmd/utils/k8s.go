package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

	yamlDocs := bytes.Split(yamlData, []byte("---\n"))
	s := runtime.NewScheme()
	scheme.AddToScheme(s)
	apiextensionsv1.AddToScheme(s)

	for i, yamlDoc := range yamlDocs {
		if len(yamlDoc) == 0 {
			continue
		}

		fmt.Printf("YAML Document %d:\n%s\n", i+1, yamlDoc)
		jsonData, err := yaml.YAMLToJSON(yamlDoc)
		if err != nil {
			fmt.Printf("Error converting YAML to JSON: %v\n", err)
			continue
		}

		obj := &unstructured.Unstructured{}
		dec := serializer.NewCodecFactory(s).UniversalDeserializer()
		_, _, err = dec.Decode(jsonData, nil, obj)
		if err != nil {
			fmt.Printf("Error decoding JSON to Unstructured object: %v", err)
			continue
		}

		gvk := obj.GroupVersionKind()
		restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gvk.GroupVersion()})
		restMapper.Add(gvk, meta.RESTScopeNamespace)
		mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return fmt.Errorf("Error getting REST mapping: %v", err)
		}
		var resourceClient dynamic.ResourceInterface

		ns := obj.GetNamespace()
		resourceClient = dynamicClient.Resource(mapping.Resource).Namespace(ns)
		if ns == "" {
			resourceClient = dynamicClient.Resource(mapping.Resource)
		}

		_, err = resourceClient.Create(context.TODO(), obj, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Error applying manifest: %v", err)
		}

		fmt.Printf("Manifest %v applied successfully!\n", manifestPath)
	}

	return nil
}