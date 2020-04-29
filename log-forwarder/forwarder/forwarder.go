package forwarder

import (
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder/fluentshims"
)

const TAG = "fluentd_dest"
const SOURCE_TYPE = "VOL"

type Forwarder struct {
	fluent fluentshims.FluentInterface
}

func NewForwarder(fluent fluentshims.FluentInterface) Forwarder {
	return Forwarder{
		fluent: fluent,
	}
}

func (f Forwarder) Forward(appId string, instanceId string, log string) error {
	msg := map[string]string{
		"app_id":      appId,
		"instance_id": instanceId,
		"source_type": SOURCE_TYPE,
		"log": log,
	}
	return f.fluent.Post(TAG, msg)
}

