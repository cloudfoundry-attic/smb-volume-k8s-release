package main_test

import (
	. "code.cloudfoundry.org/smb-broker"
	"code.cloudfoundry.org/smb-broker/store/storefakes"
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
		var (
			err                    error
			instanceID, val1, key1 string
		)
		BeforeEach(func() {
			key1 = randomString()
			val1 = randomString()
			instanceID = randomString()
			request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/v2/service_instances/%s", instanceID), nil)
			Expect(err).NotTo(HaveOccurred())
			request.Header.Add("X-Broker-API-Version", "2.14")
		})

		BeforeEach(func() {
			store.GetReturns(map[string]interface{}{
				key1: val1,
			})
		})

		It("should allow provisioning a new service", func() {
			Expect(store.GetCallCount()).To(Equal(1))
			Expect(store.GetArgsForCall(0)).To(Equal(instanceID))
			Expect(recorder.Code).To(Equal(200))
			Expect(recorder.Body).To(MatchJSON(fmt.Sprintf(`{ "service_id": "", "plan_id": "", "parameters": { "%s": "%s" } }`, key1, val1)))
		})
	})
})

func randomString() string {
	sourceSeededByGinkgo := rand.NewSource(GinkgoRandomSeed())
	return strconv.Itoa(rand.New(sourceSeededByGinkgo).Int())
}
