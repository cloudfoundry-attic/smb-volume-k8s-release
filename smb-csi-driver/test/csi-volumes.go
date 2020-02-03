package test

import (
	. "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/testpatterns"
	"k8s.io/kubernetes/test/e2e/storage/testsuites"
	"k8s.io/kubernetes/test/e2e/storage/utils"
)

var CSITestSuites = []func() testsuites.TestSuite{
	testsuites.InitVolumesTestSuite,
}

// This executes testSuites for csi volumes.
var _ = utils.SIGDescribe("CSI Volumes", func() {
	curDriver := noopTestDriver{}
	Context(testsuites.GetDriverNameWithFeatureTags(curDriver), func() {
		testsuites.DefineTestSuite(curDriver, CSITestSuites)
	})

})

type noopTestDriver struct {}

func (noopTestDriver) GetDriverInfo() *testsuites.DriverInfo {
	return &testsuites.DriverInfo{}
}

func (noopTestDriver) SkipUnsupportedTest(testpatterns.TestPattern) {
}

func (noopTestDriver) PrepareTest(f *framework.Framework) (*testsuites.PerTestConfig, func()) {
	return nil, nil
}
