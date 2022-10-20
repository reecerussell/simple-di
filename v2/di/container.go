package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Container is a simple dependency injection container.
type Container struct {
	mu       *sync.RWMutex
	services []*Service
}

// NewContainer returns a new Container.
func NewContainer() *Container {
	return &Container{
		mu:       &sync.RWMutex{},
		services: make([]*Service, 0),
	}
}

// GetService is used to resolve a service by name. If the service
// does not exist, it will panic.
//
// This function panics instead of returning an error, so that it
// can be called inline, without the extra bulk of handling an error.
func (ctn *Container) GetService(name string) interface{} {
	ctn.mu.RLock()
	defer ctn.mu.RUnlock()

	for _, s := range ctn.services {
		if s.name != name {
			continue
		}

		v, err := s.build(ctn.getService)
		if err != nil {
			panic(fmt.Errorf("container: failed to build %s, %v", s.Name(), err))
		}

		return v
	}

	panic(fmt.Errorf("container: could not find service, %s", name))
}

// getService is an internal function used to resolve a service by its type.
// This is used by Service.build() to resolve dependencies.
func (ctn *Container) getService(t reflect.Type) (interface{}, error) {
	ctn.mu.RLock()
	defer ctn.mu.RUnlock()

	for _, s := range ctn.services {
		if s.typ != t {
			continue
		}

		v, err := s.build(ctn.getService)
		if err != nil {
			return nil, fmt.Errorf("container: failed to build %s, %v", s.Name(), err)
		}

		return v, nil
	}

	return nil, fmt.Errorf("container: failed to resolve %s", t.Name())
}

// GetServices is used to retrievean array of services of a given type.
func (ctn *Container) GetServices(t reflect.Type) []interface{} {
	ctn.mu.RLock()
	defer ctn.mu.RUnlock()

	svcs := make([]interface{}, 0)
	for _, s := range ctn.services {
		if s.typ != t {
			continue
		}
		v, err := s.build(ctn.getService)
		if err != nil {
			panic(fmt.Errorf("container: failed to build %s, %v", s.Name(), err))
		}
		svcs = append(svcs, v)
	}
	return svcs
}

// AddService adds a new service definition to the container. The ctor argument
// should be the constructor function, which is used to build the service.
//
// A constructor function can contain an range of arguments, however, either
// return an interface, or an interface and error: func() MyService or
// func() (MyService, error).
func (ctn *Container) AddService(ctor interface{}) *Service {
	ctn.mu.Lock()
	defer ctn.mu.Unlock()

	s := NewService(ctor)
	ctn.services = append(ctn.services, s)
	return s
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

	for _, s := range ctn.services {
		s.Dispose(ctx)
	}
}

func (ctn *Container) getServiceInfo(name string) *Service {
	ctn.mu.RLock()
	defer ctn.mu.RUnlock()

	for _, s := range ctn.services {
		if s.name == name {
			return s
		}
	}

	panic(fmt.Errorf("container: could not find service, %s", name))
}

// CreateScope is used to create a scoped service provider.
func (ctn *Container) CreateScope() *Scope {
	return ctn.CreateScopeWithContext(context.Background())
}

// CreateScopeWithContext is used to create a scope service provider,
// with the given context.Context configured.
func (ctn *Container) CreateScopeWithContext(ctx context.Context) *Scope {
	return newScope(ctn, ctx)
}
