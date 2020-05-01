package watcher_test

import (
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/volume-services-log-forwarder/filter"
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder/forwarder_fakes"
	"code.cloudfoundry.org/volume-services-log-forwarder/watcher/k8s_fakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "code.cloudfoundry.org/volume-services-log-forwarder/watcher"
	"github.com/onsi/gomega/gbytes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./k8s_fakes/fake_pod_interface.go k8s.io/client-go/kubernetes/typed/core/v1.PodInterface

var _ = Describe("Watcher", func() {

	var (
		watcher Watcher
		err error
		logger *lagertest.TestLogger
	)

	Describe("#Watch", func() {

		var (
			fltr             filter.Filter
			fakeForwarder    *forwarder_fakes.FakeForwarder
			fakePodInterface *k8s_fakes.FakePodInterface

			event *v1.Event
		)

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("test-log-forwarder")
			fltr = filter.NewFilter(logger)
			fakeForwarder = &forwarder_fakes.FakeForwarder{}
			fakePodInterface = &k8s_fakes.FakePodInterface{}
			event = &v1.Event{}
		})

		JustBeforeEach(func() {
			watcher = NewWatcher(logger, fltr, fakeForwarder, fakePodInterface)
			watcher.Watch(event)
		})

		Context("given a MountFailed event", func() {

			BeforeEach(func() {
				event.Reason = "FailedMount"
				event.Message = "some mount failure message"
				event.InvolvedObject = v1.ObjectReference{UID: "12345" }
			})

			BeforeEach(func() {
				fakePodInterface.GetReturns(&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"cloudfoundry.org/app_guid": "app-id",
						},
					},
				}, nil)
			})

			It("should forward the event", func(){
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeForwarder.ForwardCallCount()).To(Equal(1))
				appId, instanceId, log := fakeForwarder.ForwardArgsForCall(0)
				Expect(appId).To(Equal("app-id"))
				Expect(instanceId).To(Equal("12345"))
				Expect(log).To(Equal("some mount failure message"))
			})

			Context("when the pod can't be got", func(){
				BeforeEach(func() {
					fakePodInterface.GetReturns(nil, errors.New("something bad happened"))
				})
				It("should log an error", func(){
					Expect(logger.Buffer()).To(gbytes.Say("something bad happened"))
				})
			})

			Context("when the pod isnt a cf app", func(){
				BeforeEach(func() {
					fakePodInterface.GetReturns(&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{},
						},
					}, nil)
				})
				It("should log and skip the event", func(){
					Expect(logger.Buffer()).To(gbytes.Say("app-guid-not-in-pod-spec"))
					Expect(fakeForwarder.ForwardCallCount()).To(BeZero())
				})
			})

			Context("when unable to forward the event", func(){
				BeforeEach(func() {
					fakeForwarder.ForwardReturns(errors.New("forward-err"))
				})
				It("should log", func(){
					Expect(logger.Buffer()).To(gbytes.Say("forward-err"))
				})
			})
		})

		Context("given an event that isn't a mount failure", func() {

			It("should not forward the event", func(){
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeForwarder.ForwardCallCount()).To(Equal(0))
			})
		})
	})
})

