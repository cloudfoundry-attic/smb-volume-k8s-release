package main_test

import (
	"github.com/onsi/gomega/gexec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSmbBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmbBroker Suite")
}

var smbBrokerCompiledPath string

var _ = BeforeSuite(func() {
	var err error
	smbBrokerCompiledPath, err = gexec.Build("code.cloudfoundry.org/smb-broker")
	Expect(err).NotTo(HaveOccurred())
})
