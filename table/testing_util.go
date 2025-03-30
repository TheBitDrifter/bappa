package table

import (
	"fmt"
	"reflect"
	"testing"
)

type TableSetter []struct {
	ElementType
	Values []any
}

func (ts TableSetter) Unpack(s Schema, ei EntryIndex, t *testing.T) (Table, error) {
	if len(ts) == 0 {
		t.Errorf("error: empty table setter")
	}
	typeMap := make(map[ElementType]struct{})
	longestRowLength := 0

	for _, setter := range ts {
		typeMap[setter.ElementType] = struct{}{}
		if len(setter.Values) > longestRowLength {
			longestRowLength = len(setter.Values)
		}
	}
	elTypes := make([]ElementType, 0, len(typeMap))
	for elType := range typeMap {
		elTypes = append(elTypes, elType)
	}
	tbl, err := Factory.NewTable(s, ei, elTypes...)
	if err != nil {
		return nil, fmt.Errorf("failed to create table entries: %w", err)
	}
	entries, err := tbl.NewEntries(longestRowLength)
	if err != nil {
		return nil, fmt.Errorf("failed to create table entries: %w", err)
	}
	for _, ts := range ts {
		for i := range entries {
			if i >= len(ts.Values) {
				break
			}
			val := ts.Values[i]
			reflectVal := reflect.ValueOf(val)
			if err := tbl.Set(ts.ElementType, reflectVal, i); err != nil {
				return nil, fmt.Errorf("failed to set value at index %d: %w", i, err)
			}
		}
	}
	return tbl, nil
}
