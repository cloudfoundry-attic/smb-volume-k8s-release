package main_test

import (
	"github.com/kubernetes-csi/csi-test/v3/pkg/sanity"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Main", func() {
	config := sanity.NewTestConfig()
	config.Address = "localhost:2910"
	sanity.GinkgoTest(&config)
})
