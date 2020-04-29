package fluentshims

type FluentInterface interface {
	Post(tag string, message interface{}) error
}
