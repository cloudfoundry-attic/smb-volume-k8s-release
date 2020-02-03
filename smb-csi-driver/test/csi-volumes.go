package test

import (
	. "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/testpatterns"
	"k8s.io/kubernetes/test/e2e/storage/testsuites"
	"k8s.io/kubernetes/test/e2e/storage/utils"
)

var CSITestSuites = []func() testsuites.TestSuite{ }

// This executes testSuites for csi volumes.
var _ = utils.SIGDescribe("CSI Volumes", func() {
	curDriver := noopTestDriver{}
	Context(testsuites.GetDriverNameWithFeatureTags(curDriver), func() {
		testsuites.DefineTestSuite(curDriver, CSITestSuites)
	})

})

type noopTestDriver struct {}

func (noopTestDriver) GetDriverInfo() *testsuites.DriverInfo {
	panic("implement me")
}

func (noopTestDriver) SkipUnsupportedTest(testpatterns.TestPattern) {
	panic("implement me")
}

func (noopTestDriver) PrepareTest(f *framework.Framework) (*testsuites.PerTestConfig, func()) {
	panic("implement me")
}
