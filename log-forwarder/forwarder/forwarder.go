package forwarder

import (
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder/fluentshims"
	"log"
)

const TAG = "fluentd_dest"
const SOURCE_TYPE = "VOL"

type forwarder struct {
	fluent fluentshims.FluentInterface
}

type Forwarder interface {
	Forward(appId string, instanceId string, log string) error
}

func NewForwarder(fluent fluentshims.FluentInterface) forwarder {
	return forwarder{
		fluent: fluent,
	}
}

func (f forwarder) Forward(appId string, instanceId string, message string) error {
	msg := map[string]string{
		"app_id":      appId,
		"instance_id": instanceId,
		"source_type": SOURCE_TYPE,
		"log": message,
	}
	log.Printf("Forwarding: %#v", msg)
	return f.fluent.Post(TAG, msg)
}

