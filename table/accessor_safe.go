//go:build !unsafe

package table

func (accessor Accessor[T]) Get(idx int, table Table) *T {
	tbl := table.(*quickTable)
	cachedRow := tbl.safeCache[accessor.assertedSchema(tbl).RowIndexForID(accessor.elementTypeID)].([]T)
	return &cachedRow[idx]
}

func (lAccessor LockedAccessor[T]) Get(idx int, table Table) *T {
	tbl := table.(*quickTable)
	cachedRows := tbl.rowCache.(safeCache)
	row := cachedRows[lAccessor.rowIndex].([]T)
	return &row[idx]
}
