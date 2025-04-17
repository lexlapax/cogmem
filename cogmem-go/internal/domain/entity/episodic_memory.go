package entity

import (
	"time"

	"github.com/google/uuid"
)

// EpisodicMemory represents a stored memory item in the episodic memory store.
type EpisodicMemory struct {
	// ID is the unique identifier of the memory.
	ID uuid.UUID
	// UserID associates the memory to a specific user partition.
	UserID uuid.UUID
	// EntityID optionally scopes the memory to a sub-context (e.g., session or agent ID).
	EntityID *uuid.UUID
	// Content is the raw text or data stored.
	Content string
	// Embedding is the vector representation of the content.
	Embedding []float32
	// Timestamp indicates when the memory was created.
	Timestamp time.Time
	// ShareScope defines visibility (e.g., "user", "global").
	ShareScope string
	// LastAccessed records the last retrieval time.
	LastAccessed time.Time
	// AccessibilityScore indicates current retrieval priority (default 1.0).
	AccessibilityScore float64
}

// NewEpisodicMemory constructs a new EpisodicMemory with default values.
func NewEpisodicMemory(userID uuid.UUID, entityID *uuid.UUID, content string, embedding []float32, timestamp time.Time, shareScope string) *EpisodicMemory {
	return &EpisodicMemory{
		ID:                 uuid.New(),
		UserID:             userID,
		EntityID:           entityID,
		Content:            content,
		Embedding:          embedding,
		Timestamp:          timestamp,
		ShareScope:         shareScope,
		LastAccessed:       timestamp,
		AccessibilityScore: 1.0,
	}
}
