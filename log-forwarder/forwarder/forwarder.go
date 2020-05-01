package forwarder

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/volume-services-log-forwarder/forwarder/fluentshims"
)

const TAG = "fluentd_dest"
const SOURCE_TYPE = "VOL"

type forwarder struct {
	fluent fluentshims.FluentInterface
	logger lager.Logger
}

type Forwarder interface {
	Forward(appId string, instanceId string, log string) error
}

func NewForwarder(logger lager.Logger, fluent fluentshims.FluentInterface) forwarder {
	return forwarder{
		fluent: fluent,
		logger: logger,
	}
}

func (f forwarder) Forward(appId string, instanceId string, message string) error {
	msg := map[string]string{
		"app_id":      appId,
		"instance_id": instanceId,
		"source_type": SOURCE_TYPE,
		"log": message,
	}
	f.logger.Info("forwarding", lager.Data{"message": msg})
	return f.fluent.Post(TAG, msg)
}

