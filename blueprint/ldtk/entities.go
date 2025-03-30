package ldtk

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/TheBitDrifter/warehouse"
)

// EntityHandler defines how to process an entity from the LDtk file
type EntityHandler func(entity *LDtkEntityInstance, sto warehouse.Storage) error

// LDtkEntityRegistry manages handlers for different entity types
type LDtkEntityRegistry struct {
	handlers map[string]EntityHandler
}

// Register adds a handler for a specific entity type
func (r *LDtkEntityRegistry) Register(entityType string, handler EntityHandler) {
	r.handlers[entityType] = handler
}

// NewLDtkEntityRegistry creates a new entity registry
func NewLDtkEntityRegistry() *LDtkEntityRegistry {
	return &LDtkEntityRegistry{
		handlers: make(map[string]EntityHandler),
	}
}

// LDtkEntityInstance represents an entity instance from LDtk
type LDtkEntityInstance struct {
	Identifier     string              `json:"__identifier"`
	IID            string              `json:"iid"`
	Position       [2]int              `json:"px"` // [x, y]
	Width          int                 `json:"width"`
	Height         int                 `json:"height"`
	FieldInstances []LDtkFieldInstance `json:"fieldInstances"`
}

// LDtkFieldInstance represents a field instance in an entity
type LDtkFieldInstance struct {
	Identifier string          `json:"__identifier"`
	Type       string          `json:"__type"`
	Value      json.RawMessage `json:"__value"`
}

// GetStringField extracts a string field from an entity instance
func (e *LDtkEntityInstance) GetStringField(name string) (string, error) {
	for _, field := range e.FieldInstances {
		if field.Identifier == name {
			var value string
			if err := json.Unmarshal(field.Value, &value); err != nil {
				return "", err
			}
			return value, nil
		}
	}
	return "", fmt.Errorf("field '%s' not found", name)
}

// GetIntField extracts an int field from an entity instance
func (e *LDtkEntityInstance) GetIntField(name string) (int, error) {
	for _, field := range e.FieldInstances {
		if field.Identifier == name {
			var value int
			if err := json.Unmarshal(field.Value, &value); err != nil {
				return 0, err
			}
			return value, nil
		}
	}
	return 0, fmt.Errorf("field '%s' not found", name)
}

// GetFloatField extracts a float field from an entity instance
func (e *LDtkEntityInstance) GetFloatField(name string) (float64, error) {
	for _, field := range e.FieldInstances {
		if field.Identifier == name {
			var value float64
			if err := json.Unmarshal(field.Value, &value); err != nil {
				return 0, err
			}
			return value, nil
		}
	}
	return 0, fmt.Errorf("field '%s' not found", name)
}

// GetBoolField extracts a boolean field from an entity instance
func (e *LDtkEntityInstance) GetBoolField(name string) (bool, error) {
	for _, field := range e.FieldInstances {
		if field.Identifier == name {
			var value bool
			if err := json.Unmarshal(field.Value, &value); err != nil {
				return false, err
			}
			return value, nil
		}
	}
	return false, fmt.Errorf("field '%s' not found", name)
}

// StringFieldOr gets a string field with a default fallback value
func (e *LDtkEntityInstance) StringFieldOr(name string, defaultValue string) string {
	val, err := e.GetStringField(name)
	if err != nil {
		return defaultValue
	}
	return val
}

// IntFieldOr gets an int field with a default fallback value
func (e *LDtkEntityInstance) IntFieldOr(name string, defaultValue int) int {
	val, err := e.GetIntField(name)
	if err != nil {
		return defaultValue
	}
	return val
}

// FloatFieldOr gets a float field with a default fallback value
func (e *LDtkEntityInstance) FloatFieldOr(name string, defaultValue float64) float64 {
	val, err := e.GetFloatField(name)
	if err != nil {
		return defaultValue
	}
	return val
}

// BoolFieldOr gets a boolean field with a default fallback value
func (e *LDtkEntityInstance) BoolFieldOr(name string, defaultValue bool) bool {
	val, err := e.GetBoolField(name)
	if err != nil {
		return defaultValue
	}
	return val
}

// LoadEntities loads entities from the specified level using registered handlers
func (p *LDtkProject) LoadEntities(levelName string, sto warehouse.Storage, registry *LDtkEntityRegistry) error {
	level, exists := p.parsedLevels[levelName]
	if !exists {
		log.Printf("Level '%s' not found", levelName)
		return nil
	}

	// Process each Entity layer
	entitiesProcessed := 0
	for _, entityInstances := range level.EntityRawData {
		for i := range entityInstances {
			entity := &entityInstances[i]
			handler, exists := registry.handlers[entity.Identifier]

			if !exists {
				log.Printf("No handler registered for entity type: %s", entity.Identifier)
				continue
			}

			if err := handler(entity, sto); err != nil {
				return err
			}

			entitiesProcessed++
		}
	}

	return nil
}
