package main_test

import (
	local_k8s_cluster "github.com/DennisDenuto/smb-volume-k8s-local-cluster"
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
var namespace string

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Minute)

	namespace = "smb-test-namespace"
	nodeName = "default-smb-broker-test-node"
	kubeConfigPath = "/tmp/kubeconfig"

	local_k8s_cluster.CreateK8sCluster(nodeName, kubeConfigPath)
	err := local_k8s_cluster.CreateKpackImageResource()
	Expect(err).NotTo(HaveOccurred())

	println(local_k8s_cluster.HelmStdout("install",
		"smb-broker",
		"./helm",
		"--set",
		"ingress.enabled=true",
		"--set",
		"targetNamespace="+namespace,
		"--set",
		"ingress.hosts[0].host=localhost",
		"--set",
		"ingress.hosts[0].paths={/v2}",
		"--set",
		"image.repository=registry:5000/cfpersi/smb-broker",
		"--set",
		"image.tag=latest"))

	local_k8s_cluster.Kubectl("create", "namespace", namespace)
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster(nodeName, kubeConfigPath)
})

func assertHttpResponseContainsSubstring(body io.Reader, expected string) {
	bytes, err := ioutil.ReadAll(body)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(bytes)).Should(ContainSubstring(expected))
}
