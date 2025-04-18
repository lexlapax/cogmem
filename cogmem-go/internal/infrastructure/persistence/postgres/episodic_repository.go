package postgres

import (
   "context"
   "fmt"

   "github.com/google/uuid"
   "github.com/jackc/pgx/v5/pgxpool"
   "github.com/pgvector/pgvector-go"

   "github.com/lexlapax/cogmem/internal/domain/entity"
   "github.com/lexlapax/cogmem/internal/domain/repository"
)

// PostgresEpisodicRepository implements repository.EpisodicRepository using PostgreSQL.
type PostgresEpisodicRepository struct {
   pool *pgxpool.Pool
}

// NewPostgresEpisodicRepository constructs a new Postgres-backed EpisodicRepository.
func NewPostgresEpisodicRepository(pool *pgxpool.Pool) *PostgresEpisodicRepository {
   return &PostgresEpisodicRepository{pool: pool}
}

// Save persists a new EpisodicMemory record.
func (r *PostgresEpisodicRepository) Save(ctx context.Context, mem *entity.EpisodicMemory) error {
   vec := pgvector.NewVector(mem.Embedding)
   const sql = `INSERT INTO episodic_memory
       (id, user_id, entity_id, content, embedding, timestamp, share_scope, last_accessed, accessibility_score)
       VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
   if _, err := r.pool.Exec(ctx, sql,
       mem.ID, mem.UserID, mem.EntityID,
       mem.Content, vec,
       mem.Timestamp, mem.ShareScope,
       mem.LastAccessed, mem.AccessibilityScore,
   ); err != nil {
       return fmt.Errorf("insert episodic_memory: %w", err)
   }
   return nil
}

// FindByID retrieves a memory by ID within the given partition.
func (r *PostgresEpisodicRepository) FindByID(ctx context.Context, id uuid.UUID, pCtx entity.PartitionContext) (*entity.EpisodicMemory, error) {
   const sql = `SELECT id, user_id, entity_id, content, embedding, timestamp, share_scope, last_accessed, accessibility_score
       FROM episodic_memory
       WHERE id = $1
         AND user_id = $2
         AND (entity_id IS NULL OR entity_id = $3)`
   row := r.pool.QueryRow(ctx, sql, id, pCtx.UserID, pCtx.EntityID)
   var mem entity.EpisodicMemory
   var vec pgvector.Vector
   err := row.Scan(
       &mem.ID, &mem.UserID, &mem.EntityID,
       &mem.Content, &vec,
       &mem.Timestamp, &mem.ShareScope,
       &mem.LastAccessed, &mem.AccessibilityScore,
   )
   if err != nil {
       return nil, err
   }
   mem.Embedding = vec.Slice()
   return &mem, nil
}

// FindRecent returns the most recent memories in the partition, up to limit.
func (r *PostgresEpisodicRepository) FindRecent(ctx context.Context, limit int, pCtx entity.PartitionContext) ([]*entity.EpisodicMemory, error) {
   const sql = `SELECT id, user_id, entity_id, content, embedding, timestamp, share_scope, last_accessed, accessibility_score
       FROM episodic_memory
       WHERE user_id = $1
         AND (entity_id IS NULL OR entity_id = $2)
       ORDER BY timestamp DESC
       LIMIT $3`
   rows, err := r.pool.Query(ctx, sql, pCtx.UserID, pCtx.EntityID, limit)
   if err != nil {
       return nil, fmt.Errorf("query recent: %w", err)
   }
   defer rows.Close()
   var results []*entity.EpisodicMemory
   for rows.Next() {
       var mem entity.EpisodicMemory
       var vec pgvector.Vector
       if err := rows.Scan(
           &mem.ID, &mem.UserID, &mem.EntityID,
           &mem.Content, &vec,
           &mem.Timestamp, &mem.ShareScope,
           &mem.LastAccessed, &mem.AccessibilityScore,
       ); err != nil {
           return nil, fmt.Errorf("scan recent: %w", err)
       }
       mem.Embedding = vec.Slice()
       results = append(results, &mem)
   }
   return results, nil
}

// FindByVector performs a vector similarity search within the given partition.
func (r *PostgresEpisodicRepository) FindByVector(ctx context.Context, vector []float32, limit int, pCtx entity.PartitionContext) ([]*entity.EpisodicMemory, error) {
   const sql = `SELECT id, user_id, entity_id, content, embedding, timestamp, share_scope, last_accessed, accessibility_score
       FROM episodic_memory
       WHERE user_id = $1
         AND (entity_id IS NULL OR entity_id = $2)
       ORDER BY embedding <=> $3
       LIMIT $4`
   vecParam := pgvector.NewVector(vector)
   rows, err := r.pool.Query(ctx, sql, pCtx.UserID, pCtx.EntityID, vecParam, limit)
   if err != nil {
       return nil, fmt.Errorf("vector search: %w", err)
   }
   defer rows.Close()
   var results []*entity.EpisodicMemory
   for rows.Next() {
       var mem entity.EpisodicMemory
       var vec pgvector.Vector
       if err := rows.Scan(
           &mem.ID, &mem.UserID, &mem.EntityID,
           &mem.Content, &vec,
           &mem.Timestamp, &mem.ShareScope,
           &mem.LastAccessed, &mem.AccessibilityScore,
       ); err != nil {
           return nil, fmt.Errorf("scan vector row: %w", err)
       }
       mem.Embedding = vec.Slice()
       results = append(results, &mem)
   }
   return results, nil
}

// Ensure PostgresEpisodicRepository satisfies the interface
var _ repository.EpisodicRepository = (*PostgresEpisodicRepository)(nil)