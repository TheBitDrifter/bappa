package table

import (
	"fmt"
	"iter"
	"math"
	"reflect"
	"sort"
	"unsafe"

	"github.com/TheBitDrifter/mask"
	iter_util "github.com/TheBitDrifter/util/iter"
	numbers_util "github.com/TheBitDrifter/util/numbers"
)

var _ Table = &quickTable{}

type quickTable struct {
	schema       Schema
	entryIndex   EntryIndex
	rowCache     rowCache
	safeCache    []any
	unsafeCache  []unsafe.Pointer
	elementTypes []ElementType
	mask         mask.Mask
	entryIDs     []EntryID
	rows         []Row
	len          int
	cap          int
	events       TableEvents
}

func newTable(
	schema Schema, safeTable bool, entryIndex EntryIndex, elementTypes ...ElementType,
) (*quickTable, error) {
	if schema == nil {
		return nil, TableInstantiationNilSchemaError{}
	}
	if len(elementTypes) <= 0 {
		return nil, TableInstantiationNilElementTypesError{}
	}
	if Config.AutoElementTypeRegistrationTableCreation && !Config.SchemaLess() {
		schema.Register(elementTypes...)
	}
	tbl := &quickTable{
		schema:     schema,
		entryIndex: entryIndex,
	}
	tbl.elementTypes = elementTypes
	rowCount := schema.Registered()

	tbl.rows = make([]Row, rowCount)
	if safeTable {
		tbl.rowCache = safeCache(make([]any, rowCount))
		tbl.safeCache = tbl.rowCache.(safeCache)
	} else {
		tbl.rowCache = unsafeCache(make([]unsafe.Pointer, rowCount))
		tbl.unsafeCache = tbl.rowCache.(unsafeCache)
	}
	for _, elementType := range tbl.elementTypes {
		rowIndex := tbl.schema.RowIndexFor(elementType)
		tbl.rows[rowIndex] = newRow(elementType)

		bit := rowIndex
		tbl.mask.Mark(bit)
	}
	return tbl, nil
}

func (tbl *quickTable) Entry(tableIndex int) (Entry, error) {
	if tableIndex < 0 || tableIndex >= len(tbl.entryIDs) {
		return nil, AccessError{Index: tableIndex, UpperBound: len(tbl.entryIDs)}
	}
	entryIndex := int(tbl.entryIDs[tableIndex]) - 1
	return tbl.entryIndex.Entry(entryIndex)
}

func (tbl *quickTable) NewEntries(n int) ([]Entry, error) {
	if tbl.hasEvents() {
		if err := tbl.events.OnBeforeEntriesCreated(n); err != nil {
			return nil, err
		}
	}
	if n <= 0 {
		return nil, BatchOperationError{Count: n}
	}

	defer tbl.rowCache.cacheRows(tbl)

	prevTableLength := tbl.len
	err := tbl.ensureCapacity(n)
	if err != nil {
		return nil, err
	}

	err = tbl.addLen(n)
	if err != nil {
		return nil, err
	}

	entries, entryIndexError := tbl.entryIndex.NewEntries(n, prevTableLength, tbl)
	if tbl.hasEvents() {
		defer tbl.events.OnAfterEntriesCreated(entries)
	}

	if entryIndexError != nil {
		return nil, entryIndexError
	}

	for _, entry := range entries {
		tbl.entryIDs = append(tbl.entryIDs, entry.ID())
	}

	return entries, nil
}

func (tbl *quickTable) ForceNewEntry(id, recycled int) error {
	defer tbl.rowCache.cacheRows(tbl)

	err := tbl.ensureCapacity(1)
	if err != nil {
		return err
	}

	err = tbl.addLen(1)
	if err != nil {
		return err
	}

	err = tbl.entryIndex.ForceNewEntry(id, recycled, tbl.len-1, tbl)
	if err != nil {
		return err
	}

	newEn, err := tbl.entryIndex.Entry(id - 1)
	if err != nil {
		return err
	}

	tbl.entryIDs = append(tbl.entryIDs, newEn.ID())

	return nil
}

func (tbl *quickTable) DeleteEntries(indices ...int) ([]EntryID, error) {
	if tbl.hasEvents() {
		if err := tbl.events.OnBeforeEntriesDeleted(indices); err != nil {
			return nil, err
		}
	}
	n, err := tbl.prepForPopDeletion(indices...)
	if err != nil {
		return nil, err
	}
	deleted := tbl.popEntries(n, true)

	if tbl.hasEvents() {
		tbl.events.OnAfterEntriesDeleted(deleted)
	}
	return deleted, nil
}

func (tbl *quickTable) TransferEntries(other Table, indexes ...int) error {
	// Ensure unique indexes
	indexes = numbers_util.UniqueInts(indexes)
	n := len(indexes)

	// Validation checks
	if n <= 0 {
		return BatchOperationError{Count: n}
	}
	if n > tbl.len {
		return BatchDeletionError{Capacity: tbl.len, BatchOperationError: BatchOperationError{Count: n}}
	}

	// Tables must share the same entry index
	if entryIndexTracker[tbl] != entryIndexTracker[other] {
		return TransferEntryIndexMismatchError{}
	}

	// Get destination table
	otherTbl, ok := other.(*quickTable)
	if !ok {
		return fmt.Errorf("destination table must be a quickTable")
	}

	// Find shared element types
	sharedElementTypes := tbl.sharedElementTypesWith(other)

	// Ensure capacity in destination and collect entries
	if err := otherTbl.ensureCapacity(n); err != nil {
		return err
	}

	entryIDs := make([]EntryID, n)
	for i, idx := range indexes {
		if idx < 0 || idx >= tbl.len {
			return AccessError{Index: idx, UpperBound: tbl.len}
		}
		entryIDs[i] = tbl.entryIDs[idx]
	}

	// Create space in destination
	oldOtherLen := otherTbl.len
	otherTbl.len += n

	// Update all rows in destination
	for elementType := range otherTbl.ElementTypes() {
		row, err := otherTbl.Row(elementType)
		if err != nil {
			continue
		}
		row.setLen(otherTbl.len)
	}

	// Add entry IDs to destination
	otherTbl.entryIDs = append(otherTbl.entryIDs, entryIDs...)

	// Copy data
	for i, idx := range indexes {
		destIdx := oldOtherLen + i

		// Copy component data
		for _, elementType := range sharedElementTypes {
			srcRow, err := tbl.Row(elementType)
			if err != nil {
				continue
			}

			destRow, err := otherTbl.Row(elementType)
			if err != nil {
				continue
			}

			destRow.set(destIdx, srcRow.get(idx))
		}
	}

	// Update entries AFTER all data is transferred
	for i, entryID := range entryIDs {
		destIdx := oldOtherLen + i

		// Update entry index to point to new location
		if err := tbl.entryIndex.UpdateIndex(entryID, destIdx); err != nil {
			return err
		}

		// Update table reference in entry
		entryIdx := int(entryID) - 1
		ei := tbl.entryIndex.(*entryIndex)
		if entryIdx >= 0 && entryIdx < len(ei.entries) {
			e := ei.entries[entryIdx]
			e.table = other
			ei.entries[entryIdx] = e
		}
	}

	// Remove transferred entries
	//
	// First create a copy of the indexes to avoid modifying the input
	indexesCopy := make([]int, len(indexes))
	copy(indexesCopy, indexes)

	// Sort in descending order to avoid index shifts
	sort.Sort(sort.Reverse(sort.IntSlice(indexesCopy)))

	// Remove entities one by one from highest index to lowest
	for _, idx := range indexesCopy {
		// If this isn't the last entity, swap with the last one
		if idx < tbl.len-1 {
			lastIdx := tbl.len - 1

			// Swap component data
			for _, elementType := range tbl.elementTypes {
				row, err := tbl.Row(elementType)
				if err != nil {
					continue
				}

				// Get values
				if idx < reflect.Value(row).Len() && lastIdx < reflect.Value(row).Len() {
					temp := reflect.Value(row).Index(idx)
					tempCopy := reflect.New(temp.Type()).Elem()
					tempCopy.Set(temp)

					// Copy last to idx
					reflect.Value(row).Index(idx).Set(reflect.Value(row).Index(lastIdx))

					// Copy saved temp to last
					reflect.Value(row).Index(lastIdx).Set(tempCopy)
				}
			}

			// Swap entry IDs without updating entries
			lastEntryID := tbl.entryIDs[lastIdx]
			tbl.entryIDs[idx] = lastEntryID

			// Now update the entry index for the swapped entity
			tbl.entryIndex.UpdateIndex(lastEntryID, idx)
		}

		// Decrease table length
		tbl.len--
	}

	// Update row lengths
	for _, elementType := range tbl.elementTypes {
		row, err := tbl.Row(elementType)
		if err != nil {
			continue
		}
		row.setLen(tbl.len)
	}

	// Truncate entry IDs array
	tbl.entryIDs = tbl.entryIDs[:tbl.len]

	// Update row caches
	tbl.rowCache.cacheRows(tbl)
	otherTbl.rowCache.cacheRows(other)

	return nil
}

func (tbl *quickTable) Clear() error {
	defer tbl.rowCache.cacheRows(tbl)

	tbl.len = 0
	tbl.cap = 0

	for elementType := range tbl.ElementTypes() {
		row, err := tbl.Row(elementType)
		if err != nil {
			return err
		}
		row.setLen(tbl.len)
		row.setCap(tbl.cap)
	}
	return nil
}

func (tbl *quickTable) Length() int {
	return tbl.len
}

func (tbl *quickTable) ElementTypes() iter.Seq[ElementType] {
	return func(yield func(ElementType) bool) {
		for _, elementType := range tbl.elementTypes {
			if !yield(elementType) {
				return
			}
		}
	}
}

func (tbl *quickTable) Rows() iter.Seq2[int, Row] {
	return func(yield func(int, Row) bool) {
		for i, row := range tbl.rows {
			if row.CanAddr() && !yield(i, row) {
				return
			}
		}
	}
}

func (tbl *quickTable) RowCount() int {
	return len(tbl.elementTypes)
}

func (tbl *quickTable) Row(elementType ElementType) (Row, error) {
	if !tbl.Contains(elementType) {
		return Row{}, InvalidElementAccessError{elementType, iter_util.Collect(tbl.ElementTypes())}
	}
	rowIndex := tbl.schema.RowIndexFor(elementType)
	return tbl.rows[rowIndex], nil
}

func (tbl *quickTable) Contains(elementType ElementType) bool {
	if !tbl.schema.Contains(elementType) {
		return false
	}
	bit := tbl.schema.RowIndexFor(elementType)
	return tbl.mask.Contains(bit)
}

func (tbl *quickTable) ContainsAll(elementTypes ...ElementType) bool {
	if !tbl.schema.ContainsAll(elementTypes...) {
		return false
	}
	mask := mask.Mask{}
	for _, elementType := range elementTypes {
		bit := tbl.schema.RowIndexFor(elementType)
		mask.Mark(bit)
	}
	return tbl.mask.ContainsAll(mask)
}

func (tbl *quickTable) ContainsAny(elementTypes ...ElementType) bool {
	mask := mask.Mask{}
	for _, elementType := range elementTypes {
		if !tbl.schema.Contains(elementType) {
			continue
		}
		bit := tbl.schema.RowIndexFor(elementType)
		mask.Mark(bit)
	}
	return tbl.mask.ContainsAny(mask)
}

func (tbl *quickTable) ContainsNone(elementTypes ...ElementType) bool {
	msk := mask.Mask{}
	for _, elementType := range elementTypes {
		if !tbl.schema.Contains(elementType) {
			continue
		}
		bit := tbl.schema.RowIndexFor(elementType)
		msk.Mark(bit)
	}
	if msk.IsEmpty() {
		return true
	}
	return tbl.mask.ContainsNone(msk)
}

func (tbl *quickTable) Get(elementType ElementType, idx int) (reflect.Value, error) {
	row, err := tbl.Row(elementType)
	if err != nil {
		return reflect.Value{}, err
	}
	return row.get(idx), nil
}

func (tbl *quickTable) Set(elementType ElementType, re reflect.Value, idx int) error {
	row, err := tbl.Row(elementType)
	if err != nil {
		return err
	}
	row.set(idx, re)
	rowIdx := tbl.schema.RowIndexFor(elementType)
	tbl.rowCache.cacheRow(int(rowIdx), row)

	return nil
}

// Maskable
func (tbl *quickTable) Mask() mask.Mask {
	return tbl.mask
}

// --------- private helpers --------------
func (tbl *quickTable) addLen(n int) error {
	tbl.len += n
	for elementType := range tbl.ElementTypes() {
		row, err := tbl.Row(elementType)
		if err != nil {
			return InvalidElementAccessError{elementType, nil}
		}
		row.setLen(tbl.len)
	}
	return nil
}

func (tbl *quickTable) shrink() {
	if float64(tbl.cap) > float64(tbl.len)*1.2 {
		tbl.cap = tbl.len
	}
}

func (tbl *quickTable) ensureCapacity(n int) error {
	minCapacityRequired := tbl.len + n

	if tbl.cap >= minCapacityRequired {
		return nil
	}
	twentyPercentIncrease := float64(tbl.cap) * 1.2
	newCap := math.Max(float64(minCapacityRequired), twentyPercentIncrease)
	tbl.cap = int(newCap)

	for elementType := range tbl.ElementTypes() {
		tableRow, err := tbl.Row(elementType)
		if err != nil {
			return err
		}
		rowIndex := tbl.schema.RowIndexFor(elementType)
		rowType := tableRow.Type()

		newRowVal := reflect.New(rowType).Elem()
		newRowVal.Set(
			reflect.MakeSlice(rowType, tbl.len, tbl.cap),
		)

		reflect.Copy(newRowVal, reflect.Value(tableRow))
		tbl.rows[rowIndex] = Row(newRowVal)
	}
	return nil
}

func (tbl *quickTable) sharedElementTypesWith(other Table) []ElementType {
	sharedElements := make([]ElementType, 0)
	for otherElementType := range other.ElementTypes() {
		if tbl.Contains(otherElementType) {
			sharedElements = append(sharedElements, otherElementType)
		}
	}
	return sharedElements
}

func (tbl *quickTable) popEntries(n int, recycle bool) []EntryID {
	if n > tbl.len || n <= 0 {
		panic(fmt.Sprintf("cannot pop %d, table len: %d", n, tbl.len))
	}
	defer tbl.rowCache.cacheRows(tbl)
	entryIDsToDelete := tbl.entryIDs[tbl.len-n : tbl.len]
	tbl.addLen(-n)
	tbl.shrink()
	tbl.entryIDs = tbl.entryIDs[:tbl.len]
	if recycle {
		tbl.entryIndex.RecycleEntries(entryIDsToDelete...)
	}
	return entryIDsToDelete
}

func (tbl *quickTable) prepForPopDeletion(indices ...int) (int, error) {
	sortedUnique := numbers_util.UniqueInts(indices)
	n := len(sortedUnique)
	if n <= 0 {
		return 0, BatchOperationError{Count: n}
	}
	if n > tbl.len {
		return 0, BatchDeletionError{Capacity: tbl.len, BatchOperationError: BatchOperationError{Count: n}}
	}
	defer tbl.rowCache.cacheRows(tbl)

	// Validate indices first
	sortedDescending := numbers_util.DescendingInts(sortedUnique)
	for _, idx := range sortedDescending {
		if idx < 0 || idx >= tbl.len {
			return 0, AccessError{Index: idx, UpperBound: tbl.len}
		}
	}

	endPos := tbl.len - 1
	for _, idx := range sortedDescending {
		// Skip if:
		// 1. Index is already in position (idx == endPos)
		// 2. Index is already beyond our current swap position
		if idx >= endPos {
			endPos--
			continue
		}
		tbl.swapEntries(idx, endPos)
		endPos--
	}
	return n, nil
}

func (tbl *quickTable) swapEntries(i, j int) {
	if i < 0 || i >= tbl.len || j < 0 || j >= tbl.len {
		panic(fmt.Sprintf("swap columns bound error i: %d, j: %d, len: %d", i, j, tbl.len))
	}
	for _, row := range tbl.rows {
		if !row.CanAddr() {
			continue
		}
		copyI := row.get(i)
		row.set(i, row.get(j))

		copyAsElement := copyI
		row.set(j, copyAsElement)
	}
	tbl.entryIDs[i], tbl.entryIDs[j] = tbl.entryIDs[j], tbl.entryIDs[i]
	tbl.entryIndex.UpdateIndex(tbl.entryIDs[i], i)
	tbl.entryIndex.UpdateIndex(tbl.entryIDs[j], j)
}

func (tbl *quickTable) hasEvents() bool {
	return tbl.events != nil
}
