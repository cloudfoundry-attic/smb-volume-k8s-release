package main_test

import (
	local_k8s_cluster "code.cloudfoundry.org/smb-volume-k8s-local-cluster"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSmbBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmbBroker Suite")
}

var kubeConfigPath string
var nodeName string
var namespace string
var smbBrokerUsername string
var smbBrokerPassword string

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Minute)

	namespace = "smb-test-namespace"
	nodeName = "default-smb-broker-test-node"
	kubeConfigPath = "/tmp/kubeconfig"
	smbBrokerUsername = "smb-broker-username"
	smbBrokerPassword = "smb-broker-password"

	local_k8s_cluster.CreateK8sCluster(nodeName, kubeConfigPath, os.Getenv("K8S_IMAGE"))

	local_k8s_cluster.Kubectl("create", "namespace", namespace)
	smbBrokerDeploymentYaml := local_k8s_cluster.YttStdout("-f", "./ytt", "-v", "smbBrokerUsername="+smbBrokerUsername, "-v", "smbBrokerPassword="+smbBrokerPassword, "-v", "namespace="+namespace, "-v", "image.repository=registry:5000/cfpersi/smb-broker", "-v", "image.tag=local-test")
	local_k8s_cluster.KappWithStringAsStdIn("-y", "deploy", "-a", "smb-broker", "-f")(smbBrokerDeploymentYaml)

	By("pulling the smb-broker into the docker daemon", func() {
		local_k8s_cluster.Docker("pull", "localhost:5000/cfpersi/smb-broker:local-test")
	})

	var smbBrokerDestination string
	var found bool
	if smbBrokerDestination, found = os.LookupEnv("SMB_BROKER_IMAGE_DESTINATION"); !found {
		smbBrokerDestination = "/tmp/smb-broker.tgz"
	}

	local_k8s_cluster.Docker("save", "localhost:5000/cfpersi/smb-broker:local-test", "-o", smbBrokerDestination)
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster(nodeName, kubeConfigPath)
})

func assertHttpResponseContainsSubstring(body io.Reader, expected string) {
	bytes, err := ioutil.ReadAll(body)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(bytes)).Should(ContainSubstring(expected))
}
