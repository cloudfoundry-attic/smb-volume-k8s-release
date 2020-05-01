package filter_test

import (
	"code.cloudfoundry.org/lager/lagertest"
	. "code.cloudfoundry.org/volume-services-log-forwarder/filter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./filter_fakes/fake_filter.go . Filter

var _ = Describe("Filter", func() {

	var (
		filter  Filter
		matches bool
		event   *v1.Event
		logger *lagertest.TestLogger
	)

	Describe("#Matches", func() {

		BeforeEach(func() {
			event = &v1.Event{ObjectMeta: v12.ObjectMeta{UID: "event-uid"}}
			logger = lagertest.NewTestLogger("filter")
		})

		JustBeforeEach(func() {
			filter = NewFilter(logger)
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

			It("should log", func(){
				Eventually(logger.Buffer()).Should(gbytes.Say("event-uid"))
			})
		})
	})
})
