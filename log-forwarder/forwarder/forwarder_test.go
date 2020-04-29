package forwarder_test

import (
	. "code.cloudfoundry.org/volume-services-log-forwarder/forwarder"
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder/fluentshims/fluent_fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./fluentshims/fluent_fake/fake_fluent.go ./fluentshims FluentInterface

var _ = Describe("Forwarder", func() {

	Describe("#Forward", func() {
		var (
			forwarder Forwarder
			event *v1.Event
			err error
			fakeFluent *fluent_fake.FakeFluentInterface
		)

		BeforeEach(func() {
			fakeFluent = &fluent_fake.FakeFluentInterface{}
		})

		JustBeforeEach(func() {
			forwarder = NewForwarder(fakeFluent)
			err = forwarder.Forward(event)
		})

		Context("given an event", func() {

			It("should post to fluentd", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeFluent.PostCallCount()).To(Equal(1))
			})
		})
	})
})
