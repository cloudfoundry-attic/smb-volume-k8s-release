package store

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ServiceInstanceStore
type ServiceInstanceStore interface {
	Get() map[string]interface{}
}