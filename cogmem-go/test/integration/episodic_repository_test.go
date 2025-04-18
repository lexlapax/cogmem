package integration

import (
   "context"
   "os"
   "testing"
   "time"

   "github.com/google/uuid"
   "github.com/jackc/pgx/v5/pgxpool"

   "github.com/lexlapax/cogmem/internal/domain/entity"
   "github.com/lexlapax/cogmem/internal/infrastructure/config"
   "github.com/lexlapax/cogmem/internal/infrastructure/persistence/postgres"
)

func TestPostgresEpisodicRepository_SaveFind(t *testing.T) {
   // Load configuration
   cfg, err := config.LoadConfig()
   if err != nil {
       t.Fatalf("config load failed: %v", err)
   }
   // Connect to Postgres
   ctx := context.Background()
   // Retry connecting to allow container startup
   var pool *pgxpool.Pool
   for i := 0; i < 10; i++ {
       pool, err = pgxpool.New(ctx, cfg.DatabaseURL)
       if err == nil {
           break
       }
       time.Sleep(1 * time.Second)
   }
   if err != nil {
       t.Fatalf("pgxpool connect failed: %v", err)
   }
   defer pool.Close()
   // Apply migration
   sqlBytes, err := os.ReadFile("../../migrations/0001_create_episodic_memory_table.sql")
   if err != nil {
       t.Fatalf("read migration: %v", err)
   }
   if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
       t.Fatalf("apply migration: %v", err)
   }
   // Clear table
   if _, err := pool.Exec(ctx, "TRUNCATE episodic_memory"); err != nil {
       t.Fatalf("truncate table: %v", err)
   }
   repo := postgres.NewPostgresEpisodicRepository(pool)
   // Prepare test data
   userID := uuid.New()
   pCtx := entity.PartitionContext{UserID: userID}
   now := time.Now().UTC().Truncate(time.Second)
   // Create embeddings matching configured dimension
   dim := cfg.EmbeddingDim
   e1 := make([]float32, dim)
   for i := range e1 {
       e1[i] = 0.1
   }
   mem1 := entity.NewEpisodicMemory(userID, nil, "first", e1, now, "user")
   // Save memory 1
   if err := repo.Save(ctx, mem1); err != nil {
       t.Fatalf("save mem1: %v", err)
   }
   // Save memory 2 with later timestamp, distinct embedding
   e2 := make([]float32, dim)
   for i := range e2 {
       e2[i] = 0.9
   }
   mem2 := entity.NewEpisodicMemory(userID, nil, "second", e2, now.Add(time.Minute), "user")
   if err := repo.Save(ctx, mem2); err != nil {
       t.Fatalf("save mem2: %v", err)
   }
   // FindByID
   got1, err := repo.FindByID(ctx, mem1.ID, pCtx)
   if err != nil {
       t.Fatalf("FindByID error: %v", err)
   }
   if got1.ID != mem1.ID || got1.Content != "first" {
       t.Errorf("FindByID returned %+v, want %+v", got1, mem1)
   }
   // FindRecent
   recent, err := repo.FindRecent(ctx, 10, pCtx)
   if err != nil {
       t.Fatalf("FindRecent error: %v", err)
   }
   if len(recent) != 2 {
       t.Fatalf("FindRecent length = %d; want 2", len(recent))
   }
   if recent[0].ID != mem2.ID {
       t.Errorf("Most recent ID = %v; want %v", recent[0].ID, mem2.ID)
   }
   if recent[1].ID != mem1.ID {
       t.Errorf("Second recent ID = %v; want %v", recent[1].ID, mem1.ID)
   }
   // Partition isolation
   otherCtx := entity.PartitionContext{UserID: uuid.New()}
   none, err := repo.FindRecent(ctx, 10, otherCtx)
   if err != nil {
       t.Fatalf("FindRecent otherCtx error: %v", err)
   }
   if len(none) != 0 {
       t.Errorf("Expected no rows for other partition, got %d", len(none))
   }
   // Vector search (limit 1)
   vecResults, err := repo.FindByVector(ctx, e1, 1, pCtx)
   if err != nil {
       t.Fatalf("FindByVector error: %v", err)
   }
   if len(vecResults) != 1 {
       t.Fatalf("FindByVector length = %d; want 1", len(vecResults))
   }
   if vecResults[0].ID != mem1.ID {
       t.Errorf("FindByVector returned ID %v; want %v", vecResults[0].ID, mem1.ID)
   }
}