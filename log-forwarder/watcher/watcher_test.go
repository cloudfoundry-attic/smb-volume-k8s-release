package watcher_test

import (
	"code.cloudfoundry.org/volume-services-log-forwarder/filter"
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder/forwarder_fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "code.cloudfoundry.org/volume-services-log-forwarder/watcher"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Watcher", func() {

	var (
		watcher Watcher
		err error
	)

	Describe("#Watch", func() {

		var (
			fltr          filter.Filter
			fakeForwarder *forwarder_fakes.FakeForwarder

			event *v1.Event
		)

		BeforeEach(func() {
			fltr = filter.NewFilter()
			fakeForwarder = &forwarder_fakes.FakeForwarder{}
			event = &v1.Event{}
		})

		JustBeforeEach(func() {
			watcher = NewWatcher(fltr, fakeForwarder)
			watcher.Watch(event)
		})

		Context("given a MountFailed event", func() {

			BeforeEach(func() {
				event.Reason = "MountFailed"
			})

			It("should forward the event", func(){
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeForwarder.ForwardCallCount()).To(Equal(1))
			})
		})

		Context("given an event that isnt a mount failure", func() {

			It("should not forward the event", func(){
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeForwarder.ForwardCallCount()).To(Equal(0))
			})
		})
	})
})

