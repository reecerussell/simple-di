package di

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_GetNonExistantService_Panics(t *testing.T) {
	ctn := NewContainer()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	_ = ctn.GetService("myService")
}

func TestContainer_AddService(t *testing.T) {
	builder := func(ctn *Container) interface{} {
		return nil
	}
	name := "myService"

	ctn := NewContainer()
	sb := ctn.AddService(name, builder)
	assert.NotNil(t, sb)

	srv, ok := ctn.srvConf[name]
	assert.True(t, ok)
	assert.False(t, srv.Singleton)
}

func TestContainer_AddSingleton(t *testing.T) {
	builder := func(ctn *Container) interface{} {
		return nil
	}
	name := "myService"

	ctn := NewContainer()
	sb := ctn.AddSingleton(name, builder)
	assert.NotNil(t, sb)

	srv, ok := ctn.srvConf[name]
	assert.True(t, ok)
	assert.True(t, srv.Singleton)
}

func TestContainer_SingletonNotRecreated(t *testing.T) {
	ctn := NewContainer()

	ctn.srvConf["test"] = &ServiceConfig{
		Singleton: true,
		Build: func(ctn *Container) interface{} {
			srvValue := "My super cool service"
			return &srvValue
		},
	}

	srv := ctn.GetService("test")
	srv2 := ctn.GetService("test")

	if srv != srv2 {
		t.Error("Expected the services to be equal")
	}
}

func TestContainer_TransientIsRecreated(t *testing.T) {
	ctn := NewContainer()

	ctn.srvConf["test"] = &ServiceConfig{
		Singleton: false,
		Build: func(ctn *Container) interface{} {
			srvValue := "My super cool service"
			return &srvValue
		},
	}

	srv := ctn.GetService("test")
	srv2 := ctn.GetService("test")

	if srv == srv2 {
		t.Error("Expected the services to not be equal")
	}
}

func TestContainer_Clean(t *testing.T) {
	hasBeenDisposed := false
	testCtx := context.Background()
	testValue := "My String"

	ctn := NewContainer()
	ctn.AddSingleton("MyService", func(ctn *Container) interface{} {
		return testValue
	}).Dispose(func(ctx context.Context, i interface{}) {
		assert.Equal(t, testCtx, ctx)
		assert.Equal(t, testValue, i)

		// Proves that the dispose has only been called once.
		assert.False(t, hasBeenDisposed)

		hasBeenDisposed = true
	})

	// Builds the service
	_ = ctn.GetService("MyService")

	ctn.Clean(testCtx)

	assert.True(t, hasBeenDisposed)
	assert.Nil(t, ctn.srvs["MyService"])
}

func TestServiceBuilder_Dispose(t *testing.T) {
	s := &ServiceConfig{Dispose: nil}
	b := &ServiceBuilder{s: s}

	r := b.Dispose(func(ctx context.Context, i interface{}) {})
	assert.Equal(t, b, r)
	assert.NotNil(t, s.Dispose)
}
