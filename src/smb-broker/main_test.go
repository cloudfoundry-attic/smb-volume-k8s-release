package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"os/exec"
)

var _ = Describe("Main", func() {
	JustBeforeEach(func() {
		smbBrokerCmd := exec.Command(smbBrokerCompiledPath)
		session, err := gexec.Start(smbBrokerCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Started"))
	})

})
