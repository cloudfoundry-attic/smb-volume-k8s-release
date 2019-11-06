package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"

	. "code.cloudfoundry.org/smb-broker"
)

var _ = Describe("Handlers", func() {

	Describe("catalog endpoint", func() {
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
})
