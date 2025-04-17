package entity

import "github.com/google/uuid"

// PartitionContext identifies an owner context for memory partitioning.
type PartitionContext struct {
	// UserID is the primary partition key (e.g., user identifier).
	UserID uuid.UUID
	// EntityID is an optional secondary partition qualifier (e.g., session or agent ID).
	EntityID *uuid.UUID
}
