package local_k8s_cluster

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func CreateK8sCluster(nodeName string, kubeConfigPath string, k8sImage string) {
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
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
  - containerPort: 139
    hostPort: 139
  - containerPort: 445
    hostPort: 445
  - containerPort: 443
    hostPort: 443`)),
		cluster.CreateWithNodeImage(k8sImage),
		cluster.CreateWithRetain(true),
		cluster.CreateWithWaitForReady(10*time.Minute),
		cluster.CreateWithKubeconfigPath(kubeConfigPath),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
	Expect(err).NotTo(HaveOccurred())

	kubeContext := "kind-" + nodeName
	Kubectl("--context", kubeContext, "--kubeconfig", kubeConfigPath)

	ngninxYamlTempFile, err := ioutil.TempFile("/tmp", "nginx")
	Expect(err).NotTo(HaveOccurred())

	_, err = io.WriteString(ngninxYamlTempFile, NGNIX_YAML)
	Expect(err).NotTo(HaveOccurred())
	Expect(ngninxYamlTempFile.Close()).To(Succeed())

	Kubectl("apply", "-f", ngninxYamlTempFile.Name())

	Kubectl("patch", "deployments", "-n", "ingress-nginx", "nginx-ingress-controller", "-p", `{"spec":{"template":{"spec":{"containers":[{"name":"nginx-ingress-controller","ports":[{"containerPort":80,"hostPort":80},{"containerPort":443,"hostPort":443}]}],"nodeSelector":{"ingress-ready":"true"},"tolerations":[{"key":"node-role.kubernetes.io/master","operator":"Equal","effect":"NoSchedule"}]}}}}`)

	runLocalDockerRegistry(err, nodeName)
}

func runLocalDockerRegistry(err error, nodeName string) {
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

	finished := make(chan interface{}, 1)
	go func() {
		provider.Delete(nodeName, kubeConfigPath)
		close(finished)
	}()
	timeout := time.After(10 * time.Minute)

	select {
	case <-finished:
		return
	case <-timeout:
		fmt.Fprint(GinkgoWriter, "Unable to stop the kind cluster. Skipping...")
		return
	}

}

func KubectlStdOut(cmd ...string) string {
	stdout, _ := runTestCommand("kubectl", cmd...)
	return stdout
}

func HelmStdout(cmd ...string) string {
	stdout, _ := runTestCommand("helm", cmd...)
	return stdout
}

func Kubectl(cmd ...string) string {
	stdout, stderr := runTestCommand("kubectl", cmd...)
	return stdout + stderr
}

func KubectlWithStringAsStdIn(cmd ...string) func(contents string) string {
	return func(contents string) string {
		tempYamlFile, err := ioutil.TempFile(os.TempDir(), "temp_kapply_yaml")
		Expect(err).NotTo(HaveOccurred())

		_, err = tempYamlFile.WriteString(contents)
		Expect(err).NotTo(HaveOccurred())
		Expect(tempYamlFile.Close()).NotTo(HaveOccurred())

		cmd = append(cmd, tempYamlFile.Name())
		stdout, stderr := runTestCommand("kubectl", cmd...)

		return stdout + stderr
	}
}

func KappWithStringAsStdIn(cmd ...string) func(contents string) string {
	return func(contents string) string {
		tempYamlFile, err := ioutil.TempFile(os.TempDir(), "temp_kapply_yaml")
		Expect(err).NotTo(HaveOccurred())

		_, err = tempYamlFile.WriteString(contents)
		Expect(err).NotTo(HaveOccurred())
		Expect(tempYamlFile.Close()).NotTo(HaveOccurred())

		cmd = append(cmd, tempYamlFile.Name())
		stdout, stderr := runTestCommand("kapp", cmd...)

		return stdout + stderr
	}
}

func Helm(cmd ...string) string {
	stdout, stderr := runTestCommand("helm", cmd...)
	return stdout + stderr
}

func Docker(cmd ...string) string {
	stdout, stderr := runTestCommand("docker", cmd...)
	return stdout + stderr
}

func KbldStdout(args ...string) string {
	stdout, _ := runTestCommand("kbld", args...)
	return stdout
}

func runTestCommand(name string, cmds ...string) (string, string) {
	command := exec.Command(name, cmds...)

	fmt.Println(fmt.Sprintf("Running %v", command.Args))

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, time.Minute).Should(gexec.Exit(0), string(session.Out.Contents()))
	return string(session.Out.Contents()), string(session.Err.Contents())
}
