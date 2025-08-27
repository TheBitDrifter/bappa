package table

type TableBuilder interface {
	WithSchema(Schema) TableBuilder
	WithEntryIndex(EntryIndex) TableBuilder
	WithElementTypes(...ElementType) TableBuilder
	WithEvents(TableEvents) TableBuilder
	WithInitialCapacity(int) TableBuilder
	Build() (Table, error)
}

type tableBuilder struct {
	schema          Schema
	entryIndex      EntryIndex
	elementTypes    []ElementType
	events          TableEvents
	initialCapacity int
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

	// Pass the initialCapacity to the table constructor
	table, err := Factory.NewTable(b.schema, b.entryIndex, b.initialCapacity, b.elementTypes...)
	if err != nil {
		return nil, err
	}

	qTable, ok := table.(*quickTable)
	if b.events != nil && ok {
		qTable.events = b.events
	}
	return table, nil
}

func (b *tableBuilder) WithInitialCapacity(cap int) TableBuilder {
	b.initialCapacity = cap
	return b
}
