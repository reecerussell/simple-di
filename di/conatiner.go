package di

// Container is a simple dependency injection container.
type Container struct {
	srvs    map[string]interface{}
	srvConf map[string]*ServiceConfig
}

// NewContainer returns a new Container.
func NewContainer() *Container {
	return &Container{
		srvs:    make(map[string]interface{}),
		srvConf: make(map[string]*ServiceConfig),
	}
}

// BuildFunc is a function used to build a service.
type BuildFunc func(ctn *Container) interface{}

// ServiceConfig represents a service within the Container.
type ServiceConfig struct {
	Singleton bool
	Build     BuildFunc
}

// GetService attempts to resolve a service by name.
func (ctn *Container) GetService(name string) interface{} {
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
	s := &ServiceConfig{
		Singleton: singleton,
		Build:     builder,
	}
	ctn.srvConf[name] = s

	return &ServiceBuilder{s: s}
}

// ServiceBuilder is a type used to provide a fluent-like API
// when adding a service to the container.
type ServiceBuilder struct {
	s *ServiceConfig
}
