package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/lexlapax/cogmem/internal/domain/entity"
)

// EpisodicRepository defines persistence operations for EpisodicMemory entities.
type EpisodicRepository interface {
	// Save persists a new memory item.
	Save(ctx context.Context, mem *entity.EpisodicMemory) error
	// FindByID retrieves a memory by its ID within a given partition.
	FindByID(ctx context.Context, id uuid.UUID, pCtx entity.PartitionContext) (*entity.EpisodicMemory, error)
	// FindByVector finds memories similar to the given embedding vector, limited by count, within a partition.
	FindByVector(ctx context.Context, vector []float32, limit int, pCtx entity.PartitionContext) ([]*entity.EpisodicMemory, error)
	// FindRecent retrieves the most recent memories up to limit, within a partition.
	FindRecent(ctx context.Context, limit int, pCtx entity.PartitionContext) ([]*entity.EpisodicMemory, error)
}
