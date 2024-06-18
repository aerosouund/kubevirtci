package cmd

import (
	_ "embed"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/providers"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	sshutils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/cnao"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/istio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/multus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/nfscsi"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/prometheus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/rookceph"
)

// NewRunCommand returns command that runs given cluster
func NewRun2Command() *cobra.Command {

	run := &cobra.Command{
		Use:   "run",
		Short: "run starts a given cluster",
		RunE:  run,
		Args:  cobra.ExactArgs(1),
	}
	run.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	run.Flags().UintP("numa", "u", 1, "number of NUMA nodes per node")
	run.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	run.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	run.Flags().UintP("secondary-nics", "", 0, "number of secondary nics to add")
	run.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	run.Flags().String("kernel-args", "", "additional kernel args to pass through to the nodes")
	run.Flags().BoolP("background", "b", false, "go to background after nodes are up")
	run.Flags().BoolP("reverse", "r", false, "revert node startup order")
	run.Flags().Bool("random-ports", true, "expose all ports on random localhost ports")
	run.Flags().Bool("slim", false, "use the slim flavor")
	run.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	run.Flags().Uint("http-port", 0, "port on localhost for http")
	run.Flags().Uint("https-port", 0, "port on localhost for https")
	run.Flags().Uint("registry-port", 0, "port on localhost for the docker registry")
	run.Flags().Uint("ocp-port", 0, "port on localhost for the ocp cluster")
	run.Flags().Uint("k8s-port", 0, "port on localhost for the k8s cluster")
	run.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	run.Flags().Uint("prometheus-port", 0, "port on localhost for prometheus server")
	run.Flags().Uint("grafana-port", 0, "port on localhost for grafana server")
	run.Flags().Uint("dns-port", 0, "port on localhost for dns server")
	run.Flags().String("nfs-data", "", "path to data which should be exposed via nfs to the nodes")
	run.Flags().Bool("enable-ceph", false, "enables dynamic storage provisioning using Ceph")
	run.Flags().Bool("enable-istio", false, "deploys Istio service mesh")
	run.Flags().Bool("enable-cnao", false, "enable network extensions with istio")
	run.Flags().Bool("deploy-cnao", false, "deploy the network extensions operator")
	run.Flags().Bool("deploy-multus", false, "deploy multus")
	run.Flags().Bool("enable-nfs-csi", false, "deploys nfs csi dynamic storage")
	run.Flags().Bool("enable-prometheus", false, "deploys Prometheus operator")
	run.Flags().Bool("enable-prometheus-alertmanager", false, "deploys Prometheus alertmanager")
	run.Flags().Bool("enable-grafana", false, "deploys Grafana")
	run.Flags().String("docker-proxy", "", "sets network proxy for docker daemon")
	run.Flags().String("container-registry", "quay.io", "the registry to pull cluster container from")
	run.Flags().String("container-org", "kubevirtci", "the organization at the registry to pull the container from")
	run.Flags().String("container-suffix", "", "Override container suffix stored at the cli binary")
	run.Flags().String("gpu", "", "pci address of a GPU to assign to a node")
	run.Flags().StringArrayVar(&nvmeDisks, "nvme", []string{}, "size of the emulate NVMe disk to pass to the node")
	run.Flags().StringArrayVar(&scsiDisks, "scsi", []string{}, "size of the emulate SCSI disk to pass to the node")
	run.Flags().Bool("run-etcd-on-memory", false, "configure etcd to run on RAM memory, etcd data will not be persistent")
	run.Flags().String("etcd-capacity", "512M", "set etcd data mount size.\nthis flag takes affect only when 'run-etcd-on-memory' is specified")
	run.Flags().Uint("hugepages-2m", 64, "number of hugepages of size 2M to allocate")
	run.Flags().Bool("enable-realtime-scheduler", false, "configures the kernel to allow unlimited runtime for processes that require realtime scheduling")
	run.Flags().Bool("enable-fips", false, "enables FIPS")
	run.Flags().Bool("enable-psa", false, "Pod Security Admission")
	run.Flags().Bool("single-stack", false, "enable single stack IPv6")
	run.Flags().Bool("enable-audit", false, "enable k8s audit for all metadata events")
	run.Flags().StringArrayVar(&usbDisks, "usb", []string{}, "size of the emulate USB disk to pass to the node")
	return run
}

func run2(cmd *cobra.Command, args []string) (retErr error) {
	opts := []providers.KubevirtProviderOption{}
	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	nodes, err := cmd.Flags().GetUint("nodes")
	if err != nil {
		return err
	}
	opts = append(opts, providers.WithNodes(nodes))

	memory, err := cmd.Flags().GetString("memory")
	if err != nil {
		return err
	}
	resource.MustParse(memory)

	randomPorts, err := cmd.Flags().GetBool("random-ports")
	if err != nil {
		return err
	}
	opts = append(opts, providers.WithRandomPorts(randomPorts))

	slim, err := cmd.Flags().GetBool("slim")
	if err != nil {
		return err
	}
	opts = append(opts, providers.WithSlim(slim))

	portMap := nat.PortMap{}
	utils.AppendTCPIfExplicit(portMap, utils.PortSSH, cmd.Flags(), "ssh-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortVNC, cmd.Flags(), "vnc-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortHTTP, cmd.Flags(), "http-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortHTTPS, cmd.Flags(), "https-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortAPI, cmd.Flags(), "k8s-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortOCP, cmd.Flags(), "ocp-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortRegistry, cmd.Flags(), "registry-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortPrometheus, cmd.Flags(), "prometheus-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortGrafana, cmd.Flags(), "grafana-port")
	utils.AppendUDPIfExplicit(portMap, utils.PortDNS, cmd.Flags(), "dns-port")

	qemuArgs, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}
	kernelArgs, err := cmd.Flags().GetString("kernel-args")
	if err != nil {
		return err
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	numa, err := cmd.Flags().GetUint("numa")
	if err != nil {
		return err
	}

	secondaryNics, err := cmd.Flags().GetUint("secondary-nics")
	if err != nil {
		return err
	}

	nfsData, err := cmd.Flags().GetString("nfs-data")
	if err != nil {
		return err
	}

	dockerProxy, err := cmd.Flags().GetString("docker-proxy")
	if err != nil {
		return err
	}

	cephEnabled, err := cmd.Flags().GetBool("enable-ceph")
	if err != nil {
		return err
	}

	nfsCsiEnabled, err := cmd.Flags().GetBool("enable-nfs-csi")
	if err != nil {
		return err
	}

	istioEnabled, err := cmd.Flags().GetBool("enable-istio")
	if err != nil {
		return err
	}

	cnaoEnabled, err := cmd.Flags().GetBool("enable-cnao")
	if err != nil {
		return err
	}

	deployCnao, err := cmd.Flags().GetBool("deploy-cnao")
	if err != nil {
		return err
	}

	deployMultus, err := cmd.Flags().GetBool("deploy-multus")
	if err != nil {
		return err
	}

	prometheusEnabled, err := cmd.Flags().GetBool("enable-prometheus")
	if err != nil {
		return err
	}

	prometheusAlertmanagerEnabled, err := cmd.Flags().GetBool("enable-prometheus-alertmanager")
	if err != nil {
		return err
	}

	grafanaEnabled, err := cmd.Flags().GetBool("enable-grafana")
	if err != nil {
		return err
	}

	cluster := args[0]

	background, err := cmd.Flags().GetBool("background")
	if err != nil {
		return err
	}

	containerRegistry, err := cmd.Flags().GetString("container-registry")
	if err != nil {
		return err
	}
	gpuAddress, err := cmd.Flags().GetString("gpu")
	if err != nil {
		return err
	}

	containerOrg, err := cmd.Flags().GetString("container-org")
	if err != nil {
		return err
	}

	containerSuffix, err := cmd.Flags().GetString("container-suffix")
	if err != nil {
		return err
	}

	runEtcdOnMemory, err := cmd.Flags().GetBool("run-etcd-on-memory")
	if err != nil {
		return err
	}

	etcdDataMountSize, err := cmd.Flags().GetString("etcd-capacity")
	if err != nil {
		return err
	}
	resource.MustParse(etcdDataMountSize)

	hugepages2Mcount, err := cmd.Flags().GetUint("hugepages-2m")
	if err != nil {
		return err
	}
	realtimeSchedulingEnabled, err := cmd.Flags().GetBool("enable-realtime-scheduler")
	if err != nil {
		return err
	}
	psaEnabled, err := cmd.Flags().GetBool("enable-psa")
	if err != nil {
		return err
	}
	singleStack, err := cmd.Flags().GetBool("single-stack")
	if err != nil {
		return err
	}
	enableAudit, err := cmd.Flags().GetBool("enable-audit")
	if err != nil {
		return err
	}
	fipsEnabled, err := cmd.Flags().GetBool("enable-fips")
	if err != nil {
		return err
	}

	cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	image := fmt.Sprintf("")

	b := context.Background()
	ctx, cancel := context.WithCancel(b)
	kp := providers.NewKubevirtProvider("", "", cli, opts...)
	err = kp.Start(ctx, cancel, portMap)

	err = sshutils.CopyRemoteFile(kp.SSHPort, "/etc/kubernetes/admin.conf", ".kubeconfig")
	if err != nil {
		panic(err)
	}

	config, err := k8s.InitConfig(".kubeconfig", kp.APIServerPort)
	if err != nil {
		panic(err)
	}

	k8sClient, err := k8s.NewDynamicClient(config)
	if err != nil {
		panic(err)
	}

	if cephEnabled {
		cephOpt := rookceph.NewCephOpt(k8sClient)
		if err := cephOpt.Exec(); err != nil {
			panic(err)
		}
	}

	if nfsCsiEnabled {
		csiOpt := nfscsi.NewNfsCsiOpt(k8sClient)
		if err := csiOpt.Exec(); err != nil {
			panic(err)
		}
	}

	if deployMultus {
		multusOpt := multus.NewMultusOpt(k8sClient)
		if err := multusOpt.Exec(); err != nil {
			panic(err)
		}
	}

	if deployCnao {
		cnaoOpt := cnao.NewCnaoOpt(k8sClient)
		if err := cnaoOpt.Exec(); err != nil {
			panic(err)
		}
	}

	if istioEnabled {
		istioOpt := istio.NewIstioOpt(k8sClient, kp.SSHPort, cnaoEnabled)
		if err := istioOpt.Exec(); err != nil {
			panic(err)
		}
	}

	if prometheusEnabled {
		prometheusOpt := prometheus.NewPrometheusOpt(k8sClient, grafanaEnabled, prometheusAlertmanagerEnabled)
		if err = prometheusOpt.Exec(); err != nil {
			panic(err)
		}
	}

	return nil
}
