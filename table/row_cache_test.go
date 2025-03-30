package table

import (
	"reflect"
	"testing"

	testing_util "github.com/TheBitDrifter/util/testing"
)

func TestRowCacheAndTable_POINTERSYNC(t *testing.T) {
	intType := FactoryNewElementType[int]()
	boolType := FactoryNewElementType[bool]()

	tests := []struct {
		name            string
		t1              TableSetter
		t2              TableSetter
		amountToAdd     int
		indexesToDelete []int
	}{
		{
			name: "Basic int and bool types with various deletions",
			t1: TableSetter{
				{
					ElementType: intType,
					Values:      []any{1, 5, 9, 4, 6},
				}, {
					ElementType: boolType,
					Values:      []any{false, true, true, true},
				},
			},
			amountToAdd:     20,
			indexesToDelete: []int{0, 5, 4, 2, 12, 9, 18},
			t2: TableSetter{
				{
					ElementType: boolType,
					Values:      []any{true, true, false, true, false},
				},
				{
					ElementType: intType,
					Values:      []any{3, 4, 5, 999, 14280},
				},
			},
		},
		{
			name: "Multiple int and bool rows with mixed deletions",
			t1: TableSetter{
				{
					ElementType: intType,
					Values:      []any{10, 20, 30, 40},
				}, {
					ElementType: boolType,
					Values:      []any{true, false, true, false},
				},
			},
			amountToAdd:     5,
			indexesToDelete: []int{1, 3},
			t2: TableSetter{
				{
					ElementType: intType,
					Values:      []any{100, 200, 300},
				},
			},
		},
		{
			name: "Basic entries with bool deletion sync test",
			t1: TableSetter{
				{
					ElementType: intType,
					Values:      []any{7, 14, 21, 28},
				},
				{
					ElementType: boolType,
					Values:      []any{false, false, true, true},
				},
			},
			amountToAdd:     8,
			indexesToDelete: []int{0, 2, 3},
			t2: TableSetter{
				{
					ElementType: boolType,
					Values:      []any{true, false, true},
				},
			},
		},
		{
			name: "Alternating deletion and transfer test with int and bool rows",
			t1: TableSetter{
				{
					ElementType: intType,
					Values:      []any{2, 4, 6, 8, 10},
				},
				{
					ElementType: boolType,
					Values:      []any{true, false, false, true, true},
				},
			},
			amountToAdd:     15,
			indexesToDelete: []int{2, 4, 8, 10},
			t2: TableSetter{
				{
					ElementType: intType,
					Values:      []any{1, 3, 5, 7, 9},
				},
				{
					ElementType: boolType,
					Values:      []any{false, true, false, true, false},
				},
			},
		},
		{
			name: "Simple deletion and cache sync with int and bool types",
			t1: TableSetter{
				{
					ElementType: intType,
					Values:      []any{11, 22, 33, 44, 55},
				},
				{
					ElementType: boolType,
					Values:      []any{false, true, true, false, true},
				},
			},
			amountToAdd:     3,
			indexesToDelete: []int{1, 3},
			t2: TableSetter{
				{
					ElementType: boolType,
					Values:      []any{true, true, false, false, true},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := Factory.NewSchema()
			entryIndex := Factory.NewEntryIndex()
			tbl, err := tt.t1.Unpack(schema, entryIndex, t)
			testing_util.CheckError(t, tt.t1.Unpack, err)

			tbl.NewEntries(tt.amountToAdd)
			tbl.DeleteEntries(tt.indexesToDelete...)
			qTbl := tbl.(*quickTable)

			if tt.t2 != nil {
				tbl2, err := tt.t2.Unpack(schema, entryIndex, t)
				testing_util.CheckError(t, tt.t2.Unpack, err)
				tbl.TransferEntries(tbl2, 0, 2, 5)
			}
			for i, row := range tbl.Rows() {
				if Config.Unsafe() {
				} else {
					safeCache := qTbl.rowCache.(safeCache)
					testSafeCacheSync(safeCache, row, qTbl, i, t)
				}
			}
		})
	}
}

func testSafeCacheSync(sc safeCache, row Row, qTbl *quickTable, i int, t *testing.T) {
	safeCacheRow := sc[i]
	directSafeCacheRow := qTbl.safeCache[i]
	rowDef := reflect.Value(row).Interface()
	if !reflect.DeepEqual(rowDef, safeCacheRow) {
		t.Errorf("RowCache{} sync error, table row (%v) does not match table row cache(%v)", rowDef, safeCacheRow)
	}
	if !reflect.DeepEqual(safeCacheRow, directSafeCacheRow) {
		t.Errorf("RowCache{} sync error, table row cache(%v) does not match table direct cache(%v)", safeCacheRow, directSafeCacheRow)
	}
}

func testUnsafeCacheSync(sc unsafeCache, row Row, qTbl *quickTable, i int, t *testing.T) {
	unsafeCacheRow := sc[i]
	directSafeCacheRow := qTbl.safeCache[i]
	rowDef := reflect.Value(row).Interface()
	if !reflect.DeepEqual(rowDef, unsafeCacheRow) {
		t.Errorf("RowCache{} sync error, table row (%v) does not match table row cache(%v)", rowDef, unsafeCacheRow)
	}
	if !reflect.DeepEqual(unsafeCacheRow, directSafeCacheRow) {
		t.Errorf("RowCache{} sync error, table row cache(%v) does not match table direct cache(%v)", unsafeCacheRow, directSafeCacheRow)
	}
}
