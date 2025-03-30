package table_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/TheBitDrifter/table"
	numbers_util "github.com/TheBitDrifter/util/numbers"
	testing_util "github.com/TheBitDrifter/util/testing"
)

var (
	f                = table.Factory
	blankElementType = table.FactoryNewElementType[int]()
)

func TestTable_NewEntriesAndDeleteEntriesAndRecycledEntries(t *testing.T) {
	tests := []struct {
		name        string
		entryCount  int
		wantLength  int
		expectError error
	}{
		{
			name:       "create 5 entries",
			entryCount: 5,
			wantLength: 5,
		},
		{
			name:        "create 0 entries",
			entryCount:  0,
			wantLength:  0,
			expectError: table.BatchOperationError{},
		},
		{
			name:        "create negative entries",
			entryCount:  -3,
			wantLength:  0,
			expectError: table.BatchOperationError{Count: -3},
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
			entries, err := tbl.NewEntries(tt.entryCount)
			testing_util.CheckError(t, tbl.NewEntries, err, tt.expectError)
			funcName := testing_util.FuncName(tbl.NewEntries, t)

			defer func() {
				validateIndexes(t, ei)
			}()
			if len(entries) != tt.wantLength {
				t.Errorf("%s entries length == %d, want %d", funcName, len(entries), tt.wantLength)
			}
			if tbl.Length() != tt.wantLength {
				t.Errorf("%s table length == %d, want %d", funcName, tbl.Length(), tt.wantLength)
			}
			deleted := map[table.EntryID]bool{}
			for i := 0; i < tbl.Length(); i++ {
				if i%2 == 0 {
					continue
				}
				deletedIDs, err := tbl.DeleteEntries(i)
				testing_util.CheckError(t, tbl.DeleteEntries, err, tt.expectError)

				for _, id := range deletedIDs {
					deleted[id] = true
				}
				funcName = testing_util.FuncName(tbl.DeleteEntries, t)
				if tt.wantLength-len(deleted) != tbl.Length() {
					t.Errorf("%s table length == %d wanted %d", funcName, tbl.Length(), tt.wantLength-len(deleted))
				}
				for i := 0; i < tbl.Length(); i++ {
					entry, err := tbl.Entry(i)
					testing_util.CheckError(t, tbl.DeleteEntries, err, tt.expectError)
					id := entry.ID()

					if deleted[id] {
						t.Errorf("%s failed, expected EntryID %d to be removed", funcName, id)
					}
				}
			}
		})
	}
}

func TestTable_DeleteEntries(t *testing.T) {
	tests := []struct {
		name          string
		count         int
		deleteIndexes []int
		expectError   []error
	}{
		{
			name:          "delete valid 5",
			count:         5,
			deleteIndexes: []int{0, 1, 2, 3, 4},
		},
		{
			name:          "delete dupes with variadic",
			count:         5,
			deleteIndexes: []int{4, 4, 4, 1, 2},
		},
		{
			name:          "delete invalid index",
			count:         5,
			deleteIndexes: []int{3, 8},
			expectError:   []error{table.AccessError{Index: 8, UpperBound: 5}},
		},
		{
			name:          "delete from empty table",
			count:         0,
			deleteIndexes: []int{0},
			expectError: []error{
				table.BatchDeletionError{Capacity: 0, BatchOperationError: table.BatchOperationError{Count: 1}},
				table.BatchOperationError{Count: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := f.NewSchema()
			ei := f.NewEntryIndex()
			et := table.FactoryNewElementType[int]()
			tbl, err := f.NewTable(s, ei, et)
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				validateIndexes(t, ei)
			}()
			_, err = tbl.NewEntries(tt.count)
			if testing_util.CheckError(t, tbl.NewEntries, err, tt.expectError...) {
				return
			}
			_, err = tbl.DeleteEntries(tt.deleteIndexes...)
			if testing_util.CheckError(t, tbl.DeleteEntries, err, tt.expectError...) {
				return
			}

			funcName := testing_util.FuncName(tbl.DeleteEntries, t)
			deletedCount := len(numbers_util.UniqueInts(tt.deleteIndexes))
			tblLength := tbl.Length()
			expectedLen := tt.count - deletedCount
			if tblLength != expectedLen {
				t.Errorf("%s failed, expected length %d, got %d", funcName, expectedLen, tblLength)
			}
		})
	}
}

func TestTable_TransferEntries(t *testing.T) {
	type foo struct {
		foo string
	}
	fooType := table.FactoryNewElementType[foo]()
	intType := table.FactoryNewElementType[int]()
	stringType := table.FactoryNewElementType[string]()
	boolType := table.FactoryNewElementType[bool]()
	floatType := table.FactoryNewElementType[float64]()

	tests := []struct {
		name             string
		t1, t2           table.TableSetter
		transferIndexes  []int
		uniqueSchemas    bool
		uniqueEntryIndex bool
		expectError      error
	}{
		{
			name: "valid transfer, unique schemas",
			t1: table.TableSetter{{
				ElementType: intType,
				Values:      []any{50, 63},
			}},
			t2: table.TableSetter{{
				ElementType: intType,
				Values:      []any{30, 50, 12, 99},
			}},
			transferIndexes:  []int{0, 1},
			uniqueSchemas:    true,
			uniqueEntryIndex: false,
		},
		{
			name: "valid transfer, shared schema",
			t1: table.TableSetter{{
				ElementType: intType,
				Values:      []any{50, 63},
			}},
			t2: table.TableSetter{{
				ElementType: intType,
				Values:      []any{30, 50, 12, 99},
			}},
			transferIndexes:  []int{0, 1},
			uniqueSchemas:    false,
			uniqueEntryIndex: false,
		},
		{
			name: "invalid transfer unique schemas, bounds error",
			t1: table.TableSetter{{
				ElementType: intType,
				Values:      []any{50, 63},
			}},
			t2: table.TableSetter{{
				ElementType: intType,
				Values:      []any{30, 50, 12, 99},
			}},
			transferIndexes:  []int{1, 90},
			uniqueSchemas:    false,
			uniqueEntryIndex: false,
			expectError:      table.AccessError{Index: 90, UpperBound: 2},
		},
		{
			name: "valid transfer, shared schema, no common elements",
			t1: table.TableSetter{{
				ElementType: intType,
				Values:      []any{50, 63, 40, 10},
			}},
			t2: table.TableSetter{{
				ElementType: stringType,
				Values:      []any{"foo", "bar", "shiso"},
			}},
			transferIndexes:  []int{0, 3},
			uniqueSchemas:    false,
			uniqueEntryIndex: false,
		},
		{
			name: "valid transfer, unique schema, complex(ish) element combination",
			t1: table.TableSetter{
				{
					ElementType: intType,
					Values:      []any{50, 63, 40, 10},
				},
				{
					ElementType: boolType,
					Values:      []any{true, false, false, true},
				},
			},
			t2: table.TableSetter{
				{
					ElementType: stringType,
					Values:      []any{"foo", "bar", "shiso"},
				},
				{
					ElementType: boolType,
					Values:      []any{false, true, false, true, true},
				},
				{
					ElementType: floatType,
					Values:      []any{0.4, 12.212, 304.3},
				},
				{
					ElementType: fooType,
					Values:      []any{foo{foo: "bar"}, foo{foo: "baz"}},
				},
			},
			transferIndexes:  []int{0, 3},
			uniqueSchemas:    false,
			uniqueEntryIndex: false,
		},
		{
			name: "almost valid transfer, entryIndex mismatch",
			t1: table.TableSetter{{
				ElementType: intType,
				Values:      []any{50, 63, 40, 10},
			}},
			t2: table.TableSetter{{
				ElementType: stringType,
				Values:      []any{"foo", "bar", "shiso"},
			}},
			transferIndexes:  []int{0, 3},
			uniqueSchemas:    false,
			uniqueEntryIndex: true,
			expectError:      table.TransferEntryIndexMismatchError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup schemas and entry indices
			s1 := f.NewSchema()
			s2 := s1
			if tt.uniqueSchemas {
				s2 = f.NewSchema()
			}
			ei := f.NewEntryIndex()
			ei2 := ei
			if tt.uniqueEntryIndex {
				ei2 = f.NewEntryIndex()
			}
			defer func() {
				validateIndexes(t, ei)
				validateIndexes(t, ei2)
			}()
			// Create and populate tables
			tbl1, err := tt.t1.Unpack(s1, ei, t)
			testing_util.CheckError(t, tt.t1.Unpack, err, nil)
			tbl2, err := tt.t2.Unpack(s2, ei2, t)
			testing_util.CheckError(t, tt.t2.Unpack, err, nil)

			// Collect expected entry IDs
			expectedIDs := []table.EntryID{}
			for i := 0; i < tbl2.Length(); i++ {
				entry, err := tbl2.Entry(i)
				testing_util.CheckError(t, tbl2.Entry, err, tt.expectError)
				id := entry.ID()
				expectedIDs = append(expectedIDs, id)
			}
			for _, idx := range tt.transferIndexes {
				entry, err := tbl1.Entry(idx)
				if testing_util.CheckError(t, tbl1.Entry, err, tt.expectError) {
					continue
				}
				id := entry.ID()
				expectedIDs = append(expectedIDs, id)
			}
			// Calculate expected lengths after transfer
			expectedNewLen1 := tbl1.Length() - len(tt.transferIndexes)
			expectedNewLen2 := tbl2.Length() + len(tt.transferIndexes)

			// Perform the transfer
			err = tbl1.TransferEntries(tbl2, tt.transferIndexes...)
			if testing_util.CheckError(t, tbl1.TransferEntries, err, tt.expectError) {
				return
			}
			funcName := testing_util.FuncName(tbl1.TransferEntries, t)

			// Verify table lengths
			if expectedNewLen1 != tbl1.Length() {
				t.Errorf("%s failed, expected table one length %d got, %d",
					funcName, expectedNewLen1, tbl1.Length())
			}
			if expectedNewLen2 != tbl2.Length() {
				t.Errorf("%s failed expected table two length %d, got %d",
					funcName, expectedNewLen2, tbl2.Length())
			}
			// Calculate and verify expected layout
			expectedLayout := mergeTableLayouts(tt.t1, tt.t2, tt.transferIndexes)

			// Verify values for each element type
			for elementType, expectedValues := range expectedLayout {
				// Get actual values from table 2
				for i := 0; i < tbl2.Length(); i++ {
					_, err := tbl2.Entry(i)
					if err != nil {
						t.Errorf("%s failed to get entry at index %d: %v",
							funcName, i, err)
						continue
					}
					// Get actual values
					actualVal, err := tbl2.Get(elementType, i)
					expectedVal := expectedValues[i]
					expectedReflectVal := reflect.ValueOf(expectedVal)

					// Compare the values
					if reflect.DeepEqual(expectedReflectVal, actualVal) {
						t.Errorf("%s failed at index %d for element type %v: expected %v, got %v",
							funcName, i, elementType, expectedVal, actualVal)
					}
				}
			}
			// Verify entry IDs and order
			for i := 0; i < tbl2.Length(); i++ {
				entry, err := tbl2.Entry(i)
				if err != nil {
					t.Errorf("%s failed to get entry at index %d: %v",
						funcName, i, err)
					continue
				}
				if entry.ID() != expectedIDs[i] {
					t.Errorf("%s failed: entry ID mismatch at index %d: expected %v, got %v",
						funcName, i, expectedIDs[i], entry.ID())
				}
			}
		})
	}
}

func TestTable_Clear(t *testing.T) {
	s := f.NewSchema()
	ei := f.NewEntryIndex()
	intType := table.FactoryNewElementType[int]()
	stringType := table.FactoryNewElementType[string]()
	tbl, err := f.NewTable(s, ei, intType, stringType)
	if err != nil {
		t.Fatal(err)
	}
	tbl.NewEntries(5)
	for i := 0; i < 5; i++ {
		tbl.Set(intType, reflect.ValueOf(i), i)
		tbl.Set(stringType, reflect.ValueOf(fmt.Sprintf("test%d", i)), i)
	}
	tbl.Clear()
	if tbl.Length() != 0 {
		t.Errorf("Expected table length to be 0 after Clear, got %d", tbl.Length())
	}
	rows := tbl.Rows()
	for i, row := range rows {
		rowValue := reflect.Value(row)
		if rowValue.Len() != 0 {
			t.Errorf("Expected row %d to be empty after Clear, got length %d", i, rowValue.Len())
		}
	}
}

func TestTable_ElementTypes(t *testing.T) {
	s := f.NewSchema()
	ei := f.NewEntryIndex()
	intType := table.FactoryNewElementType[int]()
	stringType := table.FactoryNewElementType[string]()
	tbl, err := f.NewTable(s, ei, intType, stringType)
	if err != nil {
		t.Fatal(err)
	}
	expectedTypes := []table.ElementType{intType, stringType}
	var gotTypes []table.ElementType
	for elementType := range tbl.ElementTypes() {
		gotTypes = append(gotTypes, elementType)
	}
	if !reflect.DeepEqual(expectedTypes, gotTypes) {
		t.Errorf("Expected element types %v, got %v", expectedTypes, gotTypes)
	}
}

func TestTable_Rows(t *testing.T) {
	s := f.NewSchema()
	ei := f.NewEntryIndex()
	intType := table.FactoryNewElementType[int]()
	stringType := table.FactoryNewElementType[string]()
	tbl, err := f.NewTable(s, ei, intType, stringType)
	if err != nil {
		t.Fatal(err)
	}
	tbl.NewEntries(3)
	for i := 0; i < 3; i++ {
		tbl.Set(intType, reflect.ValueOf(i), i)
		tbl.Set(stringType, reflect.ValueOf(fmt.Sprintf("test%d", i)), i)
	}
	rowCount := 0
	for i, row := range tbl.Rows() {
		rowCount++
		rowLen := reflect.Value(row).Len()
		if i == 0 && rowLen != 3 {
			t.Errorf("Expected row length to be 3, got %d", row.Type().Len())
		}
	}
	if rowCount != 2 {
		t.Errorf("Expected 2 rows, got %d", rowCount)
	}
}

func TestTable_RowCount(t *testing.T) {
	s := f.NewSchema()
	ei := f.NewEntryIndex()
	intType := table.FactoryNewElementType[int]()
	stringType := table.FactoryNewElementType[string]()
	tbl, err := f.NewTable(s, ei, intType, stringType)
	if err != nil {
		t.Fatal(err)
	}
	if tbl.RowCount() != 2 {
		t.Errorf("Expected row count to be 2, got %d", tbl.RowCount())
	}
	floatType := table.FactoryNewElementType[float64]()
	tbl, err = f.NewTable(s, ei, intType, stringType, floatType)
	if err != nil {
		t.Fatal(err)
	}
	if tbl.RowCount() != 3 {
		t.Errorf("Expected row count to be 3, got %d", tbl.RowCount())
	}
}

func TestTable_Contains(t *testing.T) {
	s := f.NewSchema()
	ei := f.NewEntryIndex()
	intType := table.FactoryNewElementType[int]()
	stringType := table.FactoryNewElementType[string]()
	floatType := table.FactoryNewElementType[float64]()
	tbl, err := f.NewTable(s, ei, intType, stringType)
	if err != nil {
		t.Fatal(err)
	}
	singleTests := []struct {
		name         string
		elemType     table.ElementType
		expectResult bool
	}{
		{
			name:         "Contains - has int type",
			elemType:     intType,
			expectResult: true,
		}, {
			name:         "Contains - has string type",
			elemType:     stringType,
			expectResult: true,
		}, {
			name:         "Contains - does not have float type",
			elemType:     floatType,
			expectResult: false,
		},
	}
	for _, test := range singleTests {
		t.Run(test.name, func(t *testing.T) {
			if result := tbl.Contains(test.elemType); result != test.expectResult {
				t.Errorf("%s() == %v, expected %v", test.name, result, test.expectResult)
			}
		})
	}
	variadicTests := []struct {
		name         string
		fn           func(...table.ElementType) bool
		types        []table.ElementType
		expectResult bool
	}{
		{
			name:         "ContainsAll - has both types",
			fn:           tbl.ContainsAll,
			types:        []table.ElementType{intType, stringType},
			expectResult: true,
		}, {
			name:         "ContainsAll - missing one type",
			fn:           tbl.ContainsAll,
			types:        []table.ElementType{intType, floatType},
			expectResult: false,
		}, {
			name:         "ContainsAny - has at least one type",
			fn:           tbl.ContainsAny,
			types:        []table.ElementType{intType, floatType},
			expectResult: true,
		}, {
			name:         "ContainsAny - has none of types",
			fn:           tbl.ContainsAny,
			types:        []table.ElementType{floatType},
			expectResult: false,
		}, {
			name:         "ContainsNone - has none of types",
			fn:           tbl.ContainsNone,
			types:        []table.ElementType{floatType},
			expectResult: true,
		}, {
			name:         "ContainsNone - has some of types",
			fn:           tbl.ContainsNone,
			types:        []table.ElementType{intType, floatType},
			expectResult: false,
		},
	}
	for _, test := range variadicTests {
		t.Run(test.name, func(t *testing.T) {
			if result := test.fn(test.types...); result != test.expectResult {
				t.Errorf("%s() == %v, expected %v", test.name, result, test.expectResult)
			}
		})
	}
}

func mergeTableLayouts(tableOne, tableTwo table.TableSetter, transferIndexes []int) map[table.ElementType][]any {
	result := make(map[table.ElementType][]any)

	longestRowLength := 0
	for _, setter2 := range tableTwo {
		currentLen := len(setter2.Values)
		if currentLen > longestRowLength {
			longestRowLength = currentLen
		}
	}
	for _, setter2 := range tableTwo {
		paddedValues := make([]any, longestRowLength)
		copy(paddedValues, setter2.Values)
		result[setter2.ElementType] = paddedValues
	}
	fromLoop := false
	for _, setter1 := range tableOne {
		if _, exists := result[setter1.ElementType]; !exists {
			continue
		}
		fromLoop = true
		for _, idx := range transferIndexes {
			if idx < len(setter1.Values) {
				result[setter1.ElementType] = append(result[setter1.ElementType], setter1.Values[idx])
			}
		}
	}
	if !fromLoop {
		longestRowLength += len(transferIndexes)
	} else {
		longestRowLength = 0
		for _, expectedValues := range result {
			currentLen := len(expectedValues)
			if currentLen > longestRowLength {
				longestRowLength = currentLen
			}
		}
	}
	for k, expectedValues := range result {
		for i := len(expectedValues); i < longestRowLength; i++ {
			result[k] = append(result[k], nil)
		}
	}
	return result
}

func validateIndexes(t *testing.T, ei table.EntryIndex) {
	for i := 0; ; i++ {
		entry, err := ei.Entry(i)
		if err != nil {
			break
		}
		if entry.ID() != table.EntryID(i+1) && entry.ID() != 0 {
			t.Errorf("Expected EntryID at index %d to be %d, but got %d", i, i+1, entry.ID())
		}
	}
}
