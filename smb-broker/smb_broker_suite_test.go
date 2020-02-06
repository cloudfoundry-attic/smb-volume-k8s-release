package main_test

import (
	local_k8s_cluster "code.cloudfoundry.org/local-k8s-cluster"
	"io"
	"io/ioutil"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSmbBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmbBroker Suite")
}

var kubeConfigPath string
var nodeName string

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Minute)
	nodeName = "default-smb-broker-test-node"
	kubeConfigPath = "/tmp/kubeconfig"

	local_k8s_cluster.CreateK8sCluster(nodeName, kubeConfigPath)

	local_k8s_cluster.Helm("install", "smb-broker", "./helm", "--set", "ingress.hosts[0].host=localhost", "--set", "ingress.hosts[0].paths={/v2}", "--set", "ingress.enabled=true", "--set", "image.repository=registry:5000/cfpersi/smb-broker", "--set", "image.tag=local-test")
	local_k8s_cluster.Kubectl("create", "namespace", "eirini")
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster(nodeName, kubeConfigPath)
})

func assertHttpResponseContainsSubstring(body io.Reader, expected string) {
	bytes, err := ioutil.ReadAll(body)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(bytes)).Should(ContainSubstring(expected))
}
