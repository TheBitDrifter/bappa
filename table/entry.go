package table

var _ Entry = entry{}

type entry struct {
	id       EntryID
	recycled int
	index    int
	table    Table
}

func (e entry) ID() EntryID {
	return e.id
}

func (e entry) Valid() bool {
	return e.id != 0
}

func (e entry) Recycled() int {
	return e.recycled
}

func (e entry) Index() int {
	return e.index
}

func (e entry) Table() Table {
	return e.table
}

type entryIDs []EntryID

func (eid entryIDs) toInts() []int {
	integers := make([]int, len(eid))
	for i, id := range eid {
		integers[i] = int(id)
	}
	return integers
}
