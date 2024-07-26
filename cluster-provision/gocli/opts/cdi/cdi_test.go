package cdi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestCdiOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CdiOpt Suite")
}

var _ = Describe("CdiOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		client    k8s.K8sDynamicClient
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *CdiOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		client = k8s.NewTestClient()
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewCdiOpt(client, sshClient, "")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute CdiOpt successfully", func() {
		sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pod --timeout=180s --all --namespace cdi")
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
