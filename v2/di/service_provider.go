package di

// ServiceProvider is an interface used to get a service
// from a container.
type ServiceProvider interface {
	GetService(name string) interface{}
}
