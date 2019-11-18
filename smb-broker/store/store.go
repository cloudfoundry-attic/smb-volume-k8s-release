package store

type ServiceInstance struct {
	ServiceID  string
	PlanID     string
	Parameters map[string]interface{}
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ServiceInstanceStore
type ServiceInstanceStore interface {
	Get(string) ServiceInstance
	Add(string, ServiceInstance) error
}

type InMemoryServiceInstanceStore struct{
	internalMap map[string]ServiceInstance
}

func (i *InMemoryServiceInstanceStore) Add(k string, v ServiceInstance) error {
	if i.internalMap == nil {
		i.internalMap = map[string]ServiceInstance{}
	}

	i.internalMap[k] = v
	return nil
}

func (i *InMemoryServiceInstanceStore) Get(k string) ServiceInstance {
	return i.internalMap[k]
}
