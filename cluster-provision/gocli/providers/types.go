package providers

import k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"

type KubevirtProvider struct {
	IsRunning bool
	Client    *k8s.K8sDynamicClient

	Version string
	Nodes   uint   `flag:"nodes" short:"n"`
	Numa    uint   `flag:"numa" short:"u"`
	Memory  string `flag:"memory" short:"m"`
	CPU     uint   `flag:"cpu" short:"c"`

	SecondaryNics                uint     `flag:"secondary-nics"`
	QemuArgs                     string   `flag:"qemu-args"`
	KernelArgs                   string   `flag:"kernel-args"`
	Background                   bool     `flag:"background" short:"b"`
	Reverse                      bool     `flag:"reverse" short:"r"`
	RandomPorts                  bool     `flag:"random-ports"`
	Slim                         bool     `flag:"slim"`
	VNCPort                      uint     `flag:"vnc-port"`
	HTTPPort                     uint     `flag:"http-port"`
	HTTPSPort                    uint     `flag:"https-port"`
	RegistryPort                 uint     `flag:"registry-port"`
	OCPort                       uint     `flag:"ocp-port"`
	K8sPort                      uint     `flag:"k8s-port"`
	SSHPort                      uint     `flag:"ssh-port"`
	PrometheusPort               uint     `flag:"prometheus-port"`
	GrafanaPort                  uint     `flag:"grafana-port"`
	DNSPort                      uint     `flag:"dns-port"`
	NFSData                      string   `flag:"nfs-data"`
	EnableCeph                   bool     `flag:"enable-ceph"`
	EnableIstio                  bool     `flag:"enable-istio"`
	EnableCNAO                   bool     `flag:"enable-cnao"`
	EnableNFSCSI                 bool     `flag:"enable-nfs-csi"`
	EnablePrometheus             bool     `flag:"enable-prometheus"`
	EnablePrometheusAlertManager bool     `flag:"enable-prometheus-alertmanager"`
	EnableGrafana                bool     `flag:"enable-grafana"`
	DockerProxy                  string   `flag:"docker-proxy"`
	ContainerRegistry            string   `flag:"container-registry"`
	ContainerOrg                 string   `flag:"container-org"`
	ContainerSuffix              string   `flag:"container-suffix"`
	GPU                          string   `flag:"gpu"`
	NvmeDisks                    []string `flag:"nvme"`
	ScsiDisks                    []string `flag:"scsi"`
	RunEtcdOnMemory              bool     `flag:"run-etcd-on-memory"`
	EtcdCapacity                 string   `flag:"etcd-capacity"`
	Hugepages2M                  uint     `flag:"hugepages-2m"`
	EnableRealtimeScheduler      bool     `flag:"enable-realtime-scheduler"`
	EnableFIPS                   bool     `flag:"enable-fips"`
	EnablePSA                    bool     `flag:"enable-psa"`
	SingleStack                  bool     `flag:"single-stack"`
	EnableAudit                  bool     `flag:"enable-audit"`
	USBDisks                     []string `flag:"usb"`
}

type KubevirtProviderOption func(*KubevirtProvider)
