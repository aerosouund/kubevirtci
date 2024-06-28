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
