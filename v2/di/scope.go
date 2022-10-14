package di

import "sync"

type Scope struct {
	mu  *sync.Mutex
	ctn *Container

	// A map of scoped services, where the key is the name
	// of the service and the value is the built service.
	services map[string]interface{}
}

func newScope(ctn *Container) *Scope {
	return &Scope{
		mu:       &sync.Mutex{},
		ctn:      ctn,
		services: make(map[string]interface{}),
	}
}

func (s *Scope) GetService(name string) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	info := s.ctn.getServiceInfo(name)
	if info.lifetime != LifetimeScoped {
		return s.ctn.GetService(name)
	}
	impl, ok := s.services[name]
	if ok {
		return impl
	}
	impl = s.ctn.GetService(name)
	s.services[name] = impl
	return impl
}
