package main_test

import (
	. "code.cloudfoundry.org/smb-broker"
	"code.cloudfoundry.org/smb-broker/store"
	"code.cloudfoundry.org/smb-broker/store/storefakes"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

var _ = Describe("Handlers", func() {
	var brokerHandler http.Handler
	var err error
	var recorder *httptest.ResponseRecorder
	var request *http.Request
	var fakeServiceInstanceStore store.ServiceInstanceStore

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
		fakeServiceInstanceStore = &storefakes.FakeServiceInstanceStore{}
	})

	JustBeforeEach(func() {
		brokerHandler, err = BrokerHandler(fakeServiceInstanceStore)
	})

	Describe("Validation", func() {
		Context("When missing a store", func() {
			BeforeEach(func() {
				fakeServiceInstanceStore = nil
			})

			It("should return a meaningful error message", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("missing a Service Instance Store"))
			})
		})
	})

	Describe("Endpoints", func() {
		JustBeforeEach(func() {
			brokerHandler.ServeHTTP(recorder, request)
		})

		Describe("#Catalog endpoint", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest(http.MethodGet, "/v2/catalog", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should list catalog of services offered by the SMB service broker", func() {
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body).To(MatchJSON(fixture("catalog.json")))
			})
		})

		Describe("#Provision endpoint", func() {
			var serviceInstanceKey string
			BeforeEach(func() {
				serviceInstanceKey = "123"

				var err error
				request, err = http.NewRequest(http.MethodPut, "/v2/service_instances/"+serviceInstanceKey, strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id", "parameters": { "parameter1": "1", "parameter2": "foo" } }`))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow provisioning and store the new service instance", func() {
				Expect(recorder.Code).To(Equal(201))
				Expect(recorder.Body).To(MatchJSON(`{}`))

				Expect(fakeServiceInstanceStore.(*storefakes.FakeServiceInstanceStore).AddCallCount()).To(Equal(1))
				key, serviceInstance := fakeServiceInstanceStore.(*storefakes.FakeServiceInstanceStore).AddArgsForCall(0)
				Expect(key).To(Equal(serviceInstanceKey))
				Expect(serviceInstance.Parameters).To(HaveKeyWithValue("parameter1", "1"))
				Expect(serviceInstance.Parameters).To(HaveKeyWithValue("parameter2", "foo"))
			})


			Context("when unable to store a service instance", func() {
				BeforeEach(func() {
					fakeServiceInstanceStore.(*storefakes.FakeServiceInstanceStore).AddReturns(errors.New("unable to store"))
				})

				It("should return a meaningful error", func() {
					Expect(recorder.Code).To(Equal(500))
				})
			})

			Context("when service instance parameters are not provided", func() {
				BeforeEach(func() {
					request, err = http.NewRequest(http.MethodPut, "/v2/service_instances/"+serviceInstanceKey, strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id" }`))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should allow provisioning and store the new service instance", func() {
					Expect(recorder.Code).To(Equal(201))
					Expect(recorder.Body).To(MatchJSON(`{}`))

				})
			})
		})

		Describe("#GetInstance endpoint", func() {
			var (
				err                                                   error
				instanceID, val1, val2, key1, key2, serviceID, planID string
			)
			var source = rand.NewSource(GinkgoRandomSeed())

			BeforeEach(func() {
				instanceID = randomString(source)
				request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/service_instances/%s", instanceID), nil)
				Expect(err).NotTo(HaveOccurred())
				request.Header.Add("X-Broker-API-Version", "2.14")
			})

			BeforeEach(func() {
				key1 = randomString(source)
				key2 = randomString(source)
				val1 = randomString(source)
				val2 = randomString(source)
				serviceID = randomString(source)
				planID = randomString(source)

				params := map[string]interface{}{
					key1: val1,
					key2: val2,
				}
				fakeServiceInstanceStore.(*storefakes.FakeServiceInstanceStore).GetReturns(store.ServiceInstance{
					ServiceID:  serviceID,
					PlanID:     planID,
					Parameters: params,
				}, true)
			})

			It("should retrieve a service instance from the store", func() {
				Expect(fakeServiceInstanceStore.(*storefakes.FakeServiceInstanceStore).GetCallCount()).To(Equal(1))
				Expect(fakeServiceInstanceStore.(*storefakes.FakeServiceInstanceStore).GetArgsForCall(0)).To(Equal(instanceID))
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body).To(MatchJSON(
					fmt.Sprintf(`{ "service_id": "%s", "plan_id": "%s", "parameters": { "%s": "%s", "%s": "%s" } }`, serviceID, planID, key1, val1, key2, val2)),
				)
			})

			Context("when no service instance exists in the store", func() {
				BeforeEach(func() {
					fakeServiceInstanceStore.(*storefakes.FakeServiceInstanceStore).GetReturns(store.ServiceInstance{}, false)
				})
				It("Should return an FailureError with a 404 status code", func() {
					Expect(recorder.Code).To(Equal(404))
					Expect(recorder.Body).To(MatchJSON(`{"description": "unable to find service instance"}`),
					)
				})

			})
		})
	})
})

func randomString(sourceSeededByGinkgo rand.Source) string {
	return strconv.Itoa(rand.New(sourceSeededByGinkgo).Int())
}
