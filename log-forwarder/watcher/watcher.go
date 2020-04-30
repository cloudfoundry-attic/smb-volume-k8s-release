package watcher

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/volume-services-log-forwarder/filter"
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder"
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Watcher struct {
	logger       lager.Logger
	filter       filter.Filter
	forwarder    forwarder.Forwarder
	podInterface corev1.PodInterface
}

func NewWatcher(logger lager.Logger, filter filter.Filter, forwarder forwarder.Forwarder, podInterface corev1.PodInterface) Watcher {
	return Watcher{
		logger:       logger,
		filter:       filter,
		forwarder:    forwarder,
		podInterface: podInterface,
	}
}

func (w Watcher) Watch(obj interface{}) {
	logger := w.logger.Session("watch")
	logger.Info("start")
	defer logger.Info("end")

	if w.filter.Matches(obj.(*v1.Event)) {

		pod, err := w.podInterface.Get(context.TODO(), obj.(*v1.Event).InvolvedObject.Name, metav1.GetOptions{})
		if err != nil {
			w.logger.Error("pod-fetch-failed", err)
			return
		}

		if appId, ok := pod.Labels["cloudfoundry.org/app_guid"]; ok {
			instanceId := obj.(*v1.Event).InvolvedObject.UID
			log := obj.(*v1.Event).Message

			if err = w.forwarder.Forward(appId, string(instanceId), log); err != nil {
				w.logger.Error("forwarding-failed", err)
			}
		} else {
			w.logger.Debug("app-guid-not-in-pod-spec", lager.Data{"pod-labels": pod.Labels, "pod-id": obj.(*v1.Event).InvolvedObject.Name})
		}
	}
}
