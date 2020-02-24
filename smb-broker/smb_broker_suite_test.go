package main_test

import (
	local_k8s_cluster "github.com/DennisDenuto/smb-volume-k8s-local-cluster"
	"io"
	"io/ioutil"
	"os"
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

	templateHelmOutput := local_k8s_cluster.HelmStdout("template", "smb-broker", "./helm", "--set", "image_repo_url=localhost:5000", "--set", "ingress.enabled=true", "--set", "targetNamespace="+namespace, "--set", "ingress.hosts[0].host=localhost", "--set", "ingress.hosts[0].paths={/v2}", "--set", "image.repository=cfpersi/smb-broker", "--set", "image.tag=local-test")
	f, err := ioutil.TempFile(os.TempDir(), "helm")
	Expect(err).NotTo(HaveOccurred())
	_, err = f.WriteString(templateHelmOutput)
	Expect(err).NotTo(HaveOccurred())
	Expect(f.Close()).To(Succeed())

	kbldOutput := local_k8s_cluster.KbldStdout("-f", f.Name())
	f2, err := ioutil.TempFile(os.TempDir(), "kbld")
	Expect(err).NotTo(HaveOccurred())
	_, err = f2.WriteString(kbldOutput)
	Expect(err).NotTo(HaveOccurred())
	Expect(f2.Close()).To(Succeed())
	local_k8s_cluster.Kubectl("apply", "-f", f2.Name())

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
