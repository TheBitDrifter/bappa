package table

import (
	"reflect"
)

var nextElementTypeID ElementTypeID = 1

var _ ElementType = elementType{}

type elementType struct {
	typ reflect.Type
	id  ElementTypeID
}

func newElementType[T any]() elementType {
	var t T
	typ := reflect.TypeOf(t)
	elementType := elementType{
		id:  nextElementTypeID,
		typ: typ,
	}
	nextElementTypeID++
	return elementType
}

func (et elementType) ID() ElementTypeID {
	return et.id
}

func (et elementType) Type() reflect.Type {
	return et.typ
}

func (et elementType) Size() uint32 {
	size, align := et.typ.Size(), uintptr(et.typ.Align())
	return uint32((size + (align - 1)) / align * align)
}
