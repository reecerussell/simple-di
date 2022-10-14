package di

import "reflect"

func GetService[T any](sp ServiceProvider) T {
	t := reflect.TypeOf(*new(T))
	return GetServiceByName[T](sp, t.Name())
}

func GetServiceByName[T any](sp ServiceProvider, name string) T {
	return sp.GetService(name).(T)
}
