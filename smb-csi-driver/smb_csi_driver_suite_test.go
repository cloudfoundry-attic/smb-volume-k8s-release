package main_test

import (
	"testing"

	"github.com/kubernetes-csi/csi-test/v3/pkg/sanity"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSmbCsiDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmbCsiDriver Suite")

	config := sanity.NewTestConfig()
	// Set configuration options as needed
	config.Address = ""

	// Now call the test suite
	sanity.Test(t, config)
}
