package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNewEpisodicMemoryDefaults verifies that the constructor sets default fields correctly.
func TestNewEpisodicMemoryDefaults(t *testing.T) {
	userID := uuid.New()
	// entityID is optional; test nil scenario
	var entityID *uuid.UUID
	content := "test memory"
	embedding := []float32{0.1, 0.2, 0.3}
	ts := time.Now().UTC().Truncate(time.Second)
	shareScope := "user"

	mem := NewEpisodicMemory(userID, entityID, content, embedding, ts, shareScope)

	if mem == nil {
		t.Fatal("Expected non-nil EpisodicMemory")
	}
	if mem.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", mem.UserID, userID)
	}
	if mem.EntityID != entityID {
		t.Errorf("EntityID mismatch: got %v, want %v", mem.EntityID, entityID)
	}
	if mem.Content != content {
		t.Errorf("Content mismatch: got %q, want %q", mem.Content, content)
	}
	if len(mem.Embedding) != len(embedding) {
		t.Errorf("Embedding length mismatch: got %d, want %d", len(mem.Embedding), len(embedding))
	} else {
		for i := range embedding {
			if mem.Embedding[i] != embedding[i] {
				t.Errorf("Embedding[%d] mismatch: got %v, want %v", i, mem.Embedding[i], embedding[i])
			}
		}
	}
	if !mem.Timestamp.Equal(ts) {
		t.Errorf("Timestamp mismatch: got %v, want %v", mem.Timestamp, ts)
	}
	if !mem.LastAccessed.Equal(mem.Timestamp) {
		t.Errorf("LastAccessed mismatch: got %v, want %v", mem.LastAccessed, mem.Timestamp)
	}
	if mem.AccessibilityScore != 1.0 {
		t.Errorf("AccessibilityScore mismatch: got %v, want %v", mem.AccessibilityScore, 1.0)
	}
	if mem.ID == uuid.Nil {
		t.Error("Expected non-nil ID, got Nil UUID")
	}
}
