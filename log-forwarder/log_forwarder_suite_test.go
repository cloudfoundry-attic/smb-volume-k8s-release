package main_test

import (
	local_k8s_cluster "code.cloudfoundry.org/smb-volume-k8s-local-cluster"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogForwarder Suite")
}

var kubeConfigPath string
var nodeName string
var namespace string
var targetNamespace string
var smbBrokerUsername string
var smbBrokerPassword string

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Minute)

	targetNamespace = "cf-workloads"
	namespace = "cf-smb"
	nodeName = "log-forwarder"
	kubeConfigPath = "/tmp/kubeconfig"

	local_k8s_cluster.CreateK8sCluster(nodeName, kubeConfigPath, os.Getenv("K8S_IMAGE"))

	By("creating namespaces", func(){
		local_k8s_cluster.Kubectl("create", "namespace", namespace)
		local_k8s_cluster.Kubectl("create", "namespace", targetNamespace)
	})

	By("Deploying broker", func(){
		logForwarderDeploymentYaml := local_k8s_cluster.YttStdout("-f", "./ytt", "-v", "namespace="+namespace, "-v", "image.repository=registry:5000/cfpersi/log-forwarder", "-v", "image.tag=local-test")
		local_k8s_cluster.KappWithStringAsStdIn("-y", "deploy", "-a", "volume-services-log-forwarder", "-f")(logForwarderDeploymentYaml)
	})
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster(nodeName, kubeConfigPath)
})
