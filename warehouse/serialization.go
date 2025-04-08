package warehouse

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"strings"

	"github.com/TheBitDrifter/bappa/table"
)

const (
	strPosInf = "+Infinity"
	strNegInf = "-Infinity"
	strNaN    = "NaN"
)

type SerializedStorage struct {
	Version     string             `json:"version"`
	Entities    []SerializedEntity `json:"entities"`
	CurrentTick int                `json:"current_tick"`
}

// SerializedEntity represents a single entity with its components
type SerializedEntity struct {
	ID         table.EntryID `json:"id"`
	Recycled   int           `json:"recycled"`
	Components []string      `json:"components"`

	// map[componentString]jsonComponentData
	Data map[string]any `json:"data"`
}

func (se SerializedEntity) GetComponents() []Component {
	result := []Component{}
	for _, n := range se.Components {
		c, ok := GlobalTypeRegistry.LookupComp(n)
		if ok {
			result = append(result, c)
		}
	}
	return result
}

// SerializeStorage serializes the storage
func SerializeStorage(s Storage, currentTick int) (*SerializedStorage, error) {
	world := &SerializedStorage{
		Version:     "1.0",
		CurrentTick: currentTick,
	}

	entities := s.Entities()                                    // Get entities once
	world.Entities = make([]SerializedEntity, 0, len(entities)) // Pre-allocate slice

	for _, entity := range entities { // Use the local slice
		if entity == nil || !entity.Valid() { // Add nil check for safety
			continue // Skip invalid entities
		}

		// Create serialized entity
		serializedEntity := SerializedEntity{
			ID:       entity.ID(),
			Recycled: entity.Recycled(),
			// Components: make([]string, 0), // Allocate below
			Data: make(map[string]any),
		}

		// Process each component
		components := entity.Components()                                // Get components once
		serializedEntity.Components = make([]string, 0, len(components)) // Pre-allocate

		for _, comp := range components {
			// Use registered name if available for consistency
			typeName, ok := GlobalTypeRegistry.LookupName(comp)
			if !ok {
				typeName = comp.Type().String() // Fallback to reflection type name
				// Optional: Log warning
			}

			serializedEntity.Components = append(serializedEntity.Components, typeName)

			tbl := entity.Table()
			idx := entity.Index()

			val, err := tbl.Get(comp, idx) // Gets the component value (e.g., TestPosition struct)
			if err != nil {
				log.Printf("Warning: Failed to get component %s for entity %d: %v. Skipping component.", typeName, entity.ID(), err)
				continue
			}

			serializedEntity.Data[typeName] = val.Interface()
		}

		world.Entities = append(world.Entities, serializedEntity)
	}

	return world, nil
}

// DeserializeStorage creates a new storage from a serialized world
func DeserializeStorage(storage Storage, world *SerializedStorage) (Storage, error) {
	return deserializeStorage(storage, world, true)
}

// DeserializeStorage creates a new storage from a serialized world without purging non serialized entities
func DeserializeStorageNoPurge(storage Storage, world *SerializedStorage) (Storage, error) {
	return deserializeStorage(storage, world, false)
}

func deserializeStorage(storage Storage, world *SerializedStorage, purge bool) (Storage, error) {
	updated := map[int]bool{}

	for _, serializedEntity := range world.Entities {
		// Get components slice
		entityComponents := make([]Component, 0)
		for _, compName := range serializedEntity.Components {
			comp, ok := GlobalTypeRegistry.LookupComp(compName)
			if !ok {
				log.Println("here")
				return nil, fmt.Errorf("component not found: %s", compName)
			}
			entityComponents = append(entityComponents, comp)
		}

		entityFromSerialized, err := storage.ForceSerializedEntity(serializedEntity)
		if err != nil {
			return nil, err
		}

		updated[int(entityFromSerialized.ID())] = true

		for compName, compData := range serializedEntity.Data {
			comp, ok := GlobalTypeRegistry.LookupComp(compName)
			if !ok {
				continue
			}

			tbl := entityFromSerialized.Table()
			idx := entityFromSerialized.Index()

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
	if purge {
		purge := []Entity{}
		for _, en := range storage.Entities() {
			if _, ok := updated[int(en.ID())]; !ok {
				purge = append(purge, en)
			}
		}
		return storage, storage.DestroyEntities(purge...)
	}
	return storage, nil
}

// SaveStorage saves the storage to a file
func SaveStorage(s Storage, filename string, currentTick int) error {
	world, err := SerializeStorage(s, currentTick)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}

	// This step converts Inf/NaN to strings and structs to maps *in a new structure*
	worldForJSON, err := prepareForJSONMarshal(world)
	if err != nil {
		return fmt.Errorf("failed to prepare world data for JSON marshalling: %w", err)
	}

	data, err := json.MarshalIndent(worldForJSON, "", "  ")
	if err != nil {
		log.Printf("Error marshalling prepared data: %v", err)
		return fmt.Errorf("JSON marshaling failed unexpectedly after preparation: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}

	return nil
}

// LoadStorage loads a storage from a file
func LoadStorage(filename string) (*SerializedStorage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("file read failed: %w", err)
	}

	var world SerializedStorage
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

// prepareForJSONMarshal recursively traverses data and returns a *new* structure
// suitable for standard JSON marshalling, converting non-standard floats to strings
// and potentially structs to maps.
func prepareForJSONMarshal(value any) (any, error) {
	v := reflect.ValueOf(value)

	// Handle nil input explicitly first
	if !v.IsValid() || ((v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface || v.Kind() == reflect.Map || v.Kind() == reflect.Slice) && v.IsNil()) {
		return nil, nil
	}

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsInf(f, 1) {
			return strPosInf, nil
		}
		if math.IsInf(f, -1) {
			return strNegInf, nil
		}
		if math.IsNaN(f) {
			return strNaN, nil
		}
		return value, nil // Return original value if it's a standard float

	case reflect.Ptr, reflect.Interface:
		// If it's nil, handled above. If not, recurse on the underlying element.
		return prepareForJSONMarshal(v.Elem().Interface())

	case reflect.Struct:
		// Check if it implements json.Marshaler first
		if _, ok := v.Interface().(json.Marshaler); ok {
			// If it has custom marshalling, trust it (assume it handles Inf/NaN if needed)
			return v.Interface(), nil
		}

		structMap := make(map[string]any)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)

			if !fieldType.IsExported() {
				continue
			}

			// Determine field name using json tag logic
			jsonTag := fieldType.Tag.Get("json")
			fieldName := fieldType.Name
			omitEmpty := false
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				tagFieldName := parts[0]
				if tagFieldName == "-" {
					continue
				}
				if tagFieldName != "" {
					fieldName = tagFieldName
				}
				for _, option := range parts[1:] {
					if option == "omitempty" {
						omitEmpty = true
						break
					}
				}
			}

			fieldValue := field.Interface()

			if omitEmpty && reflect.ValueOf(fieldValue).IsZero() {
				continue
			}

			preparedField, err := prepareForJSONMarshal(fieldValue)
			if err != nil {
				return nil, fmt.Errorf("error preparing field %s: %w", fieldName, err)
			}
			structMap[fieldName] = preparedField
		}
		return structMap, nil

	case reflect.Map:
		newMap := make(map[string]any)
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			val := iter.Value()

			var keyStr string
			if key.Kind() == reflect.String {
				keyStr = key.String()
			} else {
				keyStr = fmt.Sprintf("%v", key.Interface())
			}

			preparedValue, err := prepareForJSONMarshal(val.Interface())
			if err != nil {
				return nil, fmt.Errorf("error preparing map value for key %v: %w", key.Interface(), err)
			}
			newMap[keyStr] = preparedValue
		}
		return newMap, nil

	case reflect.Slice, reflect.Array:
		newSlice := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			preparedElem, err := prepareForJSONMarshal(v.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("error preparing slice element %d: %w", i, err)
			}
			newSlice[i] = preparedElem
		}
		return newSlice, nil

	default:
		return value, nil
	}
}

// Helper function to convert maps to structs or other conversions needed
func convertToType(data any, targetType reflect.Type) (any, error) {
	// If data is nil, handle based on target type nillability
	if data == nil {
		switch targetType.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			// Return a typed nil for nillable types
			return reflect.Zero(targetType).Interface(), nil
		default:
			// Return zero value for non-nillable types (struct, int, array, etc.)
			return reflect.Zero(targetType).Interface(), nil
		}
	}

	dataVal := reflect.ValueOf(data)
	dataType := dataVal.Type()

	// Handle special float strings -> float target
	if dataType.Kind() == reflect.String {
		strVal := dataVal.String()
		if targetType.Kind() == reflect.Float64 || targetType.Kind() == reflect.Float32 {
			switch strVal {
			case strPosInf:
				return math.Inf(1), nil
			case strNegInf:
				return math.Inf(-1), nil
			case strNaN:
				return math.NaN(), nil
			}
		}
	}
	if dataType.AssignableTo(targetType) {
		return data, nil
	}

	if (targetType.Kind() == reflect.Slice || targetType.Kind() == reflect.Array) && dataType.Kind() == reflect.Slice {
		if dataSlice, ok := data.([]interface{}); ok {
			targetElemType := targetType.Elem() // Get the type of elements in the target (e.g., client.SpriteBlueprint)
			targetLen := len(dataSlice)

			var newCollection reflect.Value

			if targetType.Kind() == reflect.Slice {
				newCollection = reflect.MakeSlice(targetType, targetLen, targetLen)
			} else {
				if targetType.Len() != targetLen {
					return nil, fmt.Errorf("array length mismatch: input slice has length %d, target array [%d]%s requires %d",
						targetLen, targetType.Len(), targetElemType.String(), targetType.Len())
				}
				newCollection = reflect.New(targetType).Elem()
			}

			for i, elemData := range dataSlice {
				convertedElem, err := convertToType(elemData, targetElemType)
				if err != nil {
					return nil, fmt.Errorf("error converting element %d for %s: %w", i, targetType.String(), err)
				}

				if newCollection.Index(i).CanSet() {
					if reflect.ValueOf(convertedElem).IsValid() {
						if reflect.TypeOf(convertedElem).AssignableTo(newCollection.Index(i).Type()) {
							newCollection.Index(i).Set(reflect.ValueOf(convertedElem))
						} else if reflect.ValueOf(convertedElem).CanConvert(newCollection.Index(i).Type()) {
							newCollection.Index(i).Set(reflect.ValueOf(convertedElem).Convert(newCollection.Index(i).Type()))
						} else {
							return nil, fmt.Errorf("type mismatch for element %d: cannot assign/convert %T to %s", i, convertedElem, newCollection.Index(i).Type())
						}
					} else {
						switch newCollection.Index(i).Kind() {
						case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
							newCollection.Index(i).Set(reflect.Zero(newCollection.Index(i).Type()))
						}
					}
				} else {
					log.Printf("Warning: Cannot set index %d for type %s", i, targetType.String())
				}
			}
			return newCollection.Interface(), nil
		}
	}

	if mapData, ok := data.(map[string]interface{}); ok && targetType.Kind() == reflect.Struct {
		newInstance := reflect.New(targetType).Elem()
		for i := 0; i < targetType.NumField(); i++ {
			field := targetType.Field(i)
			if !field.IsExported() {
				continue
			}
			jsonTag := field.Tag.Get("json")
			fieldNameInMap := field.Name
			if jsonTag != "" && jsonTag != "-" {
				parts := strings.Split(jsonTag, ",")
				fieldNameInMap = parts[0]
			} else if jsonTag == "-" {
				continue
			}

			if fieldValue, fieldExists := mapData[fieldNameInMap]; fieldExists {
				convertedFieldVal, err := convertToType(fieldValue, field.Type)
				if err != nil {
					return nil, fmt.Errorf("error converting field '%s' (target type %s): %w", fieldNameInMap, field.Type.String(), err)
				}
				fieldToSet := newInstance.FieldByName(field.Name)
				if fieldToSet.CanSet() {
					if reflect.ValueOf(convertedFieldVal).IsValid() {
						if reflect.TypeOf(convertedFieldVal).AssignableTo(fieldToSet.Type()) {
							fieldToSet.Set(reflect.ValueOf(convertedFieldVal))
						} else if reflect.ValueOf(convertedFieldVal).CanConvert(fieldToSet.Type()) {
							fieldToSet.Set(reflect.ValueOf(convertedFieldVal).Convert(fieldToSet.Type()))
						} else {
							return nil, fmt.Errorf("type mismatch for field '%s': cannot assign/convert %T to %s", fieldNameInMap, convertedFieldVal, fieldToSet.Type())
						}
					} else {
						switch fieldToSet.Kind() {
						case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
							fieldToSet.Set(reflect.Zero(fieldToSet.Type()))
						}
					}
				}
			}
		}
		return newInstance.Interface(), nil
	}
	if _, ok := data.(float64); ok {
		if dataVal.CanConvert(targetType) {
			return dataVal.Convert(targetType).Interface(), nil
		}
	}
	if _, ok := data.(int); ok {
		if dataVal.CanConvert(targetType) {
			return dataVal.Convert(targetType).Interface(), nil
		}
	}
	if _, ok := data.(int64); ok {
		if dataVal.CanConvert(targetType) {
			return dataVal.Convert(targetType).Interface(), nil
		}
	}

	// If we still can't convert, return an error
	return nil, fmt.Errorf("cannot convert type %T to %s (value: %#v)", data, targetType, data)
}
