package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"net/http"
	"os/exec"
)

var _ = Describe("Main", func() {
	var session *gexec.Session

	JustBeforeEach(func() {
		smbBrokerCmd := exec.Command(smbBrokerCompiledPath)
		var err error
		session, err = gexec.Start(smbBrokerCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Started"))
	})

	AfterEach(func() {
		session.Kill()
	})

	It("should list catalog of services offered by the SMB service broker", func() {
		resp, err := http.DefaultClient.Get("http://localhost:8080/v2/catalog")
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.Status).To(Equal("200 OK"))
		assertHttpResponseContainsSubstring(resp.Body, "services")
	})
})