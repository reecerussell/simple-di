package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetService_GivenType_ReturnsService(t *testing.T) {
	ctor := func() testDependency {
		return testDependency{}
	}

	ctn := NewContainer()
	ctn.AddService(ctor)

	v := GetService[testDependency](ctn)
	assert.NotNil(t, v)
}
