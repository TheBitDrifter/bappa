package drip

import "github.com/TheBitDrifter/bappa/table"

// AssignEntityIDMessage informs a client of its server-side entity ID.
type AssignEntityIDMessage struct {
	Type     string        `json:"type"`
	EntityID table.EntryID `json:"entity_id"`
}

// AssignEntityIDMessageType identifies the AssignEntityIDMessage type.
const AssignEntityIDMessageType = "assign_entity_id"
