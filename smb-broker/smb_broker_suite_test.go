package main_test

import (
	"fmt"
	"github.com/onsi/gomega/gexec"
	"io"
	"io/ioutil"
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

var _ = BeforeSuite(func() {
	createK8sCluster()

	var err error
	SetDefaultEventuallyTimeout(1 * time.Minute)
	smbBrokerCompiledPath, err = gexec.Build("code.cloudfoundry.org/smb-broker", "-mod=vendor")
	Expect(err).NotTo(HaveOccurred())
})

func createK8sCluster() {
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
	)
	// Check if the cluster name already exists
	nodeName := "default-smb-broker-test-node"
	n, err := provider.ListNodes(nodeName)
	Expect(err).NotTo(HaveOccurred())
	Expect(n).To(HaveLen(0), "node(s) already exist for a cluster with the name "+nodeName)
	//kindest/node:v1.13.12
	// create the cluster
	kubbeConfigPath := "/tmp/kubeconfig"
	err = provider.Create(
		nodeName,
		cluster.CreateWithNodeImage(defaults.Image),
		cluster.CreateWithRetain(true),
		cluster.CreateWithWaitForReady(10*time.Minute),
		cluster.CreateWithKubeconfigPath(kubbeConfigPath),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
	Expect(err).NotTo(HaveOccurred())
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
