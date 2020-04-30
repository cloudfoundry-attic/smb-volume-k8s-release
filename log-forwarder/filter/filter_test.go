package filter_test

import (
	. "code.cloudfoundry.org/volume-services-log-forwarder/filter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./filter_fakes/fake_filter.go . Filter

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

		Context("given an event that is not FailedMount", func() {

			It("should return false", func() {
				Expect(matches).To(BeFalse())
			})
		})

		Context("given an event that is FailedMount", func() {

			BeforeEach(func() {
				event.Reason = "FailedMount"
			})

			It("should return true", func() {
				Expect(matches).To(BeTrue())
			})
		})
	})
})
