//go:build !schema_enabled

package table

import (
	"testing"
)

func TestFactorySchemaLess_Init(t *testing.T) {
	f := initTableFactory()
	s := f.NewSchema()
	ei := f.NewEntryIndex()
	et := FactoryNewElementType[int]()
	tbl, err := f.NewTable(s, ei, et)
	if err != nil {
		t.Fatal(err)
	}
	qTbl, ok := tbl.(*quickTable)
	if !ok {
		t.Errorf("Expected tbl to be of type *quickTable, but got %T", tbl)
		return
	}
	_, ok = qTbl.schema.(*nilSchema)
	if !ok {
		t.Errorf("Expected schema to be of type *nilSchema, but got %T", qTbl.schema)
		return
	}
}
