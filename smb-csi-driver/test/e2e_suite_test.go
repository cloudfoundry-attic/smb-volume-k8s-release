package test_test

import (
	local_k8s_cluster "github.com/DennisDenuto/smb-volume-k8s-local-cluster"
	"github.com/onsi/gomega/gexec"
	"k8s.io/kubernetes/test/e2e/framework"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)
var smbCsiDriverPath string
var session *gexec.Session
var smbBrokerCompiledPath string
var kubeConfigPath string
var nodeName string

func init() {
	contextType := &framework.TestContext
	contextType.KubeConfig = "/tmp/csi-kubeconfig"
	contextType.KubectlPath = "/usr/local/bin/kubectl"
	framework.AfterReadingAllFlags(contextType)
}

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Minute)
	nodeName = "default-smb-csi-driver-test-node"
	kubeConfigPath = "/tmp/csi-kubeconfig"

	local_k8s_cluster.CreateK8sCluster(nodeName, kubeConfigPath)

	local_k8s_cluster.Kubectl("apply", "--kustomize", "../base")
	Eventually("/tmp/csi.sock").Should(BeAnExistingFile())
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster(nodeName, kubeConfigPath)
})