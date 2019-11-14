package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"

	. "code.cloudfoundry.org/smb-broker"
)

var _ = Describe("Handlers", func() {

	Describe("#Catalog endpoint", func() {
		makeCatalogRequest := func() *httptest.ResponseRecorder {
			recorder := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/v2/catalog", nil)

			BrokerHandler().ServeHTTP(recorder, request)
			return recorder
		}

		It("should list catalog of services offered by the SMB service broker", func() {
			response := makeCatalogRequest()
			Expect(response.Code).To(Equal(200))
			Expect(response.Body).To(MatchJSON(fixture("catalog.json")))
		})
	})

	Describe("#Provision endpoint", func() {
		makeProvisionRequest := func() *httptest.ResponseRecorder {
			recorder := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodPut, "/v2/service_instances/123", strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id" }`))

			BrokerHandler().ServeHTTP(recorder, request)
			return recorder
		}

		It("should allow provisioning a new service", func() {
			response := makeProvisionRequest()
			Expect(response.Code).To(Equal(201))
			Expect(response.Body).To(MatchJSON(`{}`))
		})
	})
})
