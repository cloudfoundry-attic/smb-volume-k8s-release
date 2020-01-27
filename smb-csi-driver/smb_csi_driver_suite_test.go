package main_test

import (
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)
var smbCsiDriverPath string
var session *gexec.Session

func TestSmbCsiDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = BeforeSuite(func() {
	var err error
	smbCsiDriverPath, err = gexec.Build("code.cloudfoundry.org/smb-csi-driver", "-mod", "vendor")
	Expect(err).NotTo(HaveOccurred())

	cmd := exec.Command(smbCsiDriverPath)
	session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(session.Kill().Wait().ExitCode()).NotTo(Equal(0))
})