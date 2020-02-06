package integration_tests_test

import (
	local_k8s_cluster "code.cloudfoundry.org/local-k8s-cluster"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTests Suite")
}

var _ = BeforeSuite(func() {
	local_k8s_cluster.CreateK8sCluster("test", "/tmp/kubeconfig")
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster("test", "/tmp/kubeconfig")
})