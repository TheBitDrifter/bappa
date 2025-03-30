package table_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/TheBitDrifter/table"
)

type mockEvents struct {
	onBeforeEntriesCreated func(count int) error
	onAfterEntriesCreated  func(entries []table.Entry)
	onBeforeEntriesDeleted func(indices []int) error
	onAfterEntriesDeleted  func(ids []table.EntryID)
}

func (m *mockEvents) OnBeforeEntriesCreated(count int) error {
	if m.onBeforeEntriesCreated != nil {
		return m.onBeforeEntriesCreated(count)
	}
	return nil
}

func (m *mockEvents) OnAfterEntriesCreated(entries []table.Entry) {
	if m.onAfterEntriesCreated != nil {
		m.onAfterEntriesCreated(entries)
	}
}

func (m *mockEvents) OnBeforeEntriesDeleted(indices []int) error {
	if m.onBeforeEntriesDeleted != nil {
		return m.onBeforeEntriesDeleted(indices)
	}
	return nil
}

func (m *mockEvents) OnAfterEntriesDeleted(ids []table.EntryID) {
	if m.onAfterEntriesDeleted != nil {
		m.onAfterEntriesDeleted(ids)
	}
}

func TestTableEvents(t *testing.T) {
	schema := table.Factory.NewSchema()
	entryIndex := table.Factory.NewEntryIndex()
	intType := table.FactoryNewElementType[int]()

	tests := []struct {
		name    string
		events  *mockEvents
		wantErr bool
	}{
		{
			name: "create entries - success",
			events: &mockEvents{
				onBeforeEntriesCreated: func(count int) error {
					if count != 5 {
						t.Errorf("expected count 5, got %d", count)
					}
					return nil
				},
				onAfterEntriesCreated: func(entries []table.Entry) {
					if len(entries) != 5 {
						t.Errorf("expected 5 entries, got %d", len(entries))
					}
				},
			},
		},
		{
			name: "create entries - error",
			events: &mockEvents{
				onBeforeEntriesCreated: func(count int) error {
					return errors.New("creation blocked")
				},
			},
			wantErr: true,
		},
		{
			name: "delete entries - success",
			events: &mockEvents{
				onBeforeEntriesDeleted: func(indices []int) error {
					if len(indices) != 2 {
						t.Errorf("expected 2 indices, got %d", len(indices))
					}
					return nil
				},
				onAfterEntriesDeleted: func(ids []table.EntryID) {
					if len(ids) != 2 {
						t.Errorf("expected 2 ids, got %d", len(ids))
					}
				},
			},
		},
		{
			name: "delete entries - error",
			events: &mockEvents{
				onBeforeEntriesDeleted: func(indices []int) error {
					return errors.New("deletion blocked")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := table.NewTableBuilder().WithSchema(schema).WithEntryIndex(entryIndex).WithElementTypes(intType)
			builder = builder.WithEvents(tt.events)
			tbl, err := builder.Build()
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(tt.name, "create") {
				_, err := tbl.NewEntries(5)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewEntries() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if strings.Contains(tt.name, "delete") {
				// Setup: Create entries first
				tbl.NewEntries(5)
				_, err := tbl.DeleteEntries(0, 1)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteEntries() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
