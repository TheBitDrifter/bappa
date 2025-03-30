//go:build !unsafe

package table_test

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

func TestAccessor_Get(t *testing.T) {
	schema := f.NewSchema()
	entryIndex := f.NewEntryIndex()
	stringElementType := table.FactoryNewElementType[string]()
	intElementType := table.FactoryNewElementType[int]()
	intAccessor := table.FactoryNewAccessor[int](intElementType)
	stringAccessor := table.FactoryNewAccessor[string](stringElementType)
	mockTableA, err := f.NewTable(schema, entryIndex, stringElementType, intElementType)
	if err != nil {
		t.Fatal(err)
	}
	mockTableB, err := f.NewTable(schema, entryIndex, intElementType, stringElementType)
	if err != nil {
		t.Fatal(err)
	}
	mockTableC, err := f.NewTable(schema, entryIndex, intElementType)
	if err != nil {
		t.Fatal(err)
	}

	entryCount := 1000
	mockTableA.NewEntries(entryCount)
	mockTableB.NewEntries(entryCount)
	mockTableC.NewEntries(entryCount)

	tests := []struct {
		name      string
		idx       int
		table     table.Table
		accessor  interface{}
		wantValue interface{}
		wantPanic bool
	}{
		{
			name:      "Set and Get first element from mockTableA",
			idx:       0,
			table:     mockTableA,
			accessor:  intAccessor,
			wantValue: 42,
			wantPanic: false,
		},
		{
			name:      "Set and Get second element from mockTableA",
			idx:       1,
			table:     mockTableA,
			accessor:  intAccessor,
			wantValue: 84,
			wantPanic: false,
		},
		{
			name:      "Invalid index for mockTableA",
			idx:       entryCount,
			table:     mockTableA,
			accessor:  intAccessor,
			wantValue: 0,
			wantPanic: true,
		},
		{
			name:      "Set and Get first element from mockTableB",
			idx:       0,
			table:     mockTableB,
			accessor:  stringAccessor,
			wantValue: "foo",
			wantPanic: false,
		},
		{
			name:      "Set and Get second element from mockTableB",
			idx:       1,
			table:     mockTableB,
			accessor:  intAccessor,
			wantValue: 84,
			wantPanic: false,
		},
		{
			name:      "Attempt to access non-existent string element in mockTableC",
			idx:       0,
			table:     mockTableC,
			accessor:  stringAccessor,
			wantValue: nil,
			wantPanic: true,
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
			case table.Accessor[string]:
				if !tt.table.Contains(stringElementType) {
					continue
				}
				got := accessor.Get(tt.idx, tt.table)
				if got != nil {
					*got = tt.wantValue.(string)
				}
			}
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Accessor.Get(%d) did not panic, expected a panic", tt.idx)
					}
				}()
			}
			switch accessor := tt.accessor.(type) {
			case table.Accessor[int]:
				got := accessor.Get(tt.idx, tt.table)
				if got == nil {
					t.Errorf("Accessor.Get(%d) == nil, expected a value", tt.idx)
				} else if *got != tt.wantValue {
					t.Errorf("Accessor.Get(%d) = %v, expected %v", tt.idx, *got, tt.wantValue)
				}
			case table.Accessor[string]:
				got := accessor.Get(tt.idx, tt.table)
				if got == nil {
					t.Errorf("Accessor.Get(%d) == nil, expected a value", tt.idx)
				} else if *got != tt.wantValue {
					t.Errorf("Accessor.Get(%d) = %v, expected %v", tt.idx, *got, tt.wantValue)
				}
			}
		})
	}
}
