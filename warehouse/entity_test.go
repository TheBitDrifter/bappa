package warehouse

import (
	"log"
	"testing"

	"github.com/TheBitDrifter/table"
)

// Test component types
type Position struct {
	X, Y float64
}

type Velocity struct {
	X, Y float64
}

type Health struct {
	Current, Max int
}

func TestEntityCreation(t *testing.T) {
	// Create component instances once to reuse
	posComp := FactoryNewComponent[Position]()
	velComp := FactoryNewComponent[Velocity]()
	healthComp := FactoryNewComponent[Health]()

	tests := []struct {
		name           string
		componentTypes []Component
		entityCount    int
		wantError      bool
	}{
		{"Empty entity", []Component{}, 1, true},
		{"Single component", []Component{posComp}, 10, false},
		{"Multiple components", []Component{posComp, velComp}, 5, false},
		{"Large batch", []Component{posComp, velComp, healthComp}, 1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := table.Factory.NewSchema()
			storage := Factory.NewStorage(schema)

			entities, err := storage.NewEntities(tt.entityCount, tt.componentTypes...)

			if (err != nil) != tt.wantError {
				t.Errorf("NewEntities() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if len(entities) != tt.entityCount {
					t.Errorf("Created %d entities, want %d", len(entities), tt.entityCount)
				}

				// Check that all entities have valid IDs
				for i, entity := range entities {
					if !entity.Valid() {
						t.Errorf("Entity %d is invalid", i)
					}
				}

				// Verify components on first entity
				if len(entities) > 0 {
					components := entities[0].Components()
					if len(components) != len(tt.componentTypes) {
						t.Errorf("Entity has %d components, want %d", len(components), len(tt.componentTypes))
					}
				}
			}
		})
	}
}

func TestComponentAddRemove(t *testing.T) {
	// Create components once
	posComp := FactoryNewComponent[Position]()
	velComp := FactoryNewComponent[Velocity]()
	healthComp := FactoryNewComponent[Health]()

	tests := []struct {
		name              string
		initialComponents []Component
		addComponents     []Component
		removeComponents  []Component
		wantError         bool
		finalCount        int
	}{
		{
			name:              "Add component",
			initialComponents: []Component{posComp},
			addComponents:     []Component{velComp},
			removeComponents:  nil,
			wantError:         false,
			finalCount:        2,
		},
		{
			name:              "Remove component",
			initialComponents: []Component{posComp, velComp},
			addComponents:     nil,
			removeComponents:  []Component{velComp},
			wantError:         false,
			finalCount:        1,
		},
		{
			name:              "Add and remove",
			initialComponents: []Component{posComp},
			addComponents:     []Component{velComp, healthComp},
			removeComponents:  []Component{posComp},
			wantError:         false,
			finalCount:        2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := table.Factory.NewSchema()
			storage := Factory.NewStorage(schema)

			entities, err := storage.NewEntities(1, tt.initialComponents...)
			if err != nil {
				t.Fatalf("Failed to create entity: %v", err)
			}

			entity := entities[0]

			// Add components
			for _, comp := range tt.addComponents {
				err = entity.AddComponent(comp)
				if (err != nil) != tt.wantError {
					t.Errorf("AddComponent() error = %v, wantError %v", err, tt.wantError)
				}
			}

			// Remove components
			for _, comp := range tt.removeComponents {
				err = entity.RemoveComponent(comp)
				if (err != nil) != tt.wantError {
					t.Errorf("RemoveComponent() error = %v, wantError %v", err, tt.wantError)
				}
			}

			// Check final component count
			components := entity.Components()
			if len(components) != tt.finalCount {
				log.Println(entity.ComponentsAsString())
				t.Errorf("Entity has %d components, want %d", len(components), tt.finalCount)
			}
		})
	}
}

func TestComponentValues(t *testing.T) {
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create components with accessor capabilities
	positionComp := FactoryNewComponent[Position]()
	velocityComp := FactoryNewComponent[Velocity]()
	healthComp := FactoryNewComponent[Health]()

	// Create initial values
	initialPos := Position{X: 1.0, Y: 2.0}
	initialVel := Velocity{X: 3.0, Y: 4.0}

	// Create entity with components and values
	entities, err := storage.NewEntities(1, healthComp)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}
	entity := entities[0]

	// Add components with values
	if err := entity.AddComponentWithValue(positionComp, initialPos); err != nil {
		t.Fatalf("Failed to add position component: %v", err)
	}
	if err := entity.AddComponentWithValue(velocityComp, initialVel); err != nil {
		t.Fatalf("Failed to add velocity component: %v", err)
	}

	// Get and check values
	posPtr := positionComp.GetFromEntity(entity)
	velPtr := velocityComp.GetFromEntity(entity)

	if posPtr.X != initialPos.X || posPtr.Y != initialPos.Y {
		t.Errorf("Position = {%v, %v}, want {%v, %v}", posPtr.X, posPtr.Y, initialPos.X, initialPos.Y)
	}

	if velPtr.X != initialVel.X || velPtr.Y != initialVel.Y {
		t.Errorf("Velocity = {%v, %v}, want {%v, %v}", velPtr.X, velPtr.Y, initialVel.X, initialVel.Y)
	}

	// Modify values
	posPtr.X = 5.0
	posPtr.Y = 6.0
	velPtr.X = 7.0
	velPtr.Y = 8.0

	// Get again and check modified values
	posPtr2 := positionComp.GetFromEntity(entity)
	velPtr2 := velocityComp.GetFromEntity(entity)

	if posPtr2.X != 5.0 || posPtr2.Y != 6.0 {
		t.Errorf("Updated Position = {%v, %v}, want {5.0, 6.0}", posPtr2.X, posPtr2.Y)
	}

	if velPtr2.X != 7.0 || velPtr2.Y != 8.0 {
		t.Errorf("Updated Velocity = {%v, %v}, want {7.0, 8.0}", velPtr2.X, velPtr2.Y)
	}
}
