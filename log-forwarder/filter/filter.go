package filter

import (
	v1 "k8s.io/api/core/v1"
)

type Filter struct {
}

func NewFilter() Filter {
	return Filter{}
}

func (filter Filter) Matches(event *v1.Event) (bool) {
	if event.Reason == "MountFailed" {
		return true
	}
	return false
}

