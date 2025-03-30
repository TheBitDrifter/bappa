package table

import (
	"reflect"
	"unsafe"
)

func newRow(elementType ElementType) Row {
	// Per: https://pkg.go.dev/reflect#Value.CanAddr
	//
	// `func (v Value) CanAddr() bool`

	// "A value is addressable if it is
	// an component of a slice, an component of an addressable array, a
	// field of an addressable struct, or the result of dereferencing
	// a pointer."

	sliceType := reflect.SliceOf(elementType.Type())

	// A pointer to a zero-initialized slice of the slice type.
	slicePtr := reflect.New(sliceType)

	// Dereference the pointer to get an addressable reflect.Value
	// slice.
	addressableSlice := reflect.Indirect(slicePtr)
	return Row(addressableSlice)
}

func (r Row) Interface() any {
	return reflect.Value(r).Interface()
}

func (r Row) UnsafePointer() unsafe.Pointer {
	return reflect.Value(r).UnsafePointer()
}

func (r Row) Type() reflect.Type {
	return reflect.Value(r).Type()
}

func (r Row) CanAddr() bool {
	return reflect.Value(r).CanAddr()
}

func (r Row) setCap(n int) {
	reflect.Value(r).SetCap(n)
}

func (r Row) setLen(n int) {
	reflect.Value(r).SetLen(n)
}

func (r Row) get(i int) reflect.Value {
	return reflect.Value(r).Index(i)
}

func (r Row) set(i int, value reflect.Value) {
	reflect.Value(r).Index(i).Set(value)
}
