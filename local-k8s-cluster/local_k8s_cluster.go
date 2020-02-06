package local_k8s_cluster

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"sigs.k8s.io/kind/pkg/apis/config/defaults"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func CreateK8sCluster(nodeName string, kubeConfigPath string) {
	Expect(os.Setenv("KUBECONFIG", kubeConfigPath)).To(Succeed())

	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
	)

	// Check if the cluster name already exists
	n, err := provider.ListNodes(nodeName)
	Expect(err).NotTo(HaveOccurred())
	Expect(n).To(HaveLen(0), "node(s) already exist for a cluster with the name "+nodeName)

	// create the cluster
	err = provider.Create(
		nodeName,
		cluster.CreateWithRawConfig([]byte(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches: 
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry:5000"]
    endpoint = ["http://registry:5000"]
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
  extraMounts:                                                
    - hostPath: /tmp/                                         
      containerPath: /var/lib/kubelet/plugins/csi-smbplugin   
    - hostPath: /tmp/                                         
      containerPath: /var/lib/kubelet/plugins_registry   
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
  - containerPort: 443
    hostPort: 443`)),
		cluster.CreateWithNodeImage(defaults.Image), // There's a v1.13 image = kindest/node:v1.13.12
		cluster.CreateWithRetain(true),
		cluster.CreateWithWaitForReady(10*time.Minute),
		cluster.CreateWithKubeconfigPath(kubeConfigPath),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
	Expect(err).NotTo(HaveOccurred())

	kubeContext := "kind-" + nodeName
	kubectl("--context", kubeContext, "--kubeconfig", kubeConfigPath)

	ngninxYamlTempFile, err := ioutil.TempFile("/tmp", "nginx")
	Expect(err).NotTo(HaveOccurred())

	_, err = io.WriteString(ngninxYamlTempFile, NGNIX_YAML)
	Expect(err).NotTo(HaveOccurred())
	Expect(ngninxYamlTempFile.Close()).To(Succeed())

	kubectl("apply", "-f", ngninxYamlTempFile.Name())

	kubectl("patch", "deployments", "-n", "ingress-nginx", "nginx-ingress-controller", "-p", `{"spec":{"template":{"spec":{"containers":[{"name":"nginx-ingress-controller","ports":[{"containerPort":80,"hostPort":80},{"containerPort":443,"hostPort":443}]}],"nodeSelector":{"ingress-ready":"true"},"tolerations":[{"key":"node-role.kubernetes.io/master","operator":"Equal","effect":"NoSchedule"}]}}}}`)

	registryBashTempFile, err := ioutil.TempFile("/tmp", "test")
	Expect(err).NotTo(HaveOccurred())
	Expect(os.Chmod(registryBashTempFile.Name(), os.ModePerm)).To(Succeed())

	_, err = io.WriteString(registryBashTempFile, SPIN_UP_LOCAL_REGISTRY_BASH)
	Expect(err).NotTo(HaveOccurred())
	Expect(registryBashTempFile.Close()).To(Succeed())

	runTestCommand("bash", "-c", registryBashTempFile.Name()+" "+nodeName+"-control-plane")
}

func DeleteK8sCluster(nodeName string, kubeConfigPath string) {
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
	)

	_ = provider.Delete(nodeName, kubeConfigPath)
}

func kubectl(cmd ...string) string {
	return runTestCommand("kubectl", cmd...)
}

func runTestCommand(name string, cmds ...string) string {
	command := exec.Command(name, cmds...)

	fmt.Println(fmt.Sprintf("Running %v", command.Args))

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, time.Minute).Should(gexec.Exit(0), string(session.Out.Contents()))
	return string(session.Out.Contents()) + string(session.Err.Contents())
}
