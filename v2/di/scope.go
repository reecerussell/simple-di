package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Scope struct {
	mu  *sync.Mutex
	ctn *Container
	ctx context.Context

	// A map of scoped services, where the key is the type
	// of the service and the value is the built service.
	services map[reflect.Type]interface{}
}

func newScope(ctn *Container, ctx context.Context) *Scope {
	return &Scope{
		mu:       &sync.Mutex{},
		ctn:      ctn,
		ctx:      ctx,
		services: make(map[reflect.Type]interface{}),
	}
}

func (s *Scope) GetService(name string) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	svc := s.ctn.getServiceInfo(name)
	if svc.lifetime != LifetimeScoped {
		return s.ctn.GetService(name)
	}
	impl, ok := s.services[svc.typ]
	if ok {
		return impl
	}
	impl, err := svc.build(s.getService)
	if err != nil {
		panic(fmt.Errorf("container: failed to build %s, %v", svc.Name(), err))
	}
	s.services[svc.typ] = impl
	return impl
}

// getService wraps the Scope's Container's implementation of
// getService(reflect.Type) to provide scoped services and the
// Scope's context.Context.
func (s *Scope) getService(typ reflect.Type) (interface{}, error) {
	if typ.String() == "context.Context" {
		return s.ctx, nil
	}
	impl, ok := s.services[typ]
	if ok {
		return impl, nil
	}
	return s.ctn.getService(typ)
}
