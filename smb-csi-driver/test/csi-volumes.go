package test

import (
	. "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
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
var _ testsuites.TestVolume = &noopTestDriver{}
var _ testsuites.PreprovisionedVolumeTestDriver = &noopTestDriver{}
var _ testsuites.PreprovisionedPVTestDriver = &noopTestDriver{}

func (d noopTestDriver) GetPersistentVolumeSource(readOnly bool, fsType string, testVolume testsuites.TestVolume) (*v1.PersistentVolumeSource, *v1.VolumeNodeAffinity) {
	return &v1.PersistentVolumeSource{
		CSI: &v1.CSIPersistentVolumeSource{
			Driver: "org.cloudfoundry.smb",
			VolumeHandle: "volume-handle",
		},
	}, nil
}

func (noopTestDriver) GetDriverInfo() *testsuites.DriverInfo {
	return &testsuites.DriverInfo{
		Name: "org.cloudfoundry.smb",
		SupportedFsType: sets.NewString("ext4"),
	}
}

func (noopTestDriver) SkipUnsupportedTest(testpatterns.TestPattern) {
}


func (n noopTestDriver) PrepareTest(f *framework.Framework) (*testsuites.PerTestConfig, func()) {
	return &testsuites.PerTestConfig{
		Driver:    n,
		Prefix:    "smb",
		Framework: f,

	}, nil
}

func (d noopTestDriver) CreateVolume(config *testsuites.PerTestConfig, volType testpatterns.TestVolType) testsuites.TestVolume {
	return d
}

func (noopTestDriver) DeleteVolume() {
}