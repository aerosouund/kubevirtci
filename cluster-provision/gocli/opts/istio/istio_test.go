package istio

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

func TestIstioOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IstioOpt Suite")
}

var _ = Describe("IstioOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		k8sclient k8s.K8sDynamicClient
		opt       *istioDeployOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		r := k8s.NewReactorConfig("create", "istiooperators", IstioReactor)
		k8sclient = k8s.NewTestClient(r)
		opt = NewIstioDeployOpt(&libssh.SSHClientImpl{}, k8sclient, false)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute IstioOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
