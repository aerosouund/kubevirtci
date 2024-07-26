package cmd

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSRIOV(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SR-IOV Test Suite")
}
