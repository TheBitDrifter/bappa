package table

type TableBuilder interface {
	WithSchema(Schema) TableBuilder
	WithEntryIndex(EntryIndex) TableBuilder
	WithElementTypes(...ElementType) TableBuilder
	WithEvents(TableEvents) TableBuilder
	Build() (Table, error)
}

type tableBuilder struct {
	schema       Schema
	entryIndex   EntryIndex
	elementTypes []ElementType
	events       TableEvents
}

func NewTableBuilder() TableBuilder {
	return &tableBuilder{}
}

func (b *tableBuilder) WithSchema(schema Schema) TableBuilder {
	b.schema = schema
	return b
}

func (b *tableBuilder) WithEntryIndex(entryIndex EntryIndex) TableBuilder {
	b.entryIndex = entryIndex
	return b
}

func (b *tableBuilder) WithElementTypes(types ...ElementType) TableBuilder {
	b.elementTypes = types
	return b
}

func (b *tableBuilder) WithEvents(events TableEvents) TableBuilder {
	b.events = events
	return b
}

func (b *tableBuilder) Build() (Table, error) {
	if b.schema == nil {
		b.schema = Factory.NewSchema()
	}
	if b.entryIndex == nil {
		b.entryIndex = Factory.NewEntryIndex()
	}

	table, err := Factory.NewTable(b.schema, b.entryIndex, b.elementTypes...)
	if err != nil {
		return nil, err
	}

	qTable, ok := table.(*quickTable)
	if b.events != nil && ok {
		qTable.events = b.events
	}
	return table, nil
}
