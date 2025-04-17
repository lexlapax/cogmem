package entity

import (
	"testing"

	"github.com/google/uuid"
)

// TestPartitionContextAssignment verifies that PartitionContext fields are set correctly.
func TestPartitionContextAssignment(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()
	pCtx := PartitionContext{
		UserID:   userID,
		EntityID: &entityID,
	}
	if pCtx.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", pCtx.UserID, userID)
	}
	if pCtx.EntityID == nil {
		t.Fatal("Expected non-nil EntityID")
	}
	if *pCtx.EntityID != entityID {
		t.Errorf("EntityID mismatch: got %v, want %v", *pCtx.EntityID, entityID)
	}
}
