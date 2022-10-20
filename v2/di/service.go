package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// ServiceLifetime is a type used to define a service's lifetime.
type ServiceLifetime uint

const (
	// LifetimeSingleton is used to define a Service as a singleton.
	// Which means only a single instance of the Service will be built,
	// then shared across other services.
	LifetimeSingleton ServiceLifetime = iota

	// LifetimeTransient is used to define a Service as transient. Which
	// means a new instance will be instantiated each time the service is resolved.
	LifetimeTransient

	// LifetimeScoped is used to define a Service as scope. Which means
	// a new instance is created for an individual scope, then re-used
	// in that scope.
	LifetimeScoped
)

// DisposeFunc is a function used to clean and dispose a singleton service.
// The argument, i, is the instance of the service.
type DisposeFunc func(ctx context.Context, i interface{})

// Service represents a service within the DI Container. It contains
// information on the type, lifetime and name of the service, as well
// as, how to build it.
type Service struct {
	name     string
	typ      reflect.Type
	lifetime ServiceLifetime
	ctor     interface{}
	mu       sync.Mutex
	impl     interface{}
	dipsose  DisposeFunc
}

// NewService is used to create a new instance of Service. The ctor argument
// should be the constructor function, which is used to build the service.
//
// A constructor function can contain an range of arguments, however, either
// return an interface, or an interface and error: func() MyService or
// func() (MyService, error).
func NewService(ctor interface{}) *Service {
	t := reflect.TypeOf(ctor)
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("service: %s is not a func", t.Name()))
	}

	switch t.NumOut() {
	case 0:
		panic(fmt.Errorf("service: %s should return a value", t.Name()))
	case 1:
		if isTypeError(t.Out(0)) {
			panic(fmt.Errorf("service: %s should return a non-error value", t.Name()))
		}
	case 2:
		if isTypeError(t.Out(0)) {
			panic(fmt.Errorf("service: %s should return (interface{}, error)", t.Name()))
		}

		if !isTypeError(t.Out(1)) {
			panic(fmt.Errorf("service: %s should return (interface{}, error)", t.Name()))
		}
	default:
		panic(fmt.Errorf("service: %s can not contain more than 2 return values", t.Name()))
	}

	st := t.Out(0)
	name := st.String()
	if st.Kind() == reflect.Ptr {
		name = st.Elem().String()
	}

	return &Service{
		name:     name,
		typ:      st,
		lifetime: LifetimeTransient,
		ctor:     ctor,
		mu:       sync.Mutex{},
	}
}

// This is used to determine whether a Type is an error or not.
func isTypeError(t reflect.Type) bool {
	err := reflect.TypeOf((*error)(nil)).Elem()
	return t.Implements(err)
}

// SetName is used to set the name of the Service. Note that
// this is can not be referred to in depedency injection, and
// only when resolving a service through the Container.
//
// If name is empty, the name will not be updated and will remain
// the name of the service interface.
func (s *Service) SetName(name string) *Service {
	if name != "" {
		s.name = name
	}

	return s
}

// Name returns the name of the service. If this has not been manually
// configured, the name of the service type will be returned.
func (s *Service) Name() string {
	return s.name
}

// SetDispose is used to configure a clean up/disposal function for a
// service. This can be used to set a dispose function for a service with
// any lifetime, however, will only be used for Singleton service.
//
// This is not required but is helpful for releasing resources consumed
// by the service.
func (s *Service) SetDispose(f DisposeFunc) *Service {
	s.dipsose = f

	return s
}

// Dispose is used to clean up singleton resources.
func (s *Service) Dispose(ctx context.Context) {
	if s.dipsose != nil {
		s.dipsose(ctx, s.impl)
	}

	s.impl = nil
}

// AsSingleton sets the lifetime of the service to Singleton.
func (s *Service) AsSingleton() *Service {
	s.lifetime = LifetimeSingleton

	return s
}

// AsTransient sets the lifetime of the service to Transient.
func (s *Service) AsTransient() *Service {
	s.lifetime = LifetimeTransient

	return s
}

// AsScoped sets the lifetime of the service to Scoped.
func (s *Service) AsScoped() *Service {
	s.lifetime = LifetimeScoped

	return s
}

// build is used to build a service as well as its dependency chain.
func (s *Service) build(sp func(reflect.Type) (interface{}, error)) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If the service is a singleton and already built, used the
	// build instance instead of creating another.
	if s.impl != nil && s.lifetime == LifetimeSingleton {
		return s.impl, nil
	}

	f := reflect.ValueOf(s.ctor)
	numIn := f.Type().NumIn()
	args := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		arg := f.Type().In(i)
		d, err := sp(arg)
		if err != nil {
			return nil, err
		}

		args[i] = reflect.ValueOf(d)
	}

	out := f.Call(args)
	if len(out) == 2 {
		err := out[1].Interface()
		if err != nil {
			return nil, err.(error)
		}
	}

	impl := out[0].Interface()

	// If the sevrice is a singleton, store the built instance in memory.
	if s.lifetime == LifetimeSingleton {
		s.impl = impl
	}

	return impl, nil
}
