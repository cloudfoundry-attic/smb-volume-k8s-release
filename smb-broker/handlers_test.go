package main_test

import (
	"code.cloudfoundry.org/smb-broker/store/storefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"

	. "code.cloudfoundry.org/smb-broker"
)

var _ = Describe("Handlers", func() {
	var recorder *httptest.ResponseRecorder
	var request *http.Request
	var store *storefakes.FakeServiceInstanceStore

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
		store = &storefakes.FakeServiceInstanceStore{}
	})

	JustBeforeEach(func() {
		BrokerHandler(store).ServeHTTP(recorder, request)
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
		BeforeEach(func() {
			var err error
			request, err = http.NewRequest(http.MethodPut, "/v2/service_instances/123", strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id" }`))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should allow provisioning a new service", func() {
			Expect(recorder.Code).To(Equal(201))
			Expect(recorder.Body).To(MatchJSON(`{}`))
		})
	})

	Describe("#GetInstance endpoint", func() {
		BeforeEach(func() {
			var err error
			request, err = http.NewRequest(http.MethodGet, "/v2/service_instances/123", nil)
			Expect(err).NotTo(HaveOccurred())
			request.Header.Add("X-Broker-API-Version","2.14")
		})


		BeforeEach(func() {
			store.GetReturns(map[string]interface{} {
				"key1": "val1",
			})
		})

		It("should allow provisioning a new service", func() {
			Expect(recorder.Code).To(Equal(200))
			Expect(recorder.Body).To(MatchJSON(`{ "service_id": "", "plan_id": "", "parameters": { "key1": "val1" } }`))
		})
	})
})
