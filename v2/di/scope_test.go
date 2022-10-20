package di

import (
	"context"
	"sync/atomic"
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
		s := newScope(ctn, context.TODO())

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

	t.Run("Where context.Context Is A Dependency", func(t *testing.T) {
		ctx := context.Background()
		ctor1 := func(c context.Context) *testDependency {
			assert.Same(t, ctx, c)
			return &testDependency{}
		}

		ctn := NewContainer()
		ctn.AddService(ctor1).SetName("MyService").AsScoped()

		s := ctn.CreateScopeWithContext(ctx)
		v, ok := s.GetService("MyService").(*testDependency)
		assert.True(t, ok)
		assert.NotNil(t, v)
	})

	t.Run("Where Service Has A Scoped Dependency (not built)", func(t *testing.T) {
		dep := &testDependency{}
		ctor1 := func() *testDependency {
			return dep
		}
		ctor2 := func(x *testDependency) *testDependency2 {
			assert.Same(t, dep, x) // ensure is the built version
			return &testDependency2{}
		}

		ctn := NewContainer()
		ctn.AddService(ctor1).SetName("MyService").AsScoped()
		ctn.AddService(ctor2).SetName("MyService2").AsScoped()

		s := ctn.CreateScope()

		v, ok := s.GetService("MyService2").(*testDependency2)
		assert.True(t, ok)
		assert.NotNil(t, v)
	})

	t.Run("Where Service Has A Scoped Dependency (built)", func(t *testing.T) {
		dep := &testDependency{}
		ctor1Count := (int32)(0)
		ctor1 := func() *testDependency {
			atomic.AddInt32(&ctor1Count, 1)
			return dep
		}
		ctor2 := func(x *testDependency) *testDependency2 {
			assert.Same(t, dep, x) // ensure is the built version
			return &testDependency2{}
		}

		ctn := NewContainer()
		ctn.AddService(ctor1).SetName("MyService").AsScoped()
		ctn.AddService(ctor2).SetName("MyService2").AsScoped()

		s := ctn.CreateScope()

		// Build Dependency
		_ = s.GetService("MyService")

		v, ok := s.GetService("MyService2").(*testDependency2)
		assert.True(t, ok)
		assert.NotNil(t, v)

		// Scoped service should only be built once.
		assert.Equal(t, int32(1), ctor1Count)
	})

	t.Run("Where Service Fails To Build", func(t *testing.T) {
		ctor := func() (*testDependency, error) {
			return nil, assert.AnError
		}

		ctn := NewContainer()
		ctn.AddService(ctor).SetName("MyService").AsScoped()

		s := ctn.CreateScope()

		assert.Panics(t, func() {
			_ = s.GetService("MyService")
		})
	})
}
