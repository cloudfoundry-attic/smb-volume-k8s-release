package main_test

import (
	"fmt"
	"github.com/onsi/gomega/gexec"
	"io"
	"io/ioutil"
	"os/exec"
	"path"
	"sigs.k8s.io/kind/pkg/apis/config/defaults"
	"sigs.k8s.io/kind/pkg/cmd"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/kind/pkg/cluster"
)

func TestSmbBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmbBroker Suite")
}

var smbBrokerCompiledPath string
var kubeConfigPath string
var nodeName string

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(1 * time.Minute)
	createK8sCluster()
})

var _ = AfterSuite(func() {
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
	)

	_ = provider.Delete(nodeName, kubeConfigPath)
})

func createK8sCluster() {
	nodeName = "default-smb-broker-test-node"
	kubeConfigPath = "/tmp/kubeconfig"

	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
	)

	// Check if the cluster name already exists
	n, err := provider.ListNodes(nodeName)
	Expect(err).NotTo(HaveOccurred())
	Expect(n).To(HaveLen(0), "node(s) already exist for a cluster with the name "+nodeName)
	//kindest/node:v1.13.12
	// create the cluster
	err = provider.Create(
		nodeName,
		cluster.CreateWithRawConfig([]byte(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    apiVersion: kubeadm.k8s.io/v1beta2
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
        authorization-mode: "AlwaysAllow"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
  - containerPort: 443
    hostPort: 443`)),
		cluster.CreateWithNodeImage(defaults.Image),
		cluster.CreateWithRetain(true),
		cluster.CreateWithWaitForReady(10*time.Minute),
		cluster.CreateWithKubeconfigPath(kubeConfigPath),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
	Expect(err).NotTo(HaveOccurred())

	kubeContext := "kind-" + nodeName
	kubectl("cluster-info", "--context", kubeContext, "--kubeconfig", kubeConfigPath)
	kubectl("apply", "-f", "./assets/ingress-nginx")
	kubectl("patch", "deployments", "-n", "ingress-nginx", "nginx-ingress-controller", "-p", `{"spec":{"template":{"spec":{"containers":[{"name":"nginx-ingress-controller","ports":[{"containerPort":80,"hostPort":80},{"containerPort":443,"hostPort":443}]}],"nodeSelector":{"ingress-ready":"true"},"tolerations":[{"key":"node-role.kubernetes.io/master","operator":"Equal","effect":"NoSchedule"}]}}}}`)
	helm("--kubeconfig", kubeConfigPath, "--kube-context", kubeContext, "install", "smb-broker", "./helm", "--set", "ingress.enabled=true")
}

func helm(cmd ...string) {
	command := exec.Command("helm", cmd...)
	command.Env = append(command.Env, "KUBECONFIG="+kubeConfigPath)
	fmt.Println(fmt.Sprintf("Running %v", command.Args))

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}

func kubectl(cmd ...string) {
	command := exec.Command("kubectl", cmd...)
	command.Env = append(command.Env, "KUBECONFIG="+kubeConfigPath)
	fmt.Println(fmt.Sprintf("Running %v", command.Args))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}

func fixture(name string) string {
	filePath := path.Join("fixtures", name)
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("Could not read fixture: %s", name))
	}

	return string(contents)
}

func assertHttpResponseContainsSubstring(body io.Reader, expected string) {
	bytes, err := ioutil.ReadAll(body)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(bytes)).Should(ContainSubstring(expected))
}
