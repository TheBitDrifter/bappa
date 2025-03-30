package table

func (accessor Accessor[T]) Check(table Table) bool {
	tbl := table.(*quickTable)
	bit := accessor.assertedSchema(tbl).RowIndexForID(accessor.elementTypeID)
	return tbl.mask.Contains(bit)
}

func (lAccessor LockedAccessor[T]) Check(table Table) bool {
	tbl := table.(*quickTable)
	return tbl.mask.Contains(lAccessor.rowIndex)
}
