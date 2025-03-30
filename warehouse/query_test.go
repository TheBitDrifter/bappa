package warehouse

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

// TestQueryFiltering tests the basic query filtering capabilities
func TestQueryFiltering(t *testing.T) {
	// Create components once to reuse
	posComp := FactoryNewComponent[Position]()
	velComp := FactoryNewComponent[Velocity]()
	healthComp := FactoryNewComponent[Health]()

	type entitySetup struct {
		components []Component
		count      int
	}

	tests := []struct {
		name            string
		entitySetups    []entitySetup
		queryType       string // "and", "or", "not", "complex"
		queryComponents []Component
		expectedMatches int
	}{
		{
			name: "And query matches exact",
			entitySetups: []entitySetup{
				{[]Component{posComp, velComp}, 5},
				{[]Component{posComp}, 10},
				{[]Component{velComp}, 15},
			},
			queryType:       "and",
			queryComponents: []Component{posComp, velComp},
			expectedMatches: 5,
		},
		{
			name: "Or query matches either",
			entitySetups: []entitySetup{
				{[]Component{posComp, velComp}, 5},
				{[]Component{posComp}, 10},
				{[]Component{velComp}, 15},
			},
			queryType:       "or",
			queryComponents: []Component{posComp, velComp},
			expectedMatches: 30, // 5 + 10 + 15
		},
		{
			name: "Not query excludes",
			entitySetups: []entitySetup{
				{[]Component{posComp, velComp}, 5},
				{[]Component{posComp}, 10},
				{[]Component{velComp}, 15},
				{[]Component{healthComp}, 20},
			},
			queryType:       "not",
			queryComponents: []Component{velComp},
			expectedMatches: 30, // 10 + 20
		},
		{
			name: "Complex query",
			entitySetups: []entitySetup{
				{[]Component{posComp, velComp, healthComp}, 5},
				{[]Component{posComp, velComp}, 10},
				{[]Component{posComp, healthComp}, 15},
				{[]Component{velComp, healthComp}, 20},
				{[]Component{posComp}, 25},
				{[]Component{velComp}, 30},
				{[]Component{healthComp}, 35},
			},
			queryType:       "complex",
			queryComponents: []Component{posComp, velComp, healthComp},
			expectedMatches: 30, // (P AND V) OR (P AND H) = 10 + 15 + 5 (counted once)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup storage and create entities
			schema := table.Factory.NewSchema()
			storage := Factory.NewStorage(schema)

			for _, setup := range tt.entitySetups {
				_, err := storage.NewEntities(setup.count, setup.components...)
				if err != nil {
					t.Fatalf("Failed to create entities: %v", err)
				}
			}

			// Create query based on test case
			query := Factory.NewQuery()
			var queryNode QueryNode

			switch tt.queryType {
			case "and":
				// Convert []Component to []interface{} by looping
				interfaceComponents := make([]interface{}, len(tt.queryComponents))
				for i, comp := range tt.queryComponents {
					interfaceComponents[i] = comp
				}
				queryNode = query.And(interfaceComponents...)
			case "or":
				interfaceComponents := make([]interface{}, len(tt.queryComponents))
				for i, comp := range tt.queryComponents {
					interfaceComponents[i] = comp
				}
				queryNode = query.Or(interfaceComponents...)
			case "not":
				interfaceComponents := make([]interface{}, len(tt.queryComponents))
				for i, comp := range tt.queryComponents {
					interfaceComponents[i] = comp
				}
				queryNode = query.Not(interfaceComponents...)
			case "complex":
				// Test a more complex query: (Position AND Velocity) OR (Position AND Health)
				andQuery1 := query.And(posComp, velComp)
				andQuery2 := query.And(posComp, healthComp)
				queryNode = query.Or(andQuery1, andQuery2)
			}

			// Create cursor and count matches
			cursor := Factory.NewCursor(queryNode, storage)
			matchCount := 0
			for range cursor.Next() {
				matchCount++
			}

			// Verify match count
			if matchCount != tt.expectedMatches {
				t.Errorf("Query matched %d entities, want %d", matchCount, tt.expectedMatches)
			}
		})
	}
}

// TestQueryWithCursor tests the cursor-based entity iteration
func TestQueryWithCursor(t *testing.T) {
	// Create components once to reuse
	posComp := FactoryNewComponent[Position]()
	velComp := FactoryNewComponent[Velocity]()
	healthComp := FactoryNewComponent[Health]()

	tests := []struct {
		name            string
		entityTypes     [][]Component
		queryComponents []Component
		expectedCount   int
	}{
		{
			name: "Query with position",
			entityTypes: [][]Component{
				{posComp},
				{posComp, velComp},
				{velComp},
			},
			queryComponents: []Component{posComp},
			expectedCount:   20, // 10 + 10
		},
		{
			name: "Query with position and velocity",
			entityTypes: [][]Component{
				{posComp},
				{posComp, velComp},
				{velComp},
			},
			queryComponents: []Component{posComp, velComp},
			expectedCount:   10,
		},
		{
			name: "Query with no matches",
			entityTypes: [][]Component{
				{posComp},
				{velComp},
			},
			queryComponents: []Component{healthComp},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup storage and create entities
			schema := table.Factory.NewSchema()
			storage := Factory.NewStorage(schema)

			for _, componentSet := range tt.entityTypes {
				_, err := storage.NewEntities(10, componentSet...)
				if err != nil {
					t.Fatalf("Failed to create entities: %v", err)
				}
			}

			// Create query
			query := Factory.NewQuery()
			// Convert []Component to []interface{}
			interfaceComponents := make([]interface{}, len(tt.queryComponents))
			for i, comp := range tt.queryComponents {
				interfaceComponents[i] = comp
			}
			queryNode := query.And(interfaceComponents...)

			// Method 1: Use cursor directly
			cursor := Factory.NewCursor(queryNode, storage)
			count1 := 0
			for range cursor.Next() {
				count1++
			}

			// Method 2: Use cursor's TotalMatched
			cursor = Factory.NewCursor(queryNode, storage)
			count2 := cursor.TotalMatched()

			// Verify counts match each other and expected
			if count1 != count2 {
				t.Errorf("Cursor counts inconsistent: %d vs %d", count1, count2)
			}

			if count1 != tt.expectedCount {
				t.Errorf("Query matched %d entities, want %d", count1, tt.expectedCount)
			}
		})
	}
}

// TestQueryComponentAccess tests accessing component data through queries
func TestQueryComponentAccess(t *testing.T) {
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create components
	posComp := FactoryNewComponent[Position]()
	velComp := FactoryNewComponent[Velocity]()

	// Create entities with initial values
	for i := 0; i < 10; i++ {
		// Create entity with position
		pos := Position{X: float64(i), Y: float64(i * 2)}
		entities, err := storage.NewEntities(1, posComp)
		if err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}
		entity := entities[0]

		// Set position value
		posPtr := posComp.GetFromEntity(entity)
		*posPtr = pos

		// Add velocity component with value
		vel := Velocity{X: float64(i) * 0.1, Y: float64(i) * 0.2}
		err = entity.AddComponentWithValue(velComp, vel)
		if err != nil {
			t.Fatalf("Failed to add velocity: %v", err)
		}
	}

	// Create query to match all entities with position and velocity
	query := Factory.NewQuery()
	// Use []interface{} with the components
	queryNode := query.And(interface{}(posComp), interface{}(velComp))
	cursor := Factory.NewCursor(queryNode, storage)

	// Iterate and update positions based on velocities
	for range cursor.Next() {
		// Get entity
		entity, err := cursor.CurrentEntity()
		if err != nil {
			t.Fatalf("Failed to get current entity: %v", err)
		}

		// Get component values
		pos := posComp.GetFromEntity(entity)
		vel := velComp.GetFromEntity(entity)

		// Update position based on velocity
		pos.X += vel.X
		pos.Y += vel.Y
	}

	// Create a new cursor to iterate and check updated values
	cursor = Factory.NewCursor(queryNode, storage)
	for range cursor.Next() {
		// Get entity
		entity, err := cursor.CurrentEntity()
		if err != nil {
			t.Fatalf("Failed to get current entity: %v", err)
		}

		// Get component values
		pos := posComp.GetFromEntity(entity)
		vel := velComp.GetFromEntity(entity)

		// Get initial values based on pattern
		expectedX := pos.X - vel.X
		expectedY := pos.Y - vel.Y

		// Check if the values match the pattern (i, i*2)
		// Allow for small floating point errors
		if !almostEqual(expectedX, vel.X*10, 0.0001) || !almostEqual(expectedY/2, vel.X*10, 0.0001) {
			t.Errorf("Position {%v, %v} with velocity {%v, %v} doesn't match expected pattern",
				pos.X-vel.X, pos.Y-vel.Y, vel.X, vel.Y)
		}
	}
}

// Helper function for float comparisons
func almostEqual(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}
