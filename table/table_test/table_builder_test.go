package table_test

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

func TestTableBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func(table.TableBuilder) table.TableBuilder
		wantErr  bool
		validate func(*testing.T, table.Table)
	}{
		{
			name: "minimal valid table",
			build: func(b table.TableBuilder) table.TableBuilder {
				intType := table.FactoryNewElementType[int]()
				return b.WithElementTypes(intType)
			},
		},
		{
			name: "custom schema and entry index",
			build: func(b table.TableBuilder) table.TableBuilder {
				intType := table.FactoryNewElementType[int]()
				return b.
					WithSchema(table.Factory.NewSchema()).
					WithEntryIndex(table.Factory.NewEntryIndex()).
					WithElementTypes(intType)
			},
			validate: func(t *testing.T, tbl table.Table) {
				if tbl.Length() != 0 {
					t.Errorf("expected empty table, got length %d", tbl.Length())
				}
			},
		},
		{
			name: "with events",
			build: func(b table.TableBuilder) table.TableBuilder {
				intType := table.FactoryNewElementType[int]()
				events := &mockEvents{}
				return b.
					WithElementTypes(intType).
					WithEvents(events)
			},
		},
		{
			name: "no element types",
			build: func(b table.TableBuilder) table.TableBuilder {
				return b
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := table.NewTableBuilder()
			builder = tt.build(builder)
			tbl, err := builder.Build()

			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.validate != nil {
				tt.validate(t, tbl)
			}
		})
	}
}
