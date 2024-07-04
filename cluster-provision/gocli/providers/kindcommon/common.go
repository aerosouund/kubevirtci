package kindcommon

import (
	"context"
	"embed"
	"encoding/json"

	"k8s.io/client-go/rest"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	kind "sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/yaml"
)

//go:embed manifests/*
var f embed.FS

type KindCommonProvider struct {
	Client k8s.K8sDynamicClient

	nodes           int
	version         string
	provider        *kind.Provider
	runEtcdOnMemory bool
	ipFamily        string
	withCPUManager  bool
}

const kind128Image = "kindest/node:v1.28.0@sha256:b7a4cad12c197af3ba43202d3efe03246b3f0793f162afb40a33c923952d5b31"

func NewKindCommondProvider(version string, nodeNum int) (*KindCommonProvider, error) {
	// use podman first
	providerCRIOpt, err := kind.DetectNodeProvider()
	if err != nil {
		return nil, err
	}

	k := kind.NewProvider(providerCRIOpt)
	return &KindCommonProvider{
		nodes:    nodeNum,
		provider: k,
		version:  version,
	}, nil
}

func (k *KindCommonProvider) Start(ctx context.Context, cancel context.CancelFunc) error {
	cluster, err := k.prepareClusterYaml()
	if err != nil {
		return err
	}

	err = k.provider.Create(k.version, kind.CreateWithRawConfig([]byte(cluster)), kind.CreateWithNodeImage(kind128Image))
	if err != nil {
		return err
	}

	kubeconf, err := k.provider.KubeConfig("kubevirt", true)
	if err != nil {
		return err
	}

	jsonData, err := yaml.YAMLToJSON([]byte(kubeconf))
	if err != nil {
		return err
	}
	config := &rest.Config{}
	err = json.Unmarshal(jsonData, config)
	if err != nil {
		return err
	}

	k8sClient, err := k8s.NewDynamicClient(config)
	if err != nil {
		return err
	}
	k.Client = k8sClient

	return nil
}

func (k *KindCommonProvider) prepareClusterYaml() (string, error) {
	cluster, err := f.ReadFile("manifests/kind.yaml")
	if err != nil {
		return "", err
	}

	wp, err := f.ReadFile("manifests/worker-patch.yaml")
	if err != nil {
		return "", err
	}

	cpump, err := f.ReadFile("manifests/cpu-manager-patch.yaml")
	if err != nil {
		return "", err
	}

	ipf, err := f.ReadFile("manifests/ip-family.yaml")
	if err != nil {
		return "", err
	}

	// cpu manager condition
	for i := 0; i < k.nodes; i++ {
		cluster = append(cluster, wp...)
		cluster = append(cluster, []byte("\n")...)
		if k.withCPUManager {
			cluster = append(cluster, cpump...)
		}
	}

	if k.ipFamily != "" {
		cluster = append(cluster, []byte(string(ipf)+k.ipFamily)...)
	}
	return string(cluster), nil
}

func (k *KindCommonProvider) Delete(prefix string) error {
	n, err := k.provider.ListNodes(prefix)
	if err != nil {
		return err
	}

	if len(n) > 0 {
		err = k.provider.Delete(prefix, "")
		if err != nil {
			return err
		}
	}
	return nil
}
