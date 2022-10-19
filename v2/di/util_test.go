package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetService_GivenType_ReturnsService(t *testing.T) {
	ctor := func() testDependency {
		return testDependency{}
	}

	ctn := NewContainer()
	ctn.AddService(ctor)

	v := GetService[testDependency](ctn)
	assert.NotNil(t, v)
}

func TestGetService_GivenPointerType_ReturnsService(t *testing.T) {
	ctor := func() *testDependency {
		return &testDependency{}
	}

	ctn := NewContainer()
	ctn.AddService(ctor)

	v := GetService[*testDependency](ctn)
	assert.NotNil(t, v)
}

func TestGetService_GivenInterfaceType_ReturnsService(t *testing.T) {
	ctor := func() TestService {
		return testService{}
	}

	ctn := NewContainer()
	ctn.AddService(ctor)

	v := GetService[TestService](ctn)
	assert.NotNil(t, v)
}
