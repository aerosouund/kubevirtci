package nodes

type NodeLinuxConfig struct {
	NodeIdx      int
	K8sVersion   string
	FipsEnabled  bool
	DockerProxy  string
	EtcdInMemory bool
	EtcdSize     string
	SingleStack  bool
	EnableAudit  bool
	GpuAddress   string
	Realtime     bool
	PSA          bool
}

func NewNodeLinuxConfig(
	nodeIdx int,
	k8sVersion string,
	fipsEnabled bool,
	dockerProxy string,
	etcdInMemory bool,
	etcdSize string,
	singleStack bool,
	enableAudit bool,
	gpuAddress string,
	realtime bool,
	psa bool,
) *NodeLinuxConfig {
	return &NodeLinuxConfig{
		NodeIdx:      nodeIdx,
		K8sVersion:   k8sVersion,
		FipsEnabled:  fipsEnabled,
		DockerProxy:  dockerProxy,
		EtcdInMemory: etcdInMemory,
		EtcdSize:     etcdSize,
		SingleStack:  singleStack,
		EnableAudit:  enableAudit,
		GpuAddress:   gpuAddress,
		Realtime:     realtime,
		PSA:          psa,
	}
}
