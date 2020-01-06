package store

type ServiceInstance struct {
	ServiceID  string
	PlanID     string
	Parameters map[string]interface{}
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ServiceInstanceStore
type ServiceInstanceStore interface {
	Get(string) (ServiceInstance, bool)
	Add(string, ServiceInstance) error
	Remove(string)
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

func (i *InMemoryServiceInstanceStore) Get(k string) (ServiceInstance, bool) {
	serviceInstance, found := i.internalMap[k]
	return serviceInstance, found
}

func (i *InMemoryServiceInstanceStore) Remove(k string) {
	delete(i.internalMap, k)
}