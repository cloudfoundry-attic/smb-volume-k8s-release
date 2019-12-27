package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"strings"
)

var _ = Describe("Main", func() {

	Describe("#Catalog", func() {
		It("should list catalog of services offered by the SMB service broker", func() {
			var resp *http.Response

			Eventually(func() string {
				resp, _ = http.DefaultClient.Get("http://localhost/v2/catalog")
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(Equal("200 OK"))

			assertHttpResponseContainsSubstring(resp.Body, "services")
		})
	})

	Describe("#Provision", func() {
		It("provision a new service", func() {
			var resp *http.Response

			Eventually(func() string {
				request, err := http.NewRequest("PUT", "http://localhost/v2/service_instances/1", strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id" }`))
				Expect(err).NotTo(HaveOccurred())

				resp, _ = http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(Equal("201 Created"))

			bytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bytes)).Should(ContainSubstring(`{}`))
		})
	})

	//XContext("Unable to start a http server", func() {
	//	var server *http.Server
	//	BeforeEach(func() {
	//		go func() {
	//			defer GinkgoRecover()
	//			server = &http.Server{Addr: "0.0.0.0", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//				w.WriteHeader(200)
	//			})}
	//			err := server.ListenAndServe()
	//			Expect(err).To(MatchError(http.ErrServerClosed))
	//		}()
	//
	//		Eventually(func() error {
	//			_, err := http.Get("http://localhost")
	//			return err
	//		}).Should(Succeed())
	//	})
	//
	//	AfterEach(func() {
	//		if server != nil {
	//			Expect(server.Close()).To(Succeed())
	//			time.Sleep(1 * time.Second)
	//		}
	//	})
	//
	//	It("should log a meaningful error", func() {
	//		Eventually(session, 10*time.Second).Should(gbytes.Say("Unable to start server"))
	//		Eventually(session).Should(gexec.Exit(1))
	//	})
	//})
})
