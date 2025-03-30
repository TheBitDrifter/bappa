//go:build !schema_enabled

package table

func (Accessor[T]) assertedSchema(tbl *quickTable) *nilSchema {
	return tbl.schema.(*nilSchema)
}
