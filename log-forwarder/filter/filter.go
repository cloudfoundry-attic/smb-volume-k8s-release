package filter

import (
	v1 "k8s.io/api/core/v1"
	"log"
)

type filter struct {
}

type Filter interface {
	Matches(event *v1.Event) (bool)
}

func NewFilter() Filter {
	return filter{}
}

func (filter filter) Matches(event *v1.Event) (bool) {
	if event.Reason == "FailedMount" {
		log.Print("matched")
		return true
	}
	return false
}

