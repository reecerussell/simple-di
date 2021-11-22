package di

import (
	"context"
	"math/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestService interface{}

type testService struct {
	dep *testDependency
	x   int
}

type testDependency struct{}

// With no imagination, this is just another test dependency.
type testDependency2 struct{}

func TestNewService_GivenValidCtorFunc_ReturnsService(t *testing.T) {
	t.Run("Where Return Value Is Interface", func(t *testing.T) {
		f := func() TestService {
			return &testService{}
		}

		s := NewService(f)
		assert.Equal(t, "di.TestService", s.name)
		assert.Equal(t, LifetimeTransient, s.lifetime)

		_, ok := reflect.New(s.typ).Interface().(*TestService)
		assert.True(t, ok)
	})

	t.Run("Where Return Value Is Ptr", func(t *testing.T) {
		f := func() *testService {
			return &testService{}
		}

		s := NewService(f)
		assert.Equal(t, "di.testService", s.name)
		assert.Equal(t, LifetimeTransient, s.lifetime)

		_, ok := reflect.New(s.typ).Elem().Interface().(*testService)
		assert.True(t, ok)
	})
}

func TestNewService_GivenInvalidCtorFunc_Panics(t *testing.T) {
	assert.Panics(t, func() {
		// f does is not valid because a func cannot only return an error.
		f := func() error {
			return nil
		}

		_ = NewService(f)
	})

	assert.Panics(t, func() {
		// f does is not valid because a func should return
		// an error value last.
		f := func() (error, interface{}) {
			return nil, nil
		}

		_ = NewService(f)
	})

	assert.Panics(t, func() {
		// f does is not valid because a func should only
		// return a single interface value.
		f := func() (interface{}, interface{}) {
			return nil, nil
		}

		_ = NewService(f)
	})

	assert.Panics(t, func() {
		// f does is not valid because a func should only
		// return between 1 and 2 values.
		f := func() (interface{}, interface{}, error) {
			return nil, nil, nil
		}

		_ = NewService(f)
	})

	assert.Panics(t, func() {
		// f does is not valid because a func should return a value.
		f := func() {}

		_ = NewService(f)
	})

	assert.Panics(t, func() {
		// f is not a func value.
		f := "this is not a func"

		_ = NewService(f)
	})
}

func TestService_SetName(t *testing.T) {
	t.Run("Given Valid Name", func(t *testing.T) {
		s := &Service{}
		name := "MyService"
		s.SetName(name)
		assert.Equal(t, name, s.name)
	})

	t.Run("Given Empty Name", func(t *testing.T) {
		name := "MyService"
		s := &Service{name: name}

		// Does not set name.
		s.SetName("")
		assert.Equal(t, name, s.name)
	})
}

func TestService_Name(t *testing.T) {
	name := "MyService"
	s := &Service{name: name}
	assert.Equal(t, name, s.Name())
}

func TestService_SetDispose(t *testing.T) {
	s := &Service{}
	s.SetDispose(func(ctx context.Context, i interface{}) {
		// empty dispose func
	})

	assert.NotNil(t, s.dipsose)
}

func TestService_Dispose(t *testing.T) {
	t.Run("Where Dispose Has Been Set", func(t *testing.T) {
		called := false
		instance := "some service"
		s := &Service{impl: instance}
		s.SetDispose(func(ctx context.Context, i interface{}) {
			assert.Equal(t, instance, i)
			called = true
		})

		s.Dispose(context.Background())
		assert.True(t, called)
		assert.Nil(t, s.impl)
	})

	t.Run("Where Dispose Has Not Been Set", func(t *testing.T) {
		s := &Service{}

		s.Dispose(context.Background())
		assert.Nil(t, s.impl)
	})
}

func TestService_AsSingleton(t *testing.T) {
	s := &Service{lifetime: LifetimeTransient}
	s.AsSingleton()

	assert.Equal(t, LifetimeSingleton, s.lifetime)
}

func TestService_AsTransient(t *testing.T) {
	s := &Service{lifetime: LifetimeSingleton}
	s.AsTransient()

	assert.Equal(t, LifetimeTransient, s.lifetime)
}

func TestService_Build(t *testing.T) {
	t.Run("Given Transient Service", func(t *testing.T) {
		ctn := NewContainer()
		ctor := func() (*testService, error) {
			return &testService{
				x: rand.Int(),
			}, nil
		}
		s := &Service{
			ctor:     ctor,
			typ:      reflect.TypeOf(&testService{}),
			lifetime: LifetimeTransient,
		}

		v1, err := s.build(ctn)
		assert.NotNil(t, v1)
		assert.Nil(t, err)

		v2, err := s.build(ctn)
		assert.NotNil(t, v2)
		assert.Nil(t, err)

		// Assert that the two builds are different
		// as the service is transient.
		assert.NotSame(t, v1, v2)
	})

	t.Run("Given Singleton Service", func(t *testing.T) {
		ctn := NewContainer()
		ctor := func() (*testService, error) {
			return &testService{
				x: rand.Int(),
			}, nil
		}
		s := &Service{
			ctor:     ctor,
			typ:      reflect.TypeOf(&testService{}),
			lifetime: LifetimeSingleton,
		}

		v1, err := s.build(ctn)
		assert.NotNil(t, v1)
		assert.Nil(t, err)

		v2, err := s.build(ctn)
		assert.NotNil(t, v2)
		assert.Nil(t, err)

		// Assert that the two builds are the same
		// as the service is a singleton.
		assert.Same(t, v1, v2)
	})

	t.Run("Where Ctor Returns Error", func(t *testing.T) {
		ctn := NewContainer()
		ctor := func() (*testService, error) {
			return nil, assert.AnError
		}
		s := &Service{
			ctor: ctor,
			typ:  reflect.TypeOf(&testService{}),
		}

		v1, err := s.build(ctn)
		assert.Nil(t, v1)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Where Service Has Dependency", func(t *testing.T) {
		ctn := NewContainer()
		ctn.services = []*Service{
			{
				name:     "di.testDependency",
				typ:      reflect.TypeOf(&testDependency{}),
				lifetime: LifetimeTransient,
				ctor: func() *testDependency {
					return &testDependency{}
				},
			},
		}
		ctor := func(d *testDependency) (*testService, error) {
			return &testService{
				dep: d,
				x:   rand.Int(),
			}, nil
		}
		s := &Service{
			ctor:     ctor,
			typ:      reflect.TypeOf(&testService{}),
			lifetime: LifetimeSingleton,
		}

		v, err := s.build(ctn)
		assert.NotNil(t, v)
		assert.Nil(t, err)

		ts := v.(*testService)
		assert.NotNil(t, ts.dep)
	})

	t.Run("Where Service Cannot Find Dependency", func(t *testing.T) {
		ctn := NewContainer()
		ctn.services = []*Service{
			{
				name:     "di.testDependency",
				typ:      reflect.TypeOf(&testDependency{}),
				lifetime: LifetimeTransient,
				ctor: func() *testDependency {
					return &testDependency{}
				},
			},
		}
		ctor := func(d *testDependency, d2 *testDependency2) (*testService, error) {
			return &testService{
				dep: d,
				x:   rand.Int(),
			}, nil
		}
		s := &Service{
			ctor:     ctor,
			typ:      reflect.TypeOf(&testService{}),
			lifetime: LifetimeSingleton,
		}

		v, err := s.build(ctn)
		assert.Nil(t, v)
		assert.NotNil(t, err)
	})

	t.Run("Where Service Dependency Failed To Build", func(t *testing.T) {
		ctn := NewContainer()
		ctn.services = []*Service{
			{
				name:     "di.testDependency",
				typ:      reflect.TypeOf(&testDependency{}),
				lifetime: LifetimeTransient,
				ctor: func() (*testDependency, error) {
					return nil, assert.AnError
				},
			},
		}
		ctor := func(d *testDependency, d2 *testDependency2) (*testService, error) {
			return &testService{
				dep: d,
				x:   rand.Int(),
			}, nil
		}
		s := &Service{
			ctor:     ctor,
			typ:      reflect.TypeOf(&testService{}),
			lifetime: LifetimeSingleton,
		}

		v, err := s.build(ctn)
		assert.Nil(t, v)
		assert.Contains(t, err.Error(), assert.AnError.Error())
	})
}
