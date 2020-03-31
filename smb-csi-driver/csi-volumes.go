package main

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/volume"
	"k8s.io/kubernetes/test/e2e/storage/testpatterns"
	"k8s.io/kubernetes/test/e2e/storage/testsuites"
	"k8s.io/kubernetes/test/e2e/storage/utils"
	local_k8s_cluster "code.cloudfoundry.org/smb-volume-k8s-local-cluster"
	"os"
)

var CSITestSuites = []func() testsuites.TestSuite{
	testsuites.InitVolumesTestSuite,
	testsuites.InitVolumeIOTestSuite,
}

// This executes testSuites for csi volumes.
var _ = utils.SIGDescribe("CSI Volumes", func() {
	curDriver := noopTestDriver{}
	Context(testsuites.GetDriverNameWithFeatureTags(curDriver), func() {

		if val, found := os.LookupEnv("VOLUME_MODE_SUITE"); found && val == "true" {
			CSITestSuites = append(CSITestSuites, testsuites.InitVolumeModeTestSuite)
		}

		testsuites.DefineTestSuite(curDriver, CSITestSuites)
	})
})

type noopTestDriver struct{}

type smbVolume struct {
	serverIP  string
	serverPod *v1.Pod
	username  string
	password  string
	namespace string
	framework *framework.Framework
	config    volume.TestConfig
}

var _ testsuites.TestDriver = &noopTestDriver{}
var _ testsuites.PreprovisionedVolumeTestDriver = &noopTestDriver{}
var _ testsuites.PreprovisionedPVTestDriver = &noopTestDriver{}

func (n noopTestDriver) GetPersistentVolumeSource(readOnly bool, fsType string, testVolume testsuites.TestVolume) (*v1.PersistentVolumeSource, *v1.VolumeNodeAffinity) {
	vol, _ := testVolume.(*smbVolume)

	share := fmt.Sprintf("//%s/example1", vol.serverIP)

	local_k8s_cluster.Kubectl(
		"create",
		"secret",
		"generic",
		"secretref",
		fmt.Sprintf("--from-literal=username=%s", vol.username),
		fmt.Sprintf("--from-literal=password=%s", vol.password),
		"-n", vol.namespace,
		)

	return &v1.PersistentVolumeSource{
		CSI: &v1.CSIPersistentVolumeSource{
			Driver:       "org.cloudfoundry.smb",
			VolumeHandle: "volume-handle",
			VolumeAttributes: map[string]string{
				"username": vol.username,
				"password": vol.password,
				"share":    share,
				"readOnly": "true",
			},
			NodePublishSecretRef: &v1.SecretReference{
				Name: "secretref",
				Namespace: vol.namespace,
			},
		},
	}, nil
}

func (n noopTestDriver) GetDriverInfo() *testsuites.DriverInfo {
	return &testsuites.DriverInfo{
		Name:            "org.cloudfoundry.smb",
		SupportedFsType: sets.NewString(""),
		Capabilities: map[testsuites.Capability]bool{
			testsuites.CapPersistence: true,
			testsuites.CapExec: true,
		},
	}
}

func (n noopTestDriver) SkipUnsupportedTest(pattern testpatterns.TestPattern) {
	if pattern.VolType == testpatterns.DynamicPV {
		framework.Skipf("SMB Driver does not support dynamic provisioning -- skipping")
	}
}

func (n noopTestDriver) PrepareTest(f *framework.Framework) (*testsuites.PerTestConfig, func()) {
	return &testsuites.PerTestConfig{
		Driver:    n,
		Prefix:    "smb",
		Framework: f,
	}, nil
}

func (n noopTestDriver) CreateVolume(config *testsuites.PerTestConfig, volType testpatterns.TestVolType) testsuites.TestVolume {
	f := config.Framework
	cs := f.ClientSet
	ns := f.Namespace

	serverConfig := volume.TestConfig{
		Namespace:          ns.Name,
		Prefix:             "smb",
		ServerImage:        "dperson/samba",
		ServerPorts:        []int{139, 445},
		ServerArgs:         []string{"-p", "-u", "example1;badpass", "-s", "example1;/example1;no;no;no;example1", "-p", "-S"},
		ServerVolumes:      map[string]string{},
		ServerReadyMessage: "finished starting up",
	}

	serverPod, serverIP := volume.CreateStorageServer(cs, serverConfig)

	return &smbVolume{
		serverIP:  serverIP,
		serverPod: serverPod,
		username: "example1",
		password: "badpass",
		namespace: ns.Name,
		framework: f,
		config:    serverConfig,
	}
}

func (v *smbVolume) DeleteVolume() {
	volume.TestCleanup(v.framework, v.config)
}
