package reflection

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Digs through the object's fields to get the field of the specified type. Will do its best to auto-cast it to T (which can be concrete or interface type).
// Use a pointer for T in order to get mutable access to the actual object.
func GetField[T any](obj any, fields ...string) T {
	reflectedObj := reflect.ValueOf(obj)
	if reflectedObj.Kind() == reflect.Ptr {
		reflectedObj = reflectedObj.Elem()
	}
	for _, field := range fields {
		if reflectedObj.Kind() == reflect.Ptr {
			reflectedObj = reflectedObj.Elem()
		}
		reflectedObj = reflectedObj.FieldByName(field)
		if !reflectedObj.IsValid() {
			panic(fmt.Sprintf("field %s not found", field))
		}
	}
	generic := reflect.TypeFor[T]()
	isAbstract := func(kind reflect.Kind) bool {
		return kind == reflect.Ptr || kind == reflect.Interface
	}
	ptr := unsafe.Pointer(reflectedObj.UnsafeAddr())
	ptrValue := reflect.NewAt(reflectedObj.Type(), ptr)
	if ptrValue.Kind() != reflect.Ptr {
		return ptrValue.Interface().(T) // not sure this scenario can happen... this probably will panic if it does...
	}
	elem := ptrValue.Elem()
	if isAbstract(elem.Kind()) {
		if isAbstract(generic.Kind()) {
			return elem.Interface().(T) // obj and T are both abstract - return directly
		}
		if elem.Kind() == reflect.Interface { // obj being interface type requires special handling
			elem = elem.Elem()
			if elem.Kind() == reflect.Ptr {
				return elem.Elem().Interface().(T) // obj is an interface of a pointer, T is concrete -> double-dereference (->ptr->struct) so it's of type T
			}
			return elem.Interface().(T) // obj is an interface of a concrete type, T is concrete -> single-dereference (->struct) so it's of type T
		}
		return elem.Elem().Interface().(T) // obj is a ptr, but T is concrete. Dereference so it's of type T
	}
	if isAbstract(generic.Kind()) {
		return elem.Addr().Interface().(T) //obj is concrete (e.g. struct), but T is abstract. Cast to T
	}
	return elem.Interface().(T) // obj and T are both concrete - return directly
}

// Digs through the object's fields to set the last field to the target value. Usage is SetField(obj, "field1", "field2", "field3")(value)
func SetField(obj any, fields ...string) func(any) {
	reflectedObj := reflect.ValueOf(obj)
	if reflectedObj.Kind() == reflect.Ptr {
		reflectedObj = reflectedObj.Elem()
	}
	for _, field := range fields {
		if reflectedObj.Kind() == reflect.Ptr {
			reflectedObj = reflectedObj.Elem()
		}
		reflectedObj = reflectedObj.FieldByName(field)
		if !reflectedObj.IsValid() {
			panic(fmt.Sprintf("field %s not found", field))
		}
	}
	return func(target any) {
		reflectedPtr := reflect.NewAt(reflectedObj.Type(), unsafe.Pointer(reflectedObj.UnsafeAddr())).Elem()
		if target == nil {
			reflectedPtr.SetZero()
		} else {
			v := reflect.ValueOf(target)
			reflectedPtr.Set(v)
		}
	}
}
