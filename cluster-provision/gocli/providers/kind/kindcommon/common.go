package kindcommon

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"

	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
	dockercri "kubevirt.io/kubevirtci/cluster-provision/gocli/cri/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

//go:embed manifests/*
var f embed.FS

type KindCommonProvider struct {
	Client k8s.K8sDynamicClient
	CRI    cri.ContainerClient

	nodes           int
	version         string
	provider        *kind.Provider
	runEtcdOnMemory bool
	ipFamily        string
	withCPUManager  bool
	registryProxy   string
}

const (
	kind128Image      = "kindest/node:v1.28.0@sha256:b7a4cad12c197af3ba43202d3efe03246b3f0793f162afb40a33c923952d5b31"
	cniArchieFilename = "cni-archive.tar.gz"
	registryImage     = "quay.io/kubevirtci/library-registry:2.7.1"
)

func NewKindCommondProvider(version string, nodeNum int) (*KindCommonProvider, error) {
	// use podman first
	providerCRIOpt, err := kind.DetectNodeProvider()
	if err != nil {
		return nil, err
	}

	k := kind.NewProvider(providerCRIOpt)
	return &KindCommonProvider{
		CRI:      dockercri.NewDockerClient(),
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
	logrus.Infof("Kind %s base cluster started\n", k.version)

	kubeconf, err := k.provider.KubeConfig(k.version, true)
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
	nodes, err := k.provider.ListNodes(k.version)
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	err = k.downloadCNI()
	if err != nil {
		return nil
	}

	k.CRI = dockercri.NewDockerClient()

	_, registryIP, err := k.runRegistry("5000") // read from flag
	if err != nil {
		return err
	}

	for _, node := range nodes {
		da := docker.NewDockerAdapter(cli, node.String())
		if err := k.setupCNI(da); err != nil {
			return err
		}
		if err = k.setupRegistryOnNode(da, registryIP); err != nil {
			return err
		}
		if err = k.setupNetwork(da); err != nil {
			return err
		}
		if k.registryProxy != "" {
			if err = k.setupRegistryProxy(da); err != nil {
				return err
			}
		}
	}

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

func (k *KindCommonProvider) Delete() error {
	if err := k.provider.Delete(k.version, ""); err != nil {
		return err
	}
	if err := k.deleteRegistry(); err != nil {
		return err
	}
	return nil
}

func (k *KindCommonProvider) setupNetwork(da *docker.DockerAdapter) error {
	cmds := []string{
		"modprobe br_netfilter",
		"sysctl -w net.bridge.bridge-nf-call-arptables=1",
		"sysctl -w net.bridge.bridge-nf-call-iptables=1",
		"sysctl -w net.bridge.bridge-nf-call-ip6tables=1",
	}

	for _, cmd := range cmds {
		if _, err := da.SSH(cmd, true); err != nil {
			return err
		}
	}
	return nil
}

func (k *KindCommonProvider) setupRegistryOnNode(da *docker.DockerAdapter, registryIP string) error {
	cmds := []string{
		"echo " + registryIP + "\tregistry | tee -a /etc/hosts",
	}
	for _, cmd := range cmds {
		if _, err := da.SSH(cmd, true); err != nil {
			return err
		}
	}
	return nil
}

func (k *KindCommonProvider) setupCNI(da *docker.DockerAdapter) error {
	file, err := os.Open(cniArchieFilename)
	if err != nil {
		return err
	}

	err = da.SCP("/opt/cni/bin", file)
	if err != nil {
		return err
	}
	return nil
}

func (k *KindCommonProvider) setupRegistryProxy(da *docker.DockerAdapter) error {
	setupUrl := "http://" + k.registryProxy + ":3128/setup/systemd"
	cmds := []string{
		"curl " + setupUrl + " > proxyscript.sh",
		"sed s/docker.service/containerd.service/g proxyscript.sh",
		`sed '/Environment/ s/$/ \"NO_PROXY=127.0.0.0\/8,10.0.0.0\/8,172.16.0.0\/12,192.168.0.0\/16\"/ proxyscript.sh`,
		"/bin/bash -c proxyscript.sh",
	}
	for _, cmd := range cmds {
		if _, err := da.SSH(cmd, true); err != nil {
			return err
		}
	}
	return nil
}

func (k *KindCommonProvider) deleteRegistry() error {
	return k.CRI.Remove(k.version + "-registry")
}

func (k *KindCommonProvider) runRegistry(hostPort string) (string, string, error) {
	registryID, err := k.CRI.Create(registryImage, &cri.CreateOpts{
		Name:          k.version + "-registry",
		Privileged:    true,
		Network:       "kind",
		RestartPolicy: "always",
		Ports: map[string]string{
			"5000": hostPort,
		},
	})
	if err != nil {
		return "", "", err
	}

	if err := k.CRI.Start(registryID); err != nil {
		return "", "", err
	}

	// check if this will work for podman
	registryJSON := []types.ContainerJSON{}

	jsonData, err := k.CRI.Inspect(registryID)
	if err != nil {
		return "", "", err
	}

	err = json.Unmarshal(jsonData, &registryJSON)
	if err != nil {
		return "", "", err
	}

	return registryID, registryJSON[0].NetworkSettings.Networks["kind"].IPAddress, nil
}

func (k *KindCommonProvider) downloadCNI() error {
	out, err := os.Create(cniArchieFilename)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get("https://github.com/containernetworking/plugins/releases/download/v0.8.5/cni-plugins-linux-" + runtime.GOARCH + "-v0.8.5.tgz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	logrus.Info("Downloaded cni archive")
	return nil
}
