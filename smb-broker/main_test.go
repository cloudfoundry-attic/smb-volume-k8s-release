package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"net/http"
	"os/exec"
	"time"
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
		var resp *http.Response

		Eventually(func() string {
			resp, _ = http.DefaultClient.Get("http://localhost:8080/v2/catalog")
			if resp == nil {
				return ""
			}
			return resp.Status
		}, 10 * time.Second).Should(Equal("200 OK"))

		assertHttpResponseContainsSubstring(resp.Body, "services")
	})
})