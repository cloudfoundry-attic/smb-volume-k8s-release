package main_test

import (
	local_k8s_cluster "code.cloudfoundry.org/local-k8s-cluster"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test", func() {
	It("hi", func() {
		println("running")
		Expect(local_k8s_cluster.CreateKpackImageResource()).NotTo(HaveOccurred())
	})
})
