package di

import "reflect"

// GetService is generic function used to get a service
// from the given ServiceProvider.
func GetService[T any](sp ServiceProvider) T {
	t := reflect.TypeOf(new(T))
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return GetServiceByName[T](sp, t.String())
}

// GetService is generic function used to get a service
// from the given ServiceProvider, using the specifed name.
func GetServiceByName[T any](sp ServiceProvider, name string) T {
	return sp.GetService(name).(T)
}
