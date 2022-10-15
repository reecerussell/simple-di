package di

import "reflect"

func GetService[T any](sp ServiceProvider) T {
	t := reflect.TypeOf(new(T))
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return GetServiceByName[T](sp, t.String())
}

func GetServiceByName[T any](sp ServiceProvider, name string) T {
	return sp.GetService(name).(T)
}
