package cmd

import (
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	nodesconfig "kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/config"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/aaq"
	bindvfio "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/bind-vfio"
	etcdinmemory "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/etcd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/istio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/nfscsi"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/node01"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/psa"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/rookceph"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

type TestSuite struct {
	suite.Suite
	sshClient *kubevirtcimocks.MockSSHClient
	k8sClient *k8s.K8sDynamicClientImpl
}

func (ts *TestSuite) SetupTest() {
	reactors := []k8s.ReactorConfig{
		k8s.NewReactorConfig("create", "istiooperators", istio.IstioReactor),
		k8s.NewReactorConfig("create", "cephblockpools", rookceph.CephReactor),
		k8s.NewReactorConfig("create", "persistentvolumeclaims", nfscsi.NfsCsiReactor),
	}

	ts.k8sClient = k8s.NewTestClient(reactors...)
	ts.sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(ts.T()))
}

func (ts *TestSuite) TearDownTest() {
	ts.sshClient = nil
}

func (ts *TestSuite) TestProvisionNode() {
	linuxConfigFuncs := []nodesconfig.LinuxConfigFunc{
		nodesconfig.WithEtcdInMemory(true),
		nodesconfig.WithEtcdSize("512M"),
		nodesconfig.WithPSA(true),
	}

	n := nodesconfig.NewNodeLinuxConfig(1, "k8s-1.30", linuxConfigFuncs)

	etcdinmemory.AddExpectCalls(ts.sshClient, "512M")
	bindvfio.AddExpectCalls(ts.sshClient, "8086:2668")
	bindvfio.AddExpectCalls(ts.sshClient, "8086:2415")
	psa.AddExpectCalls(ts.sshClient)
	node01.AddExpectCalls(ts.sshClient)

	err := provisionNode(ts.sshClient, n)
	ts.NoError(err)
}

func (ts *TestSuite) TestProvisionNodeK8sOpts() {
	k8sConfs := []nodesconfig.K8sConfigFunc{
		nodesconfig.WithCeph(true),
		nodesconfig.WithPrometheus(true),
		nodesconfig.WithAlertmanager(true),
		nodesconfig.WithGrafana(true),
		nodesconfig.WithIstio(true),
		nodesconfig.WithNfsCsi(true),
		nodesconfig.WithAAQ(true),
	}
	n := nodesconfig.NewNodeK8sConfig(k8sConfs)

	istio.AddExpectCalls(ts.sshClient)
	aaq.AddExpectCalls(ts.sshClient)

	err := provisionK8sOptions(ts.sshClient, ts.k8sClient, n, "k8s-1.30")
	ts.NoError(err)
}
