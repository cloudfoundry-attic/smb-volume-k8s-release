package main_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Main", func() {
	var instanceID string
	var source = rand.NewSource(GinkgoRandomSeed())

	BeforeEach(func() {
		http.DefaultClient.Timeout = 30 * time.Second
		instanceID = randomString(source)
	})

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
		AfterEach(func() {
			kubectl("delete", "persistentvolume", instanceID)
			kubectl("delete", "persistentvolumeclaims", instanceID)
		})

		It("provision a new service", func() {
			var resp *http.Response

			Expect(kubectl("get", "persistentvolumes")).To(ContainSubstring("No resources found"))
			Expect(kubectl("get", "persistentvolumeclaims")).To(ContainSubstring("No resources found"))

			Eventually(func() string {
				request, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost/v2/service_instances/%s", instanceID), strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id" }`))
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

			Expect(kubectl("get", "persistentvolume", instanceID)).To(ContainSubstring("Available"))
			Expect(kubectl("get", "persistentvolumeclaim", instanceID)).To(ContainSubstring("Pending"))
		})
	})

	Describe("#Deprovision", func() {
		BeforeEach(func() {
			Eventually(func() string {
				request, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost/v2/service_instances/%s", instanceID), strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id" }`))
				Expect(err).NotTo(HaveOccurred())

				resp, _ := http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(Equal("201 Created"))

		})

		It("deprovision a new service", func() {
			var resp *http.Response

			Expect(kubectl("get", "persistentvolume", instanceID)).To(ContainSubstring("Available"))
			Expect(kubectl("get", "persistentvolumeclaim", instanceID)).To(ContainSubstring("Pending"))

			Eventually(func() string {
				request, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost/v2/service_instances/%s?service_id=123&plan_id=plan-id", instanceID), nil)
				Expect(err).NotTo(HaveOccurred())

				resp, _ = http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(ContainSubstring("200"))

			bytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bytes)).Should(ContainSubstring(`{}`))

			Expect(kubectl("get", "persistentvolumes")).To(ContainSubstring("No resources found"))
			Expect(kubectl("get", "persistentvolumeclaims")).To(ContainSubstring("No resources found"))
		})
	})
})

func randomString(sourceSeededByGinkgo rand.Source) string {
	return strconv.Itoa(rand.New(sourceSeededByGinkgo).Int())
}