package sriov

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

var _ = Describe("SR-IOV functionality", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		ks        *KindSriov
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
		sshClient = nil
	})

	Describe("fetchNodePfs", func() {
		It("should execute the correct commands", func() {
			sshClient.EXPECT().Command("grep vfio_pci /proc/modules", false).Return("vfio-pci", nil)
			sshClient.EXPECT().Command("modprobe -i vfio_pci", true)
			sshClient.EXPECT().Command("find /sys/class/net/*/device/", false).Return("/sys/class/net/eth0/device", nil)

			_, err := ks.fetchNodePfs(sshClient)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
