//go:build schema_enabled

package table_benchmarks

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

func BenchmarkIterWarehouseGetWithLockedAccessor(b *testing.B) {
	b.StopTimer()
	ei := table.Factory.NewEntryIndex()
	schema := table.Factory.NewSchema()
	tbl, err := table.Factory.NewTable(schema, ei, etBool, etVec2, etString, etInt)
	if err != nil {
		b.Fatal(err)
	}
	tbl.NewEntries(1000)
	tLen := tbl.Length()
	vecAccessor := table.FactoryNewAccessor[vec2](etVec2)
	lVecAccessor := vecAccessor.NewLockedAccessor(schema)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		for k := 0; k < tLen; k++ {
			localV2 := lVecAccessor.Get(k, tbl)
			localV2.X += 1
			localV2.Y += 1
		}
	}
}
