//go:build unsafe

package table_test

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

func TestAccessor_Get(t *testing.T) {
	schema := f.NewSchema()
	entryIndex := f.NewEntryIndex()
	intElementType := table.FactoryNewElementType[int]()
	intAccessor := table.FactoryNewAccessor[int](intElementType)
	mockTable, err := f.NewTable(schema, entryIndex, intElementType)
	if err != nil {
		t.Fatal(err)
	}
	entryCount := 1000
	mockTable.NewEntries(entryCount)

	tests := []struct {
		name      string
		idx       int
		table     table.Table
		accessor  interface{}
		wantValue interface{}
		wantZero  bool
	}{
		{
			name:      "Set and Get first element",
			idx:       0,
			table:     mockTable,
			accessor:  intAccessor,
			wantValue: 42,
			wantZero:  false,
		},
		{
			name:      "Set and Get second element",
			idx:       1,
			table:     mockTable,
			accessor:  intAccessor,
			wantValue: 84,
			wantZero:  false,
		},
		{
			name:      "Invalid index should return zero value",
			idx:       entryCount,
			table:     mockTable,
			accessor:  intAccessor,
			wantValue: 0,
			wantZero:  true,
		},
	}
	for _, tt := range tests {
		if tt.idx < entryCount {
			switch accessor := tt.accessor.(type) {
			case table.Accessor[int]:
				got := accessor.Get(tt.idx, tt.table)
				if got != nil {
					*got = tt.wantValue.(int)
				}
			}
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch accessor := tt.accessor.(type) {
			case table.Accessor[int]:
				got := accessor.Get(tt.idx, tt.table)
				if got == nil {
					if !tt.wantZero {
						t.Errorf("Accessor.Get(%d) == nil, expected a value", tt.idx)
					}
				} else if *got != tt.wantValue && !tt.wantZero {
					t.Errorf("Accessor.Get(%d) = %v, expected %v", tt.idx, *got, tt.wantValue)
				} else if tt.wantZero && *got != tt.wantValue {
					t.Errorf("Accessor.Get(%d) = %v, expected zero value", tt.idx, *got)
				}
			}
		})
	}
}
