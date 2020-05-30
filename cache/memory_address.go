package cache

import "reflect"

func GetMemoryPointer(v interface{}) uintptr {
	value := reflect.ValueOf(v)
	var u uintptr
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		u = value.Pointer()
	}

	return u
}
