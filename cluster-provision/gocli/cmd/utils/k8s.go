package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
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
		homeDir(), ".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	// Create a dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	// Load the YAML manifest
	yamlFile := "path/to/your/manifest.yaml"
	yamlData, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	// Convert YAML to JSON
	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		log.Fatalf("Error converting YAML to JSON: %v", err)
	}

	// Decode the JSON into an Unstructured object
	obj := &unstructured.Unstructured{}
	dec := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	_, _, err = dec.Decode(jsonData, nil, obj)
	if err != nil {
		log.Fatalf("Error decoding JSON to Unstructured object: %v", err)
	}

	// Get the GVR (Group, Version, Resource) from the object
	gvk := obj.GroupVersionKind()
	restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gvk.GroupVersion()})
	restMapper.Add(gvk, meta.RESTScopeNamespace)
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		log.Fatalf("Error getting REST mapping: %v", err)
	}

	// Apply the manifest to the cluster
	resourceClient := dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	_, err = resourceClient.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Error applying manifest: %v", err)
	}

	fmt.Println("Manifest applied successfully!")
}
