package table

import (
	"iter"
	"reflect"

	"github.com/TheBitDrifter/mask"
)

type (
	EntryID       uint32
	ElementTypeID uint32
	RowIndex      uint32
)

type Entry interface {
	ID() EntryID
	Recycled() int
	Index() int
	Table() Table
}

type EntryIndex interface {
	Entry(idx int) (Entry, error)
	NewEntries(count, previousTableLength int, tbl Table) ([]Entry, error)
	UpdateIndex(EntryID, int) error
	RecycleEntries(...EntryID) error
	Reset() error
	Entries() []Entry
	Recyclable() []Entry
}

type Schema interface {
	Register(...ElementType)
	Registered() int
	Contains(ElementType) bool
	ContainsAll(...ElementType) bool
	RowIndexFor(ElementType) uint32
	RowIndexForID(ElementTypeID) uint32
}

type ElementType interface {
	ID() ElementTypeID
	Type() reflect.Type
	Size() uint32
}

type (
	Row reflect.Value
)

type Table interface {
	TableReader
	TableWriter
	TableQuerier
	TableIterator
}

type TableReader interface {
	Entry(int) (Entry, error)
	Get(ElementType, int) (reflect.Value, error)
	Row(ElementType) (Row, error)
	Length() int
	RowCount() int
}

type TableWriter interface {
	Set(ElementType, reflect.Value, int) error
	NewEntries(int) ([]Entry, error)
	DeleteEntries(...int) ([]EntryID, error)
	TransferEntries(Table, ...int) error
	Clear() error
}

type TableQuerier interface {
	Contains(ElementType) bool
	ContainsAll(...ElementType) bool
	ContainsAny(...ElementType) bool
	ContainsNone(...ElementType) bool
}

type TableIterator interface {
	Rows() iter.Seq2[int, Row]
	ElementTypes() iter.Seq[ElementType]
}

type TableEvents interface {
	OnBeforeEntriesCreated(count int) error
	OnAfterEntriesCreated(entries []Entry)
	OnBeforeEntriesDeleted(indices []int) error
	OnAfterEntriesDeleted(ids []EntryID)
}

// Warning: internal Dependencies abound!
type Accessor[T any] struct {
	elementTypeID ElementTypeID
}

type LockedAccessor[T any] struct {
	schema   *quickSchema
	rowIndex uint32
}

type MaskComparer interface {
	ContainsAll(mask.Mask) bool
	ContainsAny(mask.Mask) bool
	ContainsNone(mask.Mask) bool
}
