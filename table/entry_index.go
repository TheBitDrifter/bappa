package table

import (
	"fmt"

	numbers_util "github.com/TheBitDrifter/util/numbers"
)

var _ EntryIndex = &entryIndex{}

type entryIndex struct {
	currEntryID EntryID
	entries     []entry
	recyclable  []entry
	gen         int
}

func (ei *entryIndex) NewEntries(n, start int, tbl Table) ([]Entry, error) {
	if n <= 0 {
		return nil, BatchOperationError{Count: n}
	}
	amountRecyclable := min(len(ei.recyclable), n)
	newEntries := []Entry{}

	// First use recyclable entries
	for i := 0; i < amountRecyclable; i++ {
		e := entry{
			id:       ei.recyclable[i].ID(),
			recycled: ei.recyclable[i].Recycled() + 1,
			table:    tbl,
			index:    start + i,
		}
		globalIndex := int(e.ID()) - 1

		// Ensure the entries slice has enough capacity
		for globalIndex >= len(ei.entries) {
			ei.entries = append(ei.entries, entry{})
		}
		ei.entries[globalIndex] = e
		newEntries = append(newEntries, e)
	}

	// Remove used recyclable entries
	ei.recyclable = ei.recyclable[amountRecyclable:]

	// Create new entries for the remaining
	leftover := n - amountRecyclable
	for i := 0; i < leftover; i++ {
		ei.currEntryID++
		entry := entry{
			id:       ei.currEntryID,
			recycled: 0,
			table:    tbl,
			index:    start + i + amountRecyclable,
		}
		ei.entries = append(ei.entries, entry)
		newEntries = append(newEntries, entry)
	}

	ei.IncGen()
	return newEntries, nil
}

func (ei *entryIndex) Entry(i int) (Entry, error) {
	if i < 0 || i >= len(ei.entries) {
		return nil, AccessError{Index: i, UpperBound: len(ei.entries)}
	}
	entry := ei.entries[i]
	if entry.id == 0 {
		return nil, InvalidEntryAccessError{}
	}
	return entry, nil
}

func (ei *entryIndex) UpdateIndex(id EntryID, rowIndex int) error {
	entryIndex := int(id - 1)
	if entryIndex < 0 || entryIndex >= len(ei.entries) {
		return AccessError{Index: entryIndex, UpperBound: len(ei.entries) - 1}
	}
	e := ei.entries[entryIndex]
	newEntry := entry{
		id:       e.ID(),
		recycled: e.Recycled(),
		index:    rowIndex,
		table:    e.table,
	}
	ei.entries[entryIndex] = newEntry
	return nil
}

func (ei *entryIndex) RecycleEntries(ids ...EntryID) error {
	uniqueIDs := numbers_util.UniqueInts(entryIDs(ids).toInts())

	uniqCount := len(uniqueIDs)
	entriesCount := len(ei.entries)
	if uniqCount <= 0 || uniqCount >= entriesCount {
		return BatchDeletionError{Capacity: uniqCount, BatchOperationError: BatchOperationError{Count: uniqCount}}
	}

	for _, id := range uniqueIDs {
		entryID := EntryID(id)
		index := entryID - 1

		if ei.entries[index].ID() == 0 {
			continue
		}

		zeroEntry := entry{
			id:       0,
			recycled: ei.entries[index].Recycled(),
			index:    0,
		}
		recycledEntry := entry{
			id:       entryID,
			recycled: ei.entries[index].Recycled(),
			index:    0,
		}
		ei.recyclable = append(ei.recyclable, recycledEntry)
		ei.entries[index] = zeroEntry

	}

	ei.IncGen()
	return nil
}

func (ei *entryIndex) Reset() error {
	ei.entries = ei.entries[:0]
	ei.recyclable = ei.recyclable[:0]
	ei.currEntryID = 0
	ei.IncGen()
	return nil
}

func (ei *entryIndex) Entries() []Entry {
	entriesAsInterface := make([]Entry, len(ei.entries))
	for i, en := range ei.entries {
		entriesAsInterface[i] = en
	}
	return entriesAsInterface
}

func (ei *entryIndex) Recyclable() []Entry {
	recyclableEntriesAsInterface := make([]Entry, len(ei.recyclable))
	for i, en := range ei.recyclable {
		recyclableEntriesAsInterface[i] = en
	}
	return recyclableEntriesAsInterface
}

// Use with caution, primarily for deser
func (ei *entryIndex) ForceNewEntry(id int, recycledValue, tblIndex int, tbl Table) error {
	entryIDToForce := EntryID(id)
	if entryIDToForce == 0 {
		return fmt.Errorf("cannot force entry with ID 0") // ID 0 is reserved for invalid/empty
	}
	index := id - 1 // Target index for the new entry

	if index >= len(ei.entries) {
		requiredLen := index + 1

		// Create a new slice large enough
		newEntriesSlice := make([]entry, requiredLen)

		// Copy existing entries
		copy(newEntriesSlice, ei.entries)

		// Replace the old slice
		ei.entries = newEntriesSlice

		// Update highest ID tracker if needed
		if entryIDToForce > ei.currEntryID {
			ei.currEntryID = entryIDToForce
		}

	} else {
		// Slot already exists. Check if it's currently occupied by a *different* entry.
		existingEntry := ei.entries[index]

		// If the existing ID is not 0 (it's occupied) AND it's not the ID we intend to force
		if existingEntry.id != 0 && existingEntry.id != entryIDToForce {
			return fmt.Errorf("cannot force entry for ID %d: index %d is already occupied by valid ID %d", id, index, existingEntry.id)
		}
	}

	ei.entries[index] = entry{
		id:       entryIDToForce,
		table:    tbl,
		index:    tblIndex,
		recycled: recycledValue,
	}

	ei.IncGen()
	return nil
}

func (ei *entryIndex) Preallocate(capacity int) {
	if capacity > 0 && ei.entries == nil {
		ei.entries = make([]entry, 0, capacity)
	}
}

func (ei *entryIndex) Gen() int {
	return ei.gen
}

func (ei *entryIndex) IncGen() {
	ei.gen++
}
