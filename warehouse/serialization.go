package warehouse

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/TheBitDrifter/bappa/table"
)

// SerializedWorld represents the complete serialized state of the ECS
type SerializedWorld struct {
	Version        string             `json:"version"`
	ComponentTypes []string           `json:"componentTypes"`
	Entities       []SerializedEntity `json:"entities"`
}

// SerializedEntity represents a single entity with its components
type SerializedEntity struct {
	ID         table.EntryID  `json:"id"`
	Recycled   int            `json:"recycled"`
	Components []string       `json:"components"`
	Data       map[string]any `json:"data"`
}

// SerializeStorage serializes the entire ECS world
func SerializeStorage(s Storage) (*SerializedWorld, error) {
	world := &SerializedWorld{
		Version:        "1.0",
		ComponentTypes: make([]string, 0),
		Entities:       make([]SerializedEntity, 0),
	}

	// Collect all component type names
	componentNames := make(map[string]bool)

	// Process each valid entity
	for i := 0; i < len(globalEntities); i++ {
		entity := globalEntities[i]
		if !entity.Valid() {
			continue // Skip invalid entities
		}

		// Create serialized entity
		serializedEntity := SerializedEntity{
			ID:         entity.ID(),
			Recycled:   entity.Recycled(),
			Components: make([]string, 0),
			Data:       make(map[string]any),
		}

		// Process each component
		for _, comp := range entity.Components() {
			typeName := comp.Type().String() // Gets full type name including package

			// Also look up in the registry for consistent naming
			if registeredName, ok := GlobalTypeRegistry.LookupName(comp); ok {
				typeName = registeredName
			}

			componentNames[typeName] = true
			serializedEntity.Components = append(serializedEntity.Components, typeName)

			// Get component data directly from the table
			tbl := entity.Table()
			idx := entity.Index()

			val, err := tbl.Get(comp, idx)
			if err != nil {
				return nil, fmt.Errorf("failed to get component data: %w", err)
			}

			serializedEntity.Data[typeName] = val.Interface()
		}

		world.Entities = append(world.Entities, serializedEntity)
	}

	// Convert component names map to slice
	for name := range componentNames {
		world.ComponentTypes = append(world.ComponentTypes, name)
	}

	return world, nil
}

// DeserializeStorage creates a new storage from a serialized world
func DeserializeStorage(world *SerializedWorld) (Storage, error) {
	if len(globalEntities) > 0 {
		ResetAll()
	}
	// Create a fresh storage
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create component map from registered types
	componentMap := make(map[string]Component)
	for _, typeName := range world.ComponentTypes {
		// Try to find the type by name
		comp, ok := GlobalTypeRegistry.LookupComp(typeName)
		if !ok {
			return nil, fmt.Errorf("unknown component type: %s - make sure to register all types before deserialization", typeName)
		}

		componentMap[typeName] = comp
	}

	// Track expected entity ID for hole creation
	nextExpectedID := table.EntryID(1)

	// Process each entity
	for _, serializedEntity := range world.Entities {
		// Create holes if needed
		if serializedEntity.ID > nextExpectedID {
			holeCount := int(serializedEntity.ID - nextExpectedID)
			// Create holes individually
			for i := 0; i < holeCount; i++ {
				globalEntryIndex.NewHole()
				globalEntities = append(globalEntities, entity{})
			}
			nextExpectedID = serializedEntity.ID
		}
		nextExpectedID++ // Increment for next expected ID

		// Collect components for this entity
		entityComponents := make([]Component, 0)
		for _, compName := range serializedEntity.Components {
			comp, ok := componentMap[compName]
			if !ok {
				return nil, fmt.Errorf("component not found: %s", compName)
			}
			entityComponents = append(entityComponents, comp)
		}

		// Create entity with components
		entities, err := storage.NewEntitiesNoRecycle(1, entityComponents...)
		if err != nil {
			return nil, fmt.Errorf("failed to create entity: %w", err)
		}
		entity := entities[0]

		// Verify correct ID assignment
		if entity.ID() != serializedEntity.ID {
			return nil, fmt.Errorf("entity ID mismatch: expected %d, got %d", serializedEntity.ID, entity.ID())
		}

		// Set component values
		for compName, compData := range serializedEntity.Data {
			comp, ok := componentMap[compName]
			if !ok {
				continue
			}

			// Get the table and index
			tbl := entity.Table()
			idx := entity.Index()

			// Convert the data to the correct type
			targetType := comp.Type()
			convertedValue, err := convertToType(compData, targetType)
			if err != nil {
				return nil, fmt.Errorf("failed to convert component data for %s: %w", compName, err)
			}

			// Set the converted value
			err = tbl.Set(comp, reflect.ValueOf(convertedValue), idx)
			if err != nil {
				return nil, fmt.Errorf("failed to set component data: %w", err)
			}
		}
	}

	return storage, nil
}

// Helper function to convert maps to structs or other conversions needed
func convertToType(data any, targetType reflect.Type) (any, error) {
	// If data is already of the correct type, return it
	dataVal := reflect.ValueOf(data)
	if dataVal.Type().AssignableTo(targetType) {
		return data, nil
	}

	// Check if we're dealing with a map that needs to be converted to a struct
	if mapData, ok := data.(map[string]interface{}); ok && targetType.Kind() == reflect.Struct {
		// Create a new instance of the target type
		newInstance := reflect.New(targetType).Elem()

		// Copy fields from map to struct
		for i := 0; i < targetType.NumField(); i++ {
			field := targetType.Field(i)
			fieldName := field.Name

			// Look for the field in the map (case-sensitive)
			if fieldValue, ok := mapData[fieldName]; ok {
				// Convert the field value to the correct type
				fieldVal, err := convertToType(fieldValue, field.Type)
				if err != nil {
					return nil, err
				}

				// Set the field
				fieldToSet := newInstance.FieldByName(fieldName)
				if fieldToSet.CanSet() {
					fieldToSet.Set(reflect.ValueOf(fieldVal))
				}
			}
		}

		return newInstance.Interface(), nil
	}

	// Handle number type conversions for JSON numbers
	if num, ok := data.(float64); ok {
		switch targetType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int(num), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return uint(num), nil
		case reflect.Float32:
			return float32(num), nil
		}
	}

	// If we can't convert, return an error
	return nil, fmt.Errorf("cannot convert %T to %s", data, targetType)
}

// SaveStorage saves the storage to a file
func SaveStorage(s Storage, filename string) error {
	// Serialize storage
	world, err := SerializeStorage(s)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(world, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON marshaling failed: %w", err)
	}

	// Write to file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}

	return nil
}

// LoadStorage loads a storage from a file
func LoadStorage(filename string) (*SerializedWorld, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("file read failed: %w", err)
	}

	// Unmarshal from JSON
	var world SerializedWorld
	err = json.Unmarshal(data, &world)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshaling failed: %w", err)
	}

	return &world, nil
}

// Rests globalEntities && globalEntryIndex
func ResetAll() {
	globalEntities = []entity{}
	globalEntryIndex.Reset()
}
