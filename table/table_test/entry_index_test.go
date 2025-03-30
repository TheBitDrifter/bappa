package table_test

import (
	"testing"

	"github.com/TheBitDrifter/table"
	"github.com/TheBitDrifter/util/slices"
	testing_util "github.com/TheBitDrifter/util/testing"
)

func TestEntryIndex_NewEntriesAndRecycle(t *testing.T) {
	tests := []struct {
		name             string
		newEntriesCount  int
		expectedNewCount int
		expectError      error
	}{
		{
			name:             "add 5 new entries",
			newEntriesCount:  5,
			expectedNewCount: 5,
		},
		{
			name:             "add 10 new entries",
			newEntriesCount:  10,
			expectedNewCount: 10,
		},
		{
			name:             "add 0 entries",
			newEntriesCount:  0,
			expectedNewCount: 0,
			expectError:      table.BatchOperationError{Count: 0},
		},
		{
			name:             "add negative entries",
			newEntriesCount:  -1,
			expectedNewCount: 0,
			expectError:      table.BatchOperationError{Count: -1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := f.NewEntryIndex()
			schema := f.NewSchema()
			tbl, err := f.NewTable(schema, ei, blankElementType)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ei.NewEntries(tt.newEntriesCount, tbl.Length(), tbl)
			testing_util.CheckError(t, ei.NewEntries, err, tt.expectError)
			funcName := testing_util.FuncName(ei.NewEntries, t)
			entryCount := len(ei.Entries())

			if entryCount != tt.expectedNewCount {
				t.Errorf("%s = %d, expected %d", funcName, len(ei.Entries()), tt.expectedNewCount)
			}

			deleted := []table.EntryID{}
			deletedMap := map[table.EntryID]bool{}

			// Inverted to match the order of expected recycling later on...
			inverted := slices.ReverseSliceCopy(ei.Entries())

			for i, entry := range inverted {
				if i%2 == 0 {
					id := entry.ID()
					ei.RecycleEntries(id)
					deleted = append(deleted, id)
					deletedMap[id] = true
				}
			}
			for _, entry := range ei.Entries() {
				id := entry.ID()
				if deletedMap[id] {
					t.Errorf("DeleteEntries() did not clear entries properly, found ID %d", id)
				}
			}
			for _, dID := range deleted {
				entries, err := ei.NewEntries(1, tbl.Length(), tbl)
				testing_util.CheckError(t, ei.NewEntries, err)
				funcName := testing_util.FuncName(ei.NewEntries, t)
				entry := entries[0]
				eID := entry.ID()
				if eID != dID {
					t.Errorf("%s sentryID not recycled, got %d, expected %d", funcName, eID, dID)
				}
			}
		})
	}
}

func TestEntryIndex_UpdateIndex(t *testing.T) {
	tests := []struct {
		name          string
		entriesCount  int
		newIndex      int
		updateIndexes []int
		expectError   error
	}{
		{
			name:          "out of bound index",
			entriesCount:  10,
			newIndex:      -5,
			updateIndexes: []int{1, 3, 5, 20},
			expectError:   table.AccessError{Index: 19, UpperBound: 9},
		},
		{
			name:          "valid index update",
			entriesCount:  10,
			newIndex:      3,
			updateIndexes: []int{0, 2, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := f.NewSchema()
			ei := f.NewEntryIndex()
			tbl, err := f.NewTable(s, ei, blankElementType)
			if err != nil {
				t.Fatal(err)
			}
			ei.NewEntries(tt.entriesCount, tbl.Length(), tbl)

			for _, idx := range tt.updateIndexes {
				// We do not want to index ei.Entries() if invalid idx.
				if idx >= len(ei.Entries()) {
					err := ei.UpdateIndex(table.EntryID(idx), tt.newIndex)
					testing_util.CheckError(t, ei.UpdateIndex, err, tt.expectError)
					continue
				}
				entryID := ei.Entries()[idx].ID()
				err := ei.UpdateIndex(entryID, tt.newIndex)
				testing_util.CheckError(t, ei.UpdateIndex, err, tt.expectError)
			}
			for _, idx := range tt.updateIndexes {
				if idx >= len(ei.Entries()) {
					continue
				}
				entry := ei.Entries()[idx]
				funcName := testing_util.FuncName(ei.UpdateIndex, t)
				if entry.Index() != tt.newIndex {
					t.Errorf("%s failed, expected newIndex to be %d, got %d for entry %d", funcName, tt.newIndex, entry.Index(), entry.ID())
				}
			}
		})
	}
}

func TestEntryIndex_Reset(t *testing.T) {
	tests := []struct {
		name          string
		entriesCount  int
		deleteIDs     []int
		expectedNewID int
		expectedCount int
	}{
		{
			name:          "reset after some entries and deletions",
			entriesCount:  5,
			deleteIDs:     []int{2, 4},
			expectedNewID: 1,
			expectedCount: 3,
		},
		{
			name:          "reset with no entries",
			entriesCount:  0,
			deleteIDs:     []int{},
			expectedNewID: 1,
			expectedCount: 3,
		},
		{
			name:          "reset without deletions",
			entriesCount:  5,
			deleteIDs:     []int{},
			expectedNewID: 1,
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := f.NewEntryIndex()
			s := f.NewSchema()
			tbl, err := f.NewTable(s, ei, blankElementType)
			if err != nil {
				t.Fatal(err)
			}
			ei.NewEntries(tt.entriesCount, tbl.Length(), tbl)

			if len(tt.deleteIDs) > 0 {
				ids := make([]table.EntryID, len(tt.deleteIDs))
				for i, id := range tt.deleteIDs {
					ids[i] = table.EntryID(id)
				}
				err := ei.RecycleEntries(ids...)
				testing_util.CheckError(t, ei.RecycleEntries, err)
			}
			ei.Reset()

			if len(ei.Entries()) != 0 {
				t.Errorf("Expected 0 entries after Reset, got %d", len(ei.Entries()))
			}
			if len(ei.Recyclable()) != 0 {
				t.Errorf("Expected 0 recyclable entries after Reset, got %d", len(ei.Recyclable()))
			}
			newEntries, err := ei.NewEntries(tt.expectedCount, tbl.Length(), tbl)
			testing_util.CheckError(t, ei.NewEntries, err)

			if len(newEntries) != tt.expectedCount {
				t.Errorf("Expected %d new entries after Reset, got %d", tt.expectedCount, len(newEntries))
			}
			if int(newEntries[0].ID()) != tt.expectedNewID {
				t.Errorf("Expected first new entry ID to be %d after Reset, got %d", tt.expectedNewID, newEntries[0].ID())
			}
		})
	}
}

func TestEntryIndex_Recyclable(t *testing.T) {
	const totalEntries = 5
	const toDelete1, toDelete2 = 2, 4
	const newEntriesCount = 3

	ei := f.NewEntryIndex()
	s := f.NewSchema()
	tbl, err := f.NewTable(s, ei, blankElementType)
	if err != nil {
		t.Fatal(err)
	}
	ei.NewEntries(totalEntries, tbl.Length(), tbl)
	ei.RecycleEntries(toDelete1, toDelete2)
	recyclable := ei.Recyclable()
	expectedRecyclableCount := 2
	expectedRecyclableIDs := []table.EntryID{toDelete1, toDelete2}

	if len(recyclable) != expectedRecyclableCount {
		t.Errorf("Expected %d recyclable entries, got %d", expectedRecyclableCount, len(recyclable))
	}
	for i, entry := range recyclable {
		if entry.ID() != expectedRecyclableIDs[i] {
			t.Errorf("Expected recyclable entry ID %d, got %d", expectedRecyclableIDs[i], entry.ID())
		}
	}
	newEntries, err := ei.NewEntries(newEntriesCount, tbl.Length(), tbl)
	testing_util.CheckError(t, ei.NewEntries, err, nil)
	funcName := testing_util.FuncName(ei.NewEntries, t)
	reusedIDs := expectedRecyclableIDs

	for i := 0; i < len(reusedIDs); i++ {
		if newEntries[i].ID() != reusedIDs[i] {
			t.Errorf("%s failed, expected new entry to reuse recyclable ID %d, got %d", funcName, reusedIDs[i], newEntries[i].ID())
		}
	}
	if len(ei.Recyclable()) != 0 {
		t.Errorf("%s failed, expected 0 recyclable entries after reuse, got %d", funcName, len(ei.Recyclable()))
	}
}
