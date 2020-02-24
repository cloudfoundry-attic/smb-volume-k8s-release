package main_test

import (
	"github.com/DennisDenuto/smb-volume-k8s-local-cluster"
	"runtime"

	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var kubeConfigPath string
var nodeName string

func TestSmbCsiDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = BeforeSuite(func() {
	Expect(runtime.GOOS).To(Equal("linux"), "This test suite can only run in linux due to requiring connecting to a linux socket. Use `make fly` to run these test!")

	SetDefaultEventuallyTimeout(10 * time.Minute)
	nodeName = "default-smb-csi-driver-test-node"
	kubeConfigPath = "/tmp/csi-kubeconfig"

	local_k8s_cluster.CreateK8sCluster(nodeName, kubeConfigPath)

	local_k8s_cluster.Kubectl("apply", "--kustomize", "./base")
	Eventually("/tmp/csi.sock").Should(BeAnExistingFile())
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster(nodeName, kubeConfigPath)
})