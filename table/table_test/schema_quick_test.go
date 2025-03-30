//go:build schema_enabled

package table_test

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

func TestQuickSchema_Registered(t *testing.T) {
	floatType := table.FactoryNewElementType[float64]()
	tests := []struct {
		name           string
		elementTypes   []table.ElementType
		wantRegistered int
	}{
		{
			name:           "register two distinct aliases of the same type",
			elementTypes:   []table.ElementType{table.FactoryNewElementType[int](), table.FactoryNewElementType[int]()},
			wantRegistered: 2,
		},
		{
			name:           "register different element types",
			elementTypes:   []table.ElementType{table.FactoryNewElementType[int](), table.FactoryNewElementType[string]()},
			wantRegistered: 2,
		},
		{
			name:           "register same element twice",
			elementTypes:   []table.ElementType{floatType, floatType},
			wantRegistered: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := f.NewSchema()

			for _, elType := range test.elementTypes {
				schema.Register(elType)
			}
			registered := schema.Registered()
			if registered != test.wantRegistered {
				t.Errorf("Schema.Registered() == %d, expected %d", registered, test.wantRegistered)
			}
		})
	}
}

func TestQuickSchema_Contains(t *testing.T) {
	intType := table.FactoryNewElementType[int]()
	stringType := table.FactoryNewElementType[string]()
	boolType := table.FactoryNewElementType[bool]()
	tests := []struct {
		name             string
		registerType     table.ElementType
		checkTypes       []table.ElementType
		expectContainAll bool
	}{
		{
			name:             "single type - does not contain an element created earlier",
			registerType:     stringType,
			checkTypes:       []table.ElementType{intType},
			expectContainAll: false,
		}, {
			name:             "single type - does not contain an element created later",
			registerType:     intType,
			checkTypes:       []table.ElementType{stringType},
			expectContainAll: false,
		}, {
			name:             "multiple types - none contained",
			registerType:     boolType,
			checkTypes:       []table.ElementType{intType, stringType},
			expectContainAll: false,
		}, {
			name:             "multiple types - one contained one not",
			registerType:     intType,
			checkTypes:       []table.ElementType{intType, stringType},
			expectContainAll: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := f.NewSchema()
			schema.Register(test.registerType)
			contains := schema.ContainsAll(test.checkTypes...)
			if contains != test.expectContainAll {
				t.Errorf("Schema.ContainsAll() == %v, expected %v", contains, test.expectContainAll)
			}
		})
	}
}

func TestQuickSchema_RowIndexFor(t *testing.T) {
	intType := table.FactoryNewElementType[int]()
	stringType := table.FactoryNewElementType[string]()
	var maxUint32 uint32 = 4294967295
	tests := []struct {
		name          string
		registerType  table.ElementType
		checkType     table.ElementType
		expectedIndex uint32
		expectPanic   bool
	}{
		{
			name:          "row index for unregistered element created earlier then the registered element",
			registerType:  stringType,
			checkType:     intType,
			expectedIndex: maxUint32,
		},
		{
			name:         "row index for unregistered element after the registered element",
			registerType: intType,
			checkType:    stringType,
			expectPanic:  true,
		},
		{
			name:         "row index for registered element",
			registerType: stringType,
			checkType:    stringType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := f.NewSchema()
			schema.Register(test.registerType)

			if test.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but got none")
					}
				}()
			}
			rowIndex := schema.RowIndexFor(test.checkType)
			if rowIndex != test.expectedIndex && !test.expectPanic {
				t.Errorf("Schema.RowIndexFor() == %d, expected %d", rowIndex, test.expectedIndex)
			}
		})
	}
}
