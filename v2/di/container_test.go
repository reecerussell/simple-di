package di

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_GetService(t *testing.T) {
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

		v, ok := ctn.GetService("MyService").(*testDependency2)
		assert.NotNil(t, v)
		assert.True(t, ok)
	})

	t.Run("Where Build Fails", func(t *testing.T) {
		ctor := func() (*testDependency, error) {
			return nil, assert.AnError
		}

		ctn := NewContainer()
		ctn.AddService(ctor).SetName("MyService")

		defer func() {
			err := recover().(error)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), assert.AnError.Error())
		}()

		// Should panic
		_ = ctn.GetService("MyService")
	})

	t.Run("Where The Service Does Not Exist", func(t *testing.T) {
		ctn := NewContainer()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic")
			}
		}()

		_ = ctn.GetService("MyService")
	})
}

func TestContainer_GetServices(t *testing.T) {
	t.Run("Where Services Exists", func(t *testing.T) {
		srv1 := &testDependency{}
		srv2 := &testDependency{}
		ctor1 := func() *testDependency {
			return srv1
		}
		ctor2 := func() *testDependency {
			return srv2
		}
		ctor3 := func() *testDependency2 {
			return &testDependency2{}
		}

		ctn := NewContainer()
		ctn.AddService(ctor1)
		ctn.AddService(ctor2)
		ctn.AddService(ctor3)

		arr := ctn.GetServices(reflect.TypeOf(&testDependency{}))
		assert.ElementsMatch(t, arr, []interface{}{srv1, srv2})
	})

	t.Run("Where Build Fails", func(t *testing.T) {
		ctor := func() (*testDependency, error) {
			return nil, assert.AnError
		}

		ctn := NewContainer()
		ctn.AddService(ctor).SetName("MyService")

		defer func() {
			err := recover().(error)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), assert.AnError.Error())
		}()

		// Should panic
		_ = ctn.GetServices(reflect.TypeOf(&testDependency{}))
	})

	t.Run("Where No Services Exist", func(t *testing.T) {
		ctn := NewContainer()

		arr := ctn.GetServices(reflect.TypeOf(&testDependency{}))
		assert.Len(t, arr, 0)
	})
}

func TestContainer_AddService(t *testing.T) {
	ctor := func() interface{} {
		return nil
	}

	ctn := NewContainer()
	s := ctn.AddService(ctor)
	assert.NotNil(t, s)
	assert.Same(t, s, ctn.services[0])
}

func TestContainer_Clean(t *testing.T) {
	hasBeenDisposed := false
	testCtx := context.Background()
	testValue := "My String"

	ctn := NewContainer()
	ctn.AddService(func() interface{} {
		return testValue
	}).
		AsSingleton().
		SetDispose(func(ctx context.Context, i interface{}) {
			assert.Equal(t, testCtx, ctx)
			assert.Equal(t, testValue, i)

			// Proves that the dispose has only been called once.
			assert.False(t, hasBeenDisposed)

			hasBeenDisposed = true
		}).
		SetName("MyService")

	// Builds the service
	_ = ctn.GetService("MyService")

	ctn.Clean(testCtx)

	assert.True(t, hasBeenDisposed)
	assert.Nil(t, ctn.services[0].impl)
}
