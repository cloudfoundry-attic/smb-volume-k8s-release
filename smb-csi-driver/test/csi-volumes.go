package test

import (
	. "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/volume"
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
	vol, _ := testVolume.(*smbVolume)
	return &v1.PersistentVolumeSource{
		CSI: &v1.CSIPersistentVolumeSource{
			Driver: "org.cloudfoundry.smb",
			VolumeHandle: "volume-handle",
			VolumeAttributes: map[string]string{
				"server":   vol.serverIP,
				"share":    "/example1",
				"readOnly": "true",
			},
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
	f := config.Framework
	cs := f.ClientSet
	ns := f.Namespace

	serverConfig := volume.TestConfig{
		Namespace: ns.Name,
		Prefix:   "smb",
		ServerImage: "dperson/samba",
		ServerPorts: []int{139,445},
		ServerArgs: []string{"-p", "-u", "example1;badpass", "-s", "example1;/example1;no;no;no;example1", "-p", "-S"},
		ServerVolumes: map[string]string{},
		ServerReadyMessage: "finished starting up",
	}

	serverPod, serverIP := volume.CreateStorageServer(cs, serverConfig)
	return &smbVolume{
		serverIP: serverIP,
		serverPod: serverPod,
		f: f,
	}
}

type smbVolume struct {
	serverIP  string
	serverPod *v1.Pod
	f         *framework.Framework
}
func (smbVolume) DeleteVolume(){

}

func (noopTestDriver) DeleteVolume() {
}