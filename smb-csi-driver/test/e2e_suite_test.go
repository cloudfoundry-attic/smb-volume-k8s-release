package test_test

import (
	"k8s.io/kubernetes/test/e2e/framework"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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
