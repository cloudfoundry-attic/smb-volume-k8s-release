package watcher

import (
	"code.cloudfoundry.org/volume-services-log-forwarder/filter"
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder"
	v1 "k8s.io/api/core/v1"
)

type Watcher struct {
	filter    filter.Filter
	forwarder forwarder.Forwarder
}

func NewWatcher(filter filter.Filter, forwarder forwarder.Forwarder) Watcher {
	return Watcher{
		filter: filter,
		forwarder: forwarder,
	}
}

func (w Watcher) Watch(obj interface{}) {
	if w.filter.Matches(obj.(*v1.Event)) {
		w.forwarder.Forward("", "", "")
	}
}