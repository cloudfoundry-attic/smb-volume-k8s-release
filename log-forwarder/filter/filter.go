package filter

import (
	"code.cloudfoundry.org/lager"
	v1 "k8s.io/api/core/v1"
)

type filter struct {
	logger lager.Logger
}

type Filter interface {
	Matches(event *v1.Event) (bool)
}

func NewFilter(logger lager.Logger) Filter {
	return filter{
		logger,
	}
}

func (f filter) Matches(event *v1.Event) (bool) {
	if event.Reason == "FailedMount" {
		f.logger.Info("matched", lager.Data{"event": event.UID})
		return true
	}
	return false
}

