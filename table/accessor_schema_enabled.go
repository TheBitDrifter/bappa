//go:build schema_enabled

package table

func init() {
	Config.buildTags = append(Config.buildTags, schemaEnabledTag)
	Factory = initTableFactory()
}

func (Accessor[T]) assertedSchema(tbl *quickTable) *quickSchema {
	return tbl.schema.(*quickSchema)
}

func (accessor Accessor[T]) NewLockedAccessor(schema Schema) LockedAccessor[T] {
	quickSchema := schema.(*quickSchema)
	schemaIndex := schema.RowIndexForID(accessor.elementTypeID)
	return LockedAccessor[T]{quickSchema, schemaIndex}
}
