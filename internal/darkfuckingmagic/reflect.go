package darkfuckingmagic

import (
	"log"
	"reflect"
)

// UnexportedField returns the value of an unexported field.
func UnexportedField[T any](v any, name string) T {
	rv := reflect.Indirect(reflect.ValueOf(v))
	rfield := rv.FieldByName(name)
	if rfield == (reflect.Value{}) {
		log.Panicf("field %s not found in %T", name, v)
	}
	ptr := rfield.Addr().UnsafePointer()
	return *(*T)(ptr)
}
