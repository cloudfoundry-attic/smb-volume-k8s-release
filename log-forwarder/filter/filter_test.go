package filter_test

import (
	. "code.cloudfoundry.org/volume-services-log-forwarder/filter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Filter", func() {

	var (
		filter  Filter
		matches bool
		event   *v1.Event
	)

	Describe("#Matches", func() {

		BeforeEach(func() {
			event = &v1.Event{}
		})

		JustBeforeEach(func() {
			filter = NewFilter()
			matches = filter.Matches(event)
		})

		Context("given an event that is not MountFailed", func() {

			It("should return false", func() {
				Expect(matches).To(BeFalse())
			})
		})

		Context("given an event that is MountFailed", func() {

			BeforeEach(func() {
				event.Reason = "MountFailed"
			})

			It("should return true", func() {
				Expect(matches).To(BeTrue())
			})
		})
	})
})