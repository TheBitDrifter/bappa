//go:build unsafe

package table

import (
	"unsafe"
)

func init() {
	Config.buildTags = append(Config.buildTags, unsafeTag)
	Factory = initTableFactory()
}

func (accessor Accessor[T]) Get(idx int, table Table) *T {
	tbl := table.(*quickTable)
	cachedRow := tbl.unsafeCache[accessor.assertedSchema(tbl).RowIndexForID(accessor.elementTypeID)]
	var zero T

	offset := uint32(unsafe.Sizeof(zero)) * uint32(idx)
	return (*T)(unsafe.Add(cachedRow, offset))
}

func (lAccessor LockedAccessor[T]) Get(idx int, table Table) *T {
	tbl := table.(*quickTable)
	cachedRow := tbl.unsafeCache[lAccessor.rowIndex]
	var zero T
	offset := uint32(unsafe.Sizeof(zero)) * uint32(idx)
	return (*T)(unsafe.Add(cachedRow, offset))
}
