// Package warehouse provides query mechanisms for component-based entity systems
package warehouse

import (
	"fmt"

	"github.com/TheBitDrifter/bark"
	"github.com/TheBitDrifter/mask"
)

// Query represents a composable query interface for filtering entities
type Query interface {
	QueryNode
	And(items ...interface{}) QueryNode
	Or(items ...interface{}) QueryNode
	Not(items ...interface{}) QueryNode
}

// QueryNode represents a node in the query tree that can be evaluated
type QueryNode interface {
	Evaluate(archetype Archetype, storage Storage) bool
}

// QueryOperation defines the logical operations for query nodes
type QueryOperation int

const (
	OpAnd QueryOperation = iota // Logical AND operation
	OpOr                        // Logical OR operation
	OpNot                       // Logical NOT operation
)

// compositeNode implements a compound query with child nodes
type compositeNode struct {
	op         QueryOperation
	children   []QueryNode
	components []Component
}

// leafNode implements a simple query with no child nodes
type leafNode struct {
	components []Component
}

// query implements the Query interface
type query struct {
	root QueryNode
}

// newQuery creates a new empty query
func newQuery() Query {
	return &query{}
}

// newCompositeNode creates a new composite query node with the specified operation
func newCompositeNode(op QueryOperation, components []Component) *compositeNode {
	return &compositeNode{
		op:         op,
		children:   make([]QueryNode, 0),
		components: components,
	}
}

// newLeafNode creates a new leaf query node with the specified components
func newLeafNode(components []Component) *leafNode {
	return &leafNode{components: components}
}

// Evaluate implements the QueryNode interface for composite nodes
func (n *compositeNode) Evaluate(archetype Archetype, storage Storage) bool {
	var nodeMask mask.Mask
	for _, comp := range n.components {
		bit := storage.RowIndexFor(comp)
		nodeMask.Mark(bit)
	}
	archeMask := archetype.Table().(mask.Maskable).Mask()

	switch n.op {
	case OpAnd:
		if !archeMask.ContainsAll(nodeMask) {
			return false
		}
		for _, child := range n.children {
			if !child.Evaluate(archetype, storage) {
				return false
			}
		}
		return true
	case OpOr:
		if archeMask.ContainsAny(nodeMask) {
			return true
		}
		for _, child := range n.children {
			if child.Evaluate(archetype, storage) {
				return true
			}
		}
		return false
	case OpNot:
		if len(n.children) == 0 {
			return archeMask.ContainsNone(nodeMask)
		}
		if len(n.components) > 0 && !archeMask.ContainsNone(nodeMask) {
			return false
		}
		for _, child := range n.children {
			if child.Evaluate(archetype, storage) {
				return false
			}
		}
		return true
	}
	return false
}

// Evaluate implements the QueryNode interface for leaf nodes
func (n *leafNode) Evaluate(archetype Archetype, storage Storage) bool {
	var nodeMask mask.Mask
	for _, comp := range n.components {
		bit := storage.RowIndexFor(comp)
		nodeMask.Mark(bit)
	}
	archeMask := archetype.Table().(mask.Maskable).Mask()
	return archeMask.ContainsAll(nodeMask)
}

// And creates a new AND operation node with the provided items
func (q *query) And(items ...interface{}) QueryNode {
	components, children := q.processItems(items...)
	node := newCompositeNode(OpAnd, components)
	node.children = children
	if q.root == nil {
		q.root = node
	}
	return node
}

// Or creates a new OR operation node with the provided items
func (q *query) Or(items ...interface{}) QueryNode {
	components, children := q.processItems(items...)
	node := newCompositeNode(OpOr, components)
	node.children = children
	if q.root == nil {
		q.root = node
	}
	return node
}

// Not creates a new NOT operation node with the provided items
func (q *query) Not(items ...interface{}) QueryNode {
	components, children := q.processItems(items...)
	node := newCompositeNode(OpNot, components)
	node.children = children
	if q.root == nil {
		q.root = node
	}
	return node
}

// validateQueryItems checks if all items are of valid types for queries
func (q *query) validateQueryItems(items ...interface{}) error {
	for _, item := range items {
		switch item.(type) {
		case Component, []Component, QueryNode, Query:
			continue
		default:
			return fmt.Errorf("invalid query item type: %T. Only Component, []Component, or QueryNode are allowed", item)
		}
	}
	return nil
}

// processItems converts the input items into components and query nodes
func (q *query) processItems(items ...interface{}) ([]Component, []QueryNode) {
	if err := q.validateQueryItems(items...); err != nil {
		panic(bark.AddTrace(err))
	}
	components := make([]Component, 0)
	children := make([]QueryNode, 0)
	for _, item := range items {
		switch v := item.(type) {
		case Component:
			components = append(components, v)
		case []Component:
			components = append(components, v...)
		case QueryNode:
			children = append(children, v)
		}
	}
	return components, children
}

// Evaluate implements the QueryNode interface for the query type
func (q *query) Evaluate(archetype Archetype, storage Storage) bool {
	if q.root == nil {
		return false
	}
	return q.root.Evaluate(archetype, storage)
}
