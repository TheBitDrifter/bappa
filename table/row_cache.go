package table

import (
	"unsafe"
)

var (
	_ rowCache = safeCache{}
	_ rowCache = unsafeCache{}
)

type rowCache interface {
	cacheRows(Table)
	cacheRow(int, Row)
}
type (
	safeCache   []any
	unsafeCache []unsafe.Pointer
)

func (c safeCache) cacheRows(tbl Table) {
	for i, row := range tbl.Rows() {
		c[i] = row.Interface()
	}
}

func (c safeCache) cacheRow(i int, row Row) {
	c[i] = row.Interface()
}

func (c unsafeCache) cacheRows(tbl Table) {
	for i, row := range tbl.Rows() {
		c[i] = row.UnsafePointer()
	}
}

func (c unsafeCache) cacheRow(i int, row Row) {
	c[i] = row.UnsafePointer()
}
