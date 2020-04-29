package forwarder

import (
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder/fluentshims"
	v1 "k8s.io/api/core/v1"
)

type Forwarder struct {
	fluent fluentshims.FluentInterface
}

func NewForwarder(fluent fluentshims.FluentInterface) Forwarder {
	return Forwarder{
		fluent: fluent,
	}
}

func (f Forwarder) Forward(event *v1.Event) error {
	return f.fluent.Post("", nil)
}

