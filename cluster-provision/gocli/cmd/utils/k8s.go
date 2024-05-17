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

type KubeConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Cluster struct {
			Server    string `yaml:"server"`
			TLSVerify bool   `yaml:"insecure-skip-tls-verify"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	} `yaml:"clusters"`
	Contexts []struct {
		Context struct {
			Cluster string `yaml:"cluster"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string   `yaml:"current-context"`
	Kind           string   `yaml:"kind"`
	Preferences    struct{} `yaml:"preferences"`
	Users          []struct {
		Name string `yaml:"name"`
		User struct {
			InsecureSkipTLSVerify bool `yaml:"insecure-skip-tls-verify"`
		} `yaml:"user"`
	} `yaml:"users"`
}

func PrepareKubeconf(configPath string, serverPort uint16) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var kubeconfig KubeConfig
	err = yaml.Unmarshal(data, &kubeconfig)
	if err != nil {
		return err
	}

	kubeconfig.Clusters[0].Cluster.Server = "https://127.0.0.1:" + fmt.Sprintf("%d", serverPort)
	kubeconfig.Clusters[0].Cluster.TLSVerify = true

	updatedData, err := yaml.Marshal(&kubeconfig)
	if err != nil {
		log.Fatalf("Error marshalling kubeconfig: %v", err)
	}

	err = os.WriteFile(configPath, updatedData, 0644)
	if err != nil {
		log.Fatalf("Error writing updated kubeconfig file: %v", err)
	}

	fmt.Println("Kubeconfig file updated successfully.")

	return nil
}

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
