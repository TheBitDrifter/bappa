//go:build !schema_enabled

package table_test

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

type overrideID struct {
	table.ElementType
}

func (o overrideID) ID() table.ElementTypeID {
	return 999999
}

func TestNilSchema_Registered(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "small amount",
			count: 1,
		},
		{
			name:  "larger amount",
			count: table.Config.MaxElementCount() - 45,
		},
	}

	total := table.Stats.TotalElementTypes()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			total += test.count
			for i := 0; i < test.count; i++ {
				_ = table.FactoryNewElementType[int]()
			}
			schema := f.NewSchema()
			registered := schema.Registered()

			if registered != total {
				t.Errorf("Schema.Registered() == %d, expected %d", registered, total)
			}
		})
	}
}

func TestNilSchema_ContainsAll(t *testing.T) {
	tests := []struct {
		name             string
		elementTypes     []table.ElementType
		expectContainAll bool
	}{
		{
			name:             "single element contained within mask size",
			elementTypes:     []table.ElementType{table.FactoryNewElementType[int]()},
			expectContainAll: true,
		}, {
			name:             "multiple elements contained within mask size",
			elementTypes:     []table.ElementType{table.FactoryNewElementType[int](), table.FactoryNewElementType[string]()},
			expectContainAll: true,
		}, {
			name:             "element with ID outside mask size",
			elementTypes:     []table.ElementType{overrideID{table.FactoryNewElementType[int]()}},
			expectContainAll: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := f.NewSchema()
			contains := schema.ContainsAll(test.elementTypes...)
			if contains != test.expectContainAll {
				t.Errorf("Schema.Contains() == %v, expected %v", contains, test.expectContainAll)
			}
		})
	}
}

func TestNilSchema_RowIndexFor(t *testing.T) {
	intType := table.FactoryNewElementType[int]()
	floatType := table.FactoryNewElementType[float64]()
	tests := []struct {
		name          string
		elementType   table.ElementType
		expectedIndex uint32
	}{
		{
			name:          "row index for first element",
			elementType:   intType,
			expectedIndex: uint32(intType.ID()) - 1,
		},
		{
			name:          "row index for second element",
			elementType:   floatType,
			expectedIndex: uint32(floatType.ID()) - 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := f.NewSchema()
			rowIndex := schema.RowIndexFor(test.elementType)
			if rowIndex != test.expectedIndex {
				t.Errorf("Schema.RowIndexFor() == %d, expected %d", rowIndex, test.expectedIndex)
			}
		})
	}
}

func TestNilSchema_RowIndexForID(t *testing.T) {
	tests := []struct {
		name          string
		elementTypeID table.ElementTypeID
		expectedIndex uint32
	}{
		{
			name:          "row index for first ID",
			elementTypeID: 1,
			expectedIndex: 0,
		},
		{
			name:          "row index for second ID",
			elementTypeID: 2,
			expectedIndex: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := f.NewSchema()
			rowIndex := schema.RowIndexForID(test.elementTypeID)
			if rowIndex != test.expectedIndex {
				t.Errorf("Schema.RowIndexForID() == %d, expected %d", rowIndex, test.expectedIndex)
			}
		})
	}
}
