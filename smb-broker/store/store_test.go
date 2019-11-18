package store_test

import (
	"code.cloudfoundry.org/smb-broker/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var serviceInstanceStore store.ServiceInstanceStore
	Describe("InMemoryServiceInstanceStore", func() {
		BeforeEach(func() {
			serviceInstanceStore = &store.InMemoryServiceInstanceStore{}
		})

		It("should return empty when retrieving from an empty store", func() {
			serviceInstance := serviceInstanceStore.Get("")
			Expect(serviceInstance).To(Equal(store.ServiceInstance{}))
		})

		It("should succesfully add a service instance into the store", func() {
			err := serviceInstanceStore.Add("key", store.ServiceInstance{})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a store has entries populated", func() {
			var key string
			var expectedServiceInstance store.ServiceInstance

			BeforeEach(func() {
				key = "key"
				expectedServiceInstance = store.ServiceInstance{
					ServiceID: "service-id",
				}

				Expect(serviceInstanceStore.Add(key, expectedServiceInstance)).To(Succeed())
			})

			It("Should be able to retrieve the record in the store", func() {
				fetchedServiceInstance := serviceInstanceStore.Get(key)
				Expect(fetchedServiceInstance).To(Equal(expectedServiceInstance))
			})
		})

	})
})
