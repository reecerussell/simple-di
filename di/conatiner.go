package di

import (
	"context"
	"sync"
)

// Container is a simple dependency injection container.
type Container struct {
	mu      *sync.Mutex
	srvs    map[string]interface{}
	srvConf map[string]*ServiceConfig
}

// NewContainer returns a new Container.
func NewContainer() *Container {
	return &Container{
		mu:      &sync.Mutex{},
		srvs:    make(map[string]interface{}),
		srvConf: make(map[string]*ServiceConfig),
	}
}

// BuildFunc is a function used to build a service.
type BuildFunc func(ctn *Container) interface{}

// DisposeFunc is a function used to clean and dispose a service.
type DisposeFunc func(ctx context.Context, i interface{})

// ServiceConfig represents a service within the Container.
type ServiceConfig struct {
	Singleton bool
	Build     BuildFunc
	Dispose   DisposeFunc
}

// GetService attempts to resolve a service by name.
func (ctn *Container) GetService(name string) interface{} {
	ctn.mu.Lock()
	defer ctn.mu.Unlock()

	conf, ok := ctn.srvConf[name]
	if !ok {
		panic("unable to resolve service: " + name)
	}

	srv, ok := ctn.srvs[name]
	if conf.Singleton && ok {
		return srv
	}

	impl := conf.Build(ctn)

	if conf.Singleton {
		ctn.srvs[name] = impl
	}

	return impl
}

// AddService adds a new service definition to the container.
func (ctn *Container) AddService(name string, builder BuildFunc) *ServiceBuilder {
	return ctn.addService(name, false, builder)
}

// AddSingleton adds a new singleton service definition to the container.
func (ctn *Container) AddSingleton(name string, builder BuildFunc) *ServiceBuilder {
	return ctn.addService(name, true, builder)
}

func (ctn *Container) addService(name string, singleton bool, builder BuildFunc) *ServiceBuilder {
	ctn.mu.Lock()
	defer ctn.mu.Unlock()

	s := &ServiceConfig{
		Singleton: singleton,
		Build:     builder,
	}
	ctn.srvConf[name] = s

	return &ServiceBuilder{s: s}
}

// Clean is used to clean up the services in the container. Once,
// this func has been called, the container can still be used and services
// built. However, this is intended to be called at the end of a program.
//
// If a service has a DisposeFunc, this will be called before it is removed
// from the container. However, if there is no DisposeFunc, the service will
// just be removed.
func (ctn *Container) Clean(ctx context.Context) {
	ctn.mu.Lock()
	defer ctn.mu.Unlock()

	for name, value := range ctn.srvs {
		cnf := ctn.srvConf[name]
		if cnf.Dispose != nil {
			cnf.Dispose(ctx, value)
		}

		delete(ctn.srvs, name)
	}
}

// ServiceBuilder is a type used to provide a fluent-like API
// when adding a service to the container.
type ServiceBuilder struct {
	s *ServiceConfig
}

// Dispose is used to configure a function used to dispose the service.
func (b *ServiceBuilder) Dispose(f DisposeFunc) *ServiceBuilder {
	b.s.Dispose = f

	return b
}
