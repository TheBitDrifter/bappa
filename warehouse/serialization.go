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

func (se SerializedEntity) SetValue(entity Entity) error {
	for compName, compData := range se.Data {
		comp, ok := GlobalTypeRegistry.LookupComp(compName)
		if !ok {
			continue
		}

		tbl := entity.Table()
		if !tbl.Contains(comp) {
			continue
		}
		idx := entity.Index()

		targetType := comp.Type()

		convertedValue, err := convertToType(compData, targetType)
		if err != nil {
			return fmt.Errorf("failed to convert component data for %s: %w", compName, err)
		}

		err = tbl.Set(comp, reflect.ValueOf(convertedValue), idx)
		if err != nil {
			return fmt.Errorf("failed to set component data: %w", err)
		}
	}
	return nil
}

// SerializeStorage serializes the storage
func SerializeStorage(s Storage, currentTick int) (*SerializedStorage, error) {
	world := &SerializedStorage{
		Version:     "1.0",
		CurrentTick: currentTick,
	}

	entities := s.Entities()
	world.Entities = make([]SerializedEntity, 0, len(entities))

	for _, entity := range entities {
		if entity == nil || !entity.Valid() {
			continue
		}
		serializedEntity := entity.Serialize()
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
				return nil, fmt.Errorf("component not found: %s", compName)
			}
			entityComponents = append(entityComponents, comp)
		}

		entityFromSerialized, err := storage.ForceSerializedEntity(serializedEntity)
		if err != nil {
			return nil, err
		}

		updated[int(entityFromSerialized.ID())] = true
		err = serializedEntity.SetValue(entityFromSerialized)
		if err != nil {
			return nil, err
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
	worldForJSON, err := PrepareForJSONMarshal(world)
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
func PrepareForJSONMarshal(value any) (any, error) {
	v := reflect.ValueOf(value)

	// Handle nil input explicitly first
	if !v.IsValid() || ((v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface || v.Kind() == reflect.Map || v.Kind() == reflect.Slice) && v.IsNil()) {
		return nil, nil
	}

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		// ... (existing float handling code) ...
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
		return value, nil

	case reflect.Ptr, reflect.Interface:
		// ... (existing pointer/interface handling code) ...
		if v.IsNil() {
			return nil, nil
		}
		return PrepareForJSONMarshal(v.Elem().Interface())

	case reflect.Struct:
		// Check if it implements json.Marshaler first
		if marshaler, ok := v.Interface().(json.Marshaler); ok {
			return marshaler, nil
		}

		structMap := make(map[string]any)
		t := v.Type()
		typeName := t.String()

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)

			if !fieldType.IsExported() {
				continue
			}

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

			preparedField, err := PrepareForJSONMarshal(fieldValue)
			if err != nil {
				return nil, fmt.Errorf("error preparing field %s (type %s) in struct %s: %w", fieldName, fieldType.Type, typeName, err)
			}
			structMap[fieldName] = preparedField
		}
		return structMap, nil

	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			newMap := make(map[string]any)
			iter := v.MapRange()
			for iter.Next() {
				keyStr := fmt.Sprintf("%v", iter.Key().Interface())
				preparedValue, err := PrepareForJSONMarshal(iter.Value().Interface())
				if err != nil {
					return nil, fmt.Errorf("map value error for key %s: %w", keyStr, err)
				}
				newMap[keyStr] = preparedValue
			}
			return newMap, nil
		} else {
			newMap := make(map[string]any, v.Len())
			iter := v.MapRange()
			for iter.Next() {
				key := iter.Key().String()
				preparedValue, err := PrepareForJSONMarshal(iter.Value().Interface())
				if err != nil {
					return nil, fmt.Errorf("map value error for key %s: %w", key, err)
				}
				newMap[key] = preparedValue
			}
			return newMap, nil
		}

	case reflect.Slice, reflect.Array:
		newSlice := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			preparedElem, err := PrepareForJSONMarshal(v.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("slice element %d error: %w", i, err)
			}
			newSlice[i] = preparedElem
		}
		return newSlice, nil

	default: // Basic types
		return value, nil
	}
}

// Helper function to convert maps to structs or other conversions needed
func convertToType(data any, targetType reflect.Type) (any, error) {
	// If data is nil, handle based on target type nillability
	if data == nil {
		switch targetType.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			return reflect.Zero(targetType).Interface(), nil // Typed nil
		default:
			return reflect.Zero(targetType).Interface(), nil // Zero value
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

	// Direct Assignability
	if dataType.AssignableTo(targetType) {
		return data, nil
	}

	// Slice/Array Conversion
	if (targetType.Kind() == reflect.Slice || targetType.Kind() == reflect.Array) && dataType.Kind() == reflect.Slice {
		// Check if it's the expected []interface{} from json unmarshal
		if dataSlice, ok := data.([]interface{}); ok {
			targetElemType := targetType.Elem()
			targetLen := len(dataSlice)
			var newCollection reflect.Value

			if targetType.Kind() == reflect.Slice {
				newCollection = reflect.MakeSlice(targetType, targetLen, targetLen)
			} else { // Array
				if targetType.Len() != targetLen {
					return nil, fmt.Errorf("array length mismatch: input slice len %d, target array [%d]%s requires %d",
						targetLen, targetType.Len(), targetElemType.String(), targetType.Len())
				}
				newCollection = reflect.New(targetType).Elem()
			}

			for i, elemData := range dataSlice {
				// Recursive call for slice/array elements
				convertedElem, err := convertToType(elemData, targetElemType)
				if err != nil {
					return nil, fmt.Errorf("error converting element %d for %s: %w", i, targetType.String(), err)
				}

				elemToSet := newCollection.Index(i)
				if elemToSet.CanSet() {
					if ceVal := reflect.ValueOf(convertedElem); ceVal.IsValid() {
						if ceVal.Type().AssignableTo(elemToSet.Type()) {
							elemToSet.Set(ceVal)
						} else if ceVal.CanConvert(elemToSet.Type()) {
							elemToSet.Set(ceVal.Convert(elemToSet.Type()))
						} else {
							return nil, fmt.Errorf("type mismatch for slice element %d: cannot assign/convert %T to %s", i, convertedElem, elemToSet.Type())
						}
					} else { // Handle nil/invalid from recursion
						switch elemToSet.Kind() {
						case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
							elemToSet.Set(reflect.Zero(elemToSet.Type()))
						}
					}
				} else {
					// This path should generally not be reachable for valid slice/array creation
					log.Printf("Warning: Cannot set slice/array index %d for type %s", i, targetType.String())
				}
			}
			return newCollection.Interface(), nil
		}
		// else: Input is slice but not []interface{} - fall through to error later
	}

	// Map -> Struct Conversion
	if mapData, ok := data.(map[string]interface{}); ok && targetType.Kind() == reflect.Struct {
		newInstance := reflect.New(targetType).Elem()
		typeName := targetType.String() // For context in logs/errors

		for i := 0; i < targetType.NumField(); i++ {
			field := targetType.Field(i) // reflect.StructField (metadata)
			if !field.IsExported() {
				continue
			}

			// Determine the key name in the map (respecting json tags)
			jsonTag := field.Tag.Get("json")
			fieldNameInMap := field.Name // Default to Go field name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				tagFieldName := parts[0]
				if tagFieldName == "-" {
					continue
				}
				if tagFieldName != "" {
					fieldNameInMap = tagFieldName
				}
			}

			if fieldValue, fieldExists := mapData[fieldNameInMap]; fieldExists {

				// Recursive call to convert the map value to the field's type
				convertedFieldVal, err := convertToType(fieldValue, field.Type)
				if err != nil {
					// Add more context to the error
					return nil, fmt.Errorf("error converting field '%s' (target type %s) in struct '%s': %w", fieldNameInMap, field.Type.String(), typeName, err)
				}

				fieldToSet := newInstance.FieldByName(field.Name)
				if fieldToSet.CanSet() {
					cfvVal := reflect.ValueOf(convertedFieldVal)

					if cfvVal.IsValid() {
						if cfvVal.Type().AssignableTo(fieldToSet.Type()) {
							fieldToSet.Set(cfvVal)
						} else if cfvVal.CanConvert(fieldToSet.Type()) {
							fieldToSet.Set(cfvVal.Convert(fieldToSet.Type()))
						} else {
							// Error if type mismatch after conversion attempt
							return nil, fmt.Errorf("type mismatch field '%s': cannot assign or convert %T to %s", fieldNameInMap, convertedFieldVal, fieldToSet.Type())
						}
					} else {
						// Handle nil/invalid converted value (only set zero for nillable kinds)
						if k := fieldToSet.Kind(); k == reflect.Chan || k == reflect.Func || k == reflect.Interface || k == reflect.Map || k == reflect.Ptr || k == reflect.Slice {
							fieldToSet.Set(reflect.Zero(fieldToSet.Type()))
						}
					}
				}
			}
		}
		return newInstance.Interface(), nil
	}
	// Numeric Conversion (JSON numbers often float64)
	if dataType.Kind() == reflect.Float64 {
		// Check if target is some kind of integer
		if targetType.Kind() >= reflect.Int && targetType.Kind() <= reflect.Uint64 {
			// Optional precision loss check: if f64 != math.Trunc(f64) { log warning }
			// Use Convert for float->int truncation
			if dataVal.CanConvert(targetType) {
				return dataVal.Convert(targetType).Interface(), nil
			}
		} else if dataVal.CanConvert(targetType) { // Handle float64 -> float32 etc.
			return dataVal.Convert(targetType).Interface(), nil
		}
	} else if (dataType.Kind() >= reflect.Int && dataType.Kind() <= reflect.Int64) || (dataType.Kind() >= reflect.Uint && dataType.Kind() <= reflect.Uint64) {
		// Handle integer to integer/float conversions
		if dataVal.CanConvert(targetType) {
			return dataVal.Convert(targetType).Interface(), nil
		}
	}
	// If we still haven't returned, conversion failed
	return nil, fmt.Errorf("cannot convert type %T to %s (value: %#v)", data, targetType, data)
}
