package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func K8sApply(config *rest.Config, manifestPath string) error {
	// dynamicClient, err := dynamic.NewForConfig(config)
	// if err != nil {
	// 	log.Fatalf("Error creating dynamic client: %v", err)
	// }

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	// List namespaces

	yamlData, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		log.Fatalf("Error converting YAML to JSON: %v", err)
	}

	obj := &appsv1.Deployment{}
	dec := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
	_, _, err = dec.Decode(jsonData, nil, obj)
	if err != nil {
		log.Fatalf("Error decoding JSON to Unstructured object: %v", err)
	}
	deployments := clientset.AppsV1().Deployments(obj.GetNamespace())
	// if err != nil {
	// 	log.Fatalf("Error listing namespaces: %v", err)
	// }

	// gvk := obj.GroupVersionKind()
	// restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gvk.GroupVersion()})
	// restMapper.Add(gvk, meta.RESTScopeNamespace)
	// mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	// if err != nil {
	// 	log.Fatalf("Error getting REST mapping: %v", err)
	// }

	// resourceClient := dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	// _, err = resourceClient.Create(context.TODO(), obj, metav1.CreateOptions{})
	// if err != nil {
	// 	log.Fatalf("Error applying manifest: %v", err)
	// }
	obj, err = deployments.Create(context.TODO(), obj, metav1.CreateOptions{})

	fmt.Println("Manifest applied successfully!")
	return nil
}
