package integration_tests_test

import (
	local_k8s_cluster "code.cloudfoundry.org/smb-volume-k8s-local-cluster"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTests Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Minute)

	local_k8s_cluster.CreateK8sCluster("test", "/tmp/kubeconfig", "")
	local_k8s_cluster.CreateKpackImageResource()

	By("deploying the smb broker into the k8s cluster", func() {
		local_k8s_cluster.Kubectl("create", "namespace", "cf-workloads")
		smbBrokerUsername := "foo"
		smbBrokerPassword := "bar"
		namespace := "cf-workloads"

		smbBrokerDeploymentYaml := local_k8s_cluster.YttStdout("-f", "../smb-broker/ytt", "-v", "smbBrokerUsername="+smbBrokerUsername, "-v", "smbBrokerPassword="+smbBrokerPassword, "-v", "namespace="+namespace, "-v", "image.repository=registry:5000/cfpersi/smb-broker", "-v", "image.tag=local-test")
		local_k8s_cluster.KappWithStringAsStdIn("-y", "deploy", "-a", "smb-broker", "-f")(smbBrokerDeploymentYaml)
	})

	By("deploying smb csi driver into the k8s cluster", func() {
		kubectlStdOut := local_k8s_cluster.YttStdout("-f", "../smb-csi-driver/ytt/base", "-f", "../smb-csi-driver/ytt/test.yaml")
		local_k8s_cluster.KappWithStringAsStdIn("-y", "deploy", "-a", "smb-csi-driver", "-f")(kubectlStdOut)
		Eventually(func()string{
			return local_k8s_cluster.Kubectl("get", "pod", "-l", "app=csi-nodeplugin-smbplugin")
		}, 10 * time.Minute, 1 * time.Second).Should(ContainSubstring("Running"))
	})
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster("test", "/tmp/kubeconfig")
})
