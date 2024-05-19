package common

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
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type K8sDynamicClient struct {
	scheme *runtime.Scheme
	client *dynamic.DynamicClient
}

func NewK8sDynamicClient(configPath string, apiServerPort uint16) (*K8sDynamicClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", ".kubeconfig")
	if err != nil {
		return nil, fmt.Errorf("Error building kubeconfig: %v", err)
	}
	config.Host = "https://127.0.0.1:" + fmt.Sprintf("%d", apiServerPort)
	config.Insecure = true
	config.CAData = []byte{}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating dynamic client: %v", err)
	}
	s := runtime.NewScheme()
	scheme.AddToScheme(s)
	apiextensionsv1.AddToScheme(s)

	return &K8sDynamicClient{
		client: dynamicClient,
		scheme: s,
	}, nil
}

func (c *K8sDynamicClient) Apply(manifestPath string) error {
	yamlData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("Error reading YAML file: %v", err)

	}

	yamlDocs := bytes.Split(yamlData, []byte("---\n"))
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
		dec := serializer.NewCodecFactory(c.scheme).UniversalDeserializer()
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
		resourceClient = c.client.Resource(mapping.Resource).Namespace(ns)
		if ns == "" {
			resourceClient = c.client.Resource(mapping.Resource)
		}

		_, err = resourceClient.Create(context.TODO(), obj, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Error applying manifest: %v", err)
		}

		fmt.Printf("Manifest %v applied successfully!\n", manifestPath)
	}

	return nil
}
