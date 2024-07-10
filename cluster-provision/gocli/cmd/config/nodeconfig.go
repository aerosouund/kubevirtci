package config

// NodeLinuxConfig type is a holder for all the config params that a node can have for its linux system
type NodeLinuxConfig struct {
	NodeIdx         int
	K8sVersion      string
	FipsEnabled     bool
	DockerProxy     string
	EtcdInMemory    bool
	EtcdSize        string
	SingleStack     bool
	EnableAudit     bool
	GpuAddress      string
	Realtime        bool
	PSA             bool
	KsmEnabled      bool
	SwapEnabled     bool
	KsmPageCount    int
	KsmScanInterval int
	Swapiness       int
	UnlimitedSwap   bool
	SwapSize        string
}

// NodeK8sConfig type is a holder for all the config k8s options for kubevirt cluster
type NodeK8sConfig struct {
	Ceph         bool
	Prometheus   bool
	Alertmanager bool
	Grafana      bool
	Istio        bool
	NfsCsi       bool
	Cnao         bool
	Multus       bool
	Cdi          bool
	CdiVersion   string
	AAQ          bool
	AAQVersion   string
}

func NewNodeK8sConfig(confs []K8sConfigFunc) *NodeK8sConfig {
	n := &NodeK8sConfig{}

	for _, conf := range confs {
		conf(n)
	}

	return n
}

func NewNodeLinuxConfig(nodeIdx int, k8sVersion string, confs []LinuxConfigFunc) *NodeLinuxConfig {
	n := &NodeLinuxConfig{
		NodeIdx:    nodeIdx,
		K8sVersion: k8sVersion,
	}

	for _, conf := range confs {
		conf(n)
	}

	return n
}
