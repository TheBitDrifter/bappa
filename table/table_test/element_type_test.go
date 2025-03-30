package table_test

import (
	"reflect"
	"testing"

	"github.com/TheBitDrifter/table"
)

func TestElementType_ID(t *testing.T) {
	startingID := table.Stats.TotalElementTypes()
	tests := []struct {
		name        string
		elementType table.ElementType
		expectedID  int
	}{
		{
			name:        "first element type ID",
			elementType: table.FactoryNewElementType[int](),
			expectedID:  1 + startingID,
		},
		{
			name:        "second element type ID",
			elementType: table.FactoryNewElementType[string](),
			expectedID:  2 + startingID,
		},
		{
			name:        "third element type ID",
			elementType: table.FactoryNewElementType[float64](),
			expectedID:  3 + startingID,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := int(test.elementType.ID()); got != test.expectedID {
				t.Errorf("elementType.ID() == %d, expected %d", got, test.expectedID)
			}
		})
	}
}

func TestElementType_Type(t *testing.T) {
	tests := []struct {
		name         string
		elementType  table.ElementType
		expectedType reflect.Type
	}{
		{
			name:         "int type",
			elementType:  table.FactoryNewElementType[int](),
			expectedType: reflect.TypeOf(int(0)),
		},
		{
			name:         "string type",
			elementType:  table.FactoryNewElementType[string](),
			expectedType: reflect.TypeOf(""),
		},
		{
			name:         "float64 type",
			elementType:  table.FactoryNewElementType[float64](),
			expectedType: reflect.TypeOf(float64(0)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.elementType.Type(); got != test.expectedType {
				t.Errorf("elementType.Type() == %v, expected %v", got, test.expectedType)
			}
		})
	}
}

func TestElementType_Size(t *testing.T) {
	tests := []struct {
		name         string
		elementType  table.ElementType
		expectedSize uint32
	}{
		{
			name:         "size of int",
			elementType:  table.FactoryNewElementType[int](),
			expectedSize: uint32(reflect.TypeOf(int(0)).Size()),
		},
		{
			name:         "size of string",
			elementType:  table.FactoryNewElementType[string](),
			expectedSize: uint32(reflect.TypeOf("").Size()),
		},
		{
			name:         "size of float64",
			elementType:  table.FactoryNewElementType[float64](),
			expectedSize: uint32(reflect.TypeOf(float64(0)).Size()),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.elementType.Size(); got != test.expectedSize {
				t.Errorf("elementType.Size() == %d, expected %d", got, test.expectedSize)
			}
		})
	}
}
