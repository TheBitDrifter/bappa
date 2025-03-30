package table

import (
	"fmt"
	"strings"
)

type (
	AccessError struct {
		Index      int
		UpperBound int
	}
	InvalidElementAccessError struct {
		ElementType       ElementType
		ValidElementTypes []ElementType
	}
	InvalidEntryAccessError struct{}
)

func (e AccessError) Error() string {
	return fmt.Sprintf("access error: index %d out of bounds [0, %d]", e.Index, e.UpperBound)
}

func (e InvalidElementAccessError) Error() string {
	var validTypes []string
	for _, et := range e.ValidElementTypes {
		validTypes = append(validTypes, et.Type().String())
	}
	return fmt.Sprintf("invalid element access error: elementType(%v) is invalid, valid types are [%s]",
		e.ElementType.Type(), strings.Join(validTypes, ", "))
}

func (e InvalidEntryAccessError) Error() string {
	return "cannot access invalid entry"
}

type (
	BatchOperationError struct {
		Count int
	}

	BatchDeletionError struct {
		BatchOperationError
		Capacity int
	}
)

func (e BatchOperationError) Error() string {
	return fmt.Sprintf("batch operation error: amount %d is invalid", e.Count)
}

func (e BatchDeletionError) Error() string {
	return fmt.Sprintf("batch deletion error: amount %d is invalid for capacity %d", e.Count, e.Capacity)
}

type TransferEntryIndexMismatchError struct{}

func (e TransferEntryIndexMismatchError) Error() string {
	return "transfer error entry index mismatch"
}

type TableInstantiationNilSchemaError struct{}

func (e TableInstantiationNilSchemaError) Error() string {
	return "cannot create table without schema, consider using default 'nilSchema' like object"
}

type TableInstantiationNilElementTypesError struct{}

func (e TableInstantiationNilElementTypesError) Error() string {
	return "cannot create table without any elementTypes"
}
