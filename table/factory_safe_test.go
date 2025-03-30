//go:build !unsafe

package table

import (
	"testing"
)

func TestFactorySafe_Init(t *testing.T) {
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
	_, ok = qTbl.rowCache.(safeCache)
	if !ok {
		t.Errorf("Expected rowCache to be of type safeCache, but got %T", qTbl.rowCache)
		return
	}
}
