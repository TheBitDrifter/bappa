package table

import (
	"reflect"
)

var (
	buildTags                      = []string{}
	Factory           TableFactory = initTableFactory()
	entryIndexTracker              = map[Table]EntryIndex{}
)

func initTableFactory() TableFactory {
	var tableB iTableFactory = safeTableFactory{}
	if Config.Unsafe() {
		tableB = unsafeTableFactory{}
	}
	var schemaB iSchemaFactory = nilSchemaFactory{}
	if !Config.SchemaLess() {
		schemaB = quickSchemaFactory{}
	}
	return tableFactory{
		iTableFactory:  tableB,
		iSchemaFactory: schemaB,
	}
}

type TableFactory interface {
	iTableFactory
	iSchemaFactory
	NewEntryIndex() EntryIndex
}

type iTableFactory interface {
	NewTable(Schema, EntryIndex, ...ElementType) (Table, error)
}

type iSchemaFactory interface {
	NewSchema() Schema
}

type (
	safeTableFactory   struct{}
	unsafeTableFactory struct{}
	nilSchemaFactory   struct{}
	quickSchemaFactory struct{}

	tableFactory struct {
		iTableFactory
		iSchemaFactory
	}
)

func (tableFactory) NewEntryIndex() EntryIndex {
	return &entryIndex{}
}

func (factory safeTableFactory) NewTable(schema Schema, entryIndex EntryIndex, elementTypes ...ElementType) (Table, error) {
	tbl, err := newTable(schema, true, entryIndex, elementTypes...)
	if err != nil {
		return nil, err
	}
	entryIndexTracker[tbl] = entryIndex
	return tbl, nil
}

func (factory unsafeTableFactory) NewTable(schema Schema, entryIndex EntryIndex, elementTypes ...ElementType) (Table, error) {
	tbl, err := newTable(schema, false, entryIndex, elementTypes...)
	if err != nil {
		return nil, err
	}
	entryIndexTracker[tbl] = entryIndex
	return tbl, err
}

func (factory quickSchemaFactory) NewSchema() Schema {
	return newQuickSchema()
}

func (factory nilSchemaFactory) NewSchema() Schema {
	return newNilSchema()
}

func FactoryNewElementType[T any]() ElementType {
	return newElementType[T]()
}

// Warning: Internal dependency abound!
func FactoryNewAccessor[T any](elementType ElementType) Accessor[T] {
	var zero T
	tType := reflect.TypeOf(zero)

	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	if elementType.Type() != tType {
		panic("mismatch")
	}
	return Accessor[T]{elementType.ID()}
}
