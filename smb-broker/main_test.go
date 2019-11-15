package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
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

	Describe("#Catalog", func() {
		It("should list catalog of services offered by the SMB service broker", func() {
			var resp *http.Response

			Eventually(func() string {
				resp, _ = http.DefaultClient.Get("http://localhost:8080/v2/catalog")
				if resp == nil {
					return ""
				}
				return resp.Status
			}, 10*time.Second).Should(Equal("200 OK"))

			assertHttpResponseContainsSubstring(resp.Body, "services")
		})
	})

	Describe("#Provision", func() {
		It("provision a new service", func() {
			var resp *http.Response

			Eventually(func() string {
				request, err := http.NewRequest("PUT", "http://localhost:8080/v2/service_instances/1", strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id" }`))
				Expect(err).NotTo(HaveOccurred())

				resp, _ = http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}, 10*time.Second).Should(Equal("201 Created"))

			bytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bytes)).Should(ContainSubstring(`{}`))
		})
	})
})
