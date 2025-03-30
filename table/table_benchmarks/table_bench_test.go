package table_benchmarks

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

type vec2 struct{ X, Y int }

var (
	etInt    = table.FactoryNewElementType[int]()
	etString = table.FactoryNewElementType[string]()
	etBool   = table.FactoryNewElementType[bool]()
	etVec2   = table.FactoryNewElementType[vec2]()
)

func BenchmarkIterWarehouseGet(b *testing.B) {
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
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < tLen; k++ {
			localV2 := vecAccessor.Get(k, tbl)
			localV2.X += 1
			localV2.Y += 1
		}
	}
}
