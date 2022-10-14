package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScope_GetService(t *testing.T) {
	t.Run("Where Service Exists", func(t *testing.T) {
		ctor1 := func() *testDependency {
			return &testDependency{}
		}
		ctor2 := func() *testDependency2 {
			return &testDependency2{}
		}

		ctn := NewContainer()
		ctn.AddService(ctor1)
		ctn.AddService(ctor2).SetName("MyService")

		s := ctn.CreateScope()

		v, ok := s.GetService("MyService").(*testDependency2)
		assert.NotNil(t, v)
		assert.True(t, ok)
	})

	t.Run("Where Build Fails", func(t *testing.T) {
		ctor := func() (*testDependency, error) {
			return nil, assert.AnError
		}

		ctn := NewContainer()
		ctn.AddService(ctor).SetName("MyService")

		s := ctn.CreateScope()

		defer func() {
			err := recover().(error)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), assert.AnError.Error())
		}()

		// Should panic
		_ = s.GetService("MyService")
	})

	t.Run("Where The Service Does Not Exist", func(t *testing.T) {
		ctn := NewContainer()
		s := newScope(ctn)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic")
			}
		}()

		_ = s.GetService("MyService")
	})

	t.Run("Where Service Is Singleton", func(t *testing.T) {
		ctor1 := func() *testDependency {
			return &testDependency{}
		}

		ctn := NewContainer()
		ctn.AddService(ctor1).SetName("MyService").AsSingleton()

		initial := ctn.GetService("MyService")

		s := ctn.CreateScope()

		v, ok := s.GetService("MyService").(*testDependency)
		assert.NotNil(t, v)
		assert.Same(t, initial, v)
		assert.True(t, ok)
	})

	t.Run("Where Service Is Transient", func(t *testing.T) {
		ctor1 := func() *testDependency {
			return &testDependency{}
		}

		ctn := NewContainer()
		ctn.AddService(ctor1).SetName("MyService").AsTransient()

		s := ctn.CreateScope()

		initial := s.GetService("MyService")

		v, ok := s.GetService("MyService").(*testDependency)
		assert.NotNil(t, v)
		assert.Same(t, initial, v)
		assert.True(t, ok)
	})

	t.Run("Where Service Is Scoped", func(t *testing.T) {
		ctor1 := func() *testDependency {
			return &testDependency{}
		}

		ctn := NewContainer()
		ctn.AddService(ctor1).SetName("MyService").AsScoped()

		s := ctn.CreateScope()

		initial := s.GetService("MyService")
		second := s.GetService("MyService")

		// Services should be the same object as the scope should
		// retain the build services.
		assert.Same(t, initial, second)
	})
}
