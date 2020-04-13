package main_test

import (
	"code.cloudfoundry.org/smb-volume-k8s-local-cluster"
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
	var instanceID, serviceID, planID, basicAuth string
	var source = rand.NewSource(GinkgoRandomSeed())

	BeforeEach(func() {
		http.DefaultClient.Timeout = 30 * time.Second
		instanceID = randomString(source)
		basicAuth = smbBrokerUsername + ":" + smbBrokerPassword
		serviceID = "123"
		planID = "plan-id"
	})

	Describe("#Catalog", func() {
		It("should list catalog of services offered by the SMB service broker", func() {
			var resp *http.Response

			Eventually(func() string {
				resp, _ = http.DefaultClient.Get(fmt.Sprintf("http://%s@localhost/v2/catalog", basicAuth))
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
			local_k8s_cluster.Kubectl("-n", targetNamespace, "delete", "persistentvolumeclaims", instanceID)
			local_k8s_cluster.Kubectl("-n", targetNamespace, "delete", "persistentvolume", instanceID)
		})

		It("provision a new service", func() {
			var resp *http.Response

			Expect(local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "persistentvolumes")).NotTo(ContainSubstring(instanceID))
			Expect(local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "persistentvolumeclaims")).NotTo(ContainSubstring(instanceID))

			Eventually(func() string {
				request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s", basicAuth, instanceID),
					strings.NewReader(fmt.Sprintf(`{ "service_id": "%s", "plan_id": "%s", "parameters": { "username": "foo", "password": "bar", "share": "//unc.path/share" } }`, serviceID, planID)))
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

			Eventually(func() string {
				return local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "persistentvolume", instanceID)
			}).Should(ContainSubstring("Available"))
			Eventually(func() string {
				return local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "persistentvolumeclaim", instanceID)
			}).Should(ContainSubstring("Bound"))
			Expect(local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "secret", instanceID)).To(ContainSubstring(instanceID))
		})
	})

	Describe("#Deprovision", func() {
		BeforeEach(func() {
			Eventually(func() string {
				request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s", basicAuth, instanceID), strings.NewReader(fmt.Sprintf(`{ "service_id": "%s", "plan_id": "%s", "parameters": { "username": "foo", "password": "bar", "share": "//unc.path/share" } }`, serviceID, planID)))
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

			Expect(local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "persistentvolume", instanceID)).To(Or(ContainSubstring("Available"), ContainSubstring("Bound")))
			Expect(local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "persistentvolumeclaim", instanceID)).To(ContainSubstring("Pending"))
			Expect(local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "secret", instanceID)).To(ContainSubstring(instanceID))

			Eventually(func() string {
				request, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s?service_id=%s&plan_id=%s", basicAuth, instanceID, serviceID, planID), nil)
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

			Eventually(func() string {
				return local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "persistentvolumes")
			}).ShouldNot(ContainSubstring(instanceID))

			Eventually(func() string {
				return local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "persistentvolumeclaims")
			}).ShouldNot(ContainSubstring(instanceID))

			Eventually(func() string {
				return local_k8s_cluster.Kubectl("-n", targetNamespace, "get", "secrets")
			}).Should(Not(ContainSubstring(instanceID)))
		})
	})

	Describe("#Bind", func() {
		var resp *http.Response
		var bindingID string

		AfterEach(func() {
			local_k8s_cluster.Kubectl("-n", targetNamespace, "delete", "persistentvolumeclaims", instanceID)
			local_k8s_cluster.Kubectl("-n", targetNamespace, "delete", "persistentvolume", instanceID)
		})

		BeforeEach(func() {
			bindingID = randomString(source)

			Eventually(func() string {
				request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s", basicAuth, instanceID),
					strings.NewReader(fmt.Sprintf(`{ "service_id": "%s", "plan_id": "%s", "parameters": { "username": "foo", "password": "bar", "share": "//unc.path/share" } }`, serviceID, planID)))
				Expect(err).NotTo(HaveOccurred())

				resp, _ = http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(Equal("201 Created"))
		})

		It("returns 200", func() {
			Eventually(func() string {
				request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s/service_bindings/%s", basicAuth, instanceID, bindingID),
					strings.NewReader(fmt.Sprintf(`{"service_id": "%s", "plan_id": "plan_id", "bind_resource": {"app_guid": "123"}}`, serviceID)))

				Expect(err).NotTo(HaveOccurred())
				resp, _ = http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(Equal("201 Created"))

			bytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bytes)).To(MatchJSON(fmt.Sprintf(`{"credentials": {}, "volume_mounts": [{"driver": "smb", "container_dir": "/home/vcap/data/", "mode": "rw", "device_type": "shared", "device": {"volume_id": "%s", "mount_config": {"name": "%s"}} }]}`, bindingID, instanceID)))
		})
	})

	Describe("#Unbind", func() {
		var resp *http.Response
		var bindingID string

		AfterEach(func() {
			local_k8s_cluster.Kubectl("-n", targetNamespace, "delete", "persistentvolumeclaims", instanceID)
			local_k8s_cluster.Kubectl("-n", targetNamespace, "delete", "persistentvolume", instanceID)
		})

		BeforeEach(func() {
			bindingID = randomString(source)

			Eventually(func() string {
				request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s", basicAuth, instanceID),
					strings.NewReader(fmt.Sprintf(`{ "service_id": "%s", "plan_id": "%s", "parameters": { "username": "foo", "password": "bar", "share": "//unc.path/share" } }`, serviceID, planID)))
				Expect(err).NotTo(HaveOccurred())

				resp, _ = http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(Equal("201 Created"))

			Eventually(func() string {
				request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s/service_bindings/%s", basicAuth, instanceID, bindingID),
					strings.NewReader(fmt.Sprintf(`{"service_id": "%s", "plan_id": "plan_id", "bind_resource": {"app_guid": "123"}}`, serviceID)))

				Expect(err).NotTo(HaveOccurred())
				resp, _ = http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(Equal("201 Created"))
		})

		It("returns 200", func() {
			Eventually(func() string {
				request, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s/service_bindings/%s?service_id=%s&plan_id=plan_id", basicAuth, instanceID, bindingID, serviceID), nil)
				Expect(err).NotTo(HaveOccurred())

				resp, _ = http.DefaultClient.Do(request)
				if resp == nil {
					return ""
				}
				return resp.Status
			}).Should(ContainSubstring("200"))
		})
	})

	Describe("#GetInstance", func() {
		var resp *http.Response

		Context("when a service instance has been provisioned", func() {

			AfterEach(func() {
				local_k8s_cluster.Kubectl("-n", targetNamespace, "delete", "persistentvolumeclaims", instanceID)
				local_k8s_cluster.Kubectl("-n", targetNamespace, "delete", "persistentvolume", instanceID)
			})

			BeforeEach(func() {
				By("provisioning a service", func() {
					Eventually(func() string {
						request, err := http.NewRequest("PUT",
							fmt.Sprintf("http://%s@localhost/v2/service_instances/%s", basicAuth, instanceID),
							strings.NewReader(fmt.Sprintf(
								`{ "service_id": "%s", "plan_id": "%s", "parameters": { "share": "//unc.path/share", "username": "username_value", "password": "password_value"} }`, serviceID, planID)))
						Expect(err).NotTo(HaveOccurred())

						resp, _ = http.DefaultClient.Do(request)
						if resp == nil {
							return ""
						}
						return resp.Status
					}).Should(Equal("201 Created"))
				})
			})

			It("returns 200", func() {
				Eventually(func() string {
					request, err := http.NewRequest("GET", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s", basicAuth, instanceID), nil)
					Expect(err).NotTo(HaveOccurred())

					request.Header = map[string][]string{
						"X-Broker-API-Version": {"2.14"},
					}
					resp, _ = http.DefaultClient.Do(request)
					if resp == nil {
						return ""
					}
					return resp.Status
				}).Should(Equal("200 OK"))

				bytes, err := ioutil.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bytes)).To(MatchJSON(fmt.Sprintf(`{"service_id": "%s", "plan_id": "%s", "parameters": { "share": "//unc.path/share", "username": "username_value" } }`, serviceID, planID)))
			})

		})

		Context("when attempting to retreive a service instance that hasn't been provisioned", func() {
			It("returns 404", func() {
				Eventually(func() string {
					request, err := http.NewRequest("GET", fmt.Sprintf("http://%s@localhost/v2/service_instances/%s", basicAuth, instanceID), nil)
					Expect(err).NotTo(HaveOccurred())
					request.Header = map[string][]string{
						"X-Broker-API-Version": {"2.14"},
					}
					resp, _ = http.DefaultClient.Do(request)

					if resp == nil {
						return ""
					}
					return resp.Status
				}).Should(Equal("404 Not Found"))

				bytes, err := ioutil.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bytes)).To(MatchJSON(`{"description": "unable to find service instance"}`))
			})

		})
	})

})

func randomString(sourceSeededByGinkgo rand.Source) string {
	return strconv.Itoa(rand.New(sourceSeededByGinkgo).Int())
}
