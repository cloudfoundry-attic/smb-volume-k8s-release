package store

type ServiceInstance struct {
	ServiceID string
	PlanID string
	Parameters map[string]interface{}
}
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ServiceInstanceStore
type ServiceInstanceStore interface {
	Get(string) ServiceInstance
}

type InMemoryServiceInstanceStore struct {}

func (InMemoryServiceInstanceStore) Get(string) ServiceInstance {
	panic("implement me")
}