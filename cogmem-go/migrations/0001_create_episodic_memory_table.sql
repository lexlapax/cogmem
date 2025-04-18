-- Migration 0001: Create episodic_memory table with pgvector embedding

-- Phase 1 storage schema with pgvector
-- Ensure the pgvector extension is available
CREATE EXTENSION IF NOT EXISTS vector;
-- Drop existing table (e.g., from previous JSONB schema) to allow re-creation
DROP TABLE IF EXISTS episodic_memory;
CREATE TABLE episodic_memory (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    entity_id UUID,
    content TEXT NOT NULL,
    embedding vector(1536) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    share_scope TEXT NOT NULL,
    last_accessed TIMESTAMPTZ NOT NULL,
    accessibility_score DOUBLE PRECISION NOT NULL
);

-- Index for retrieving recent memories efficiently
CREATE INDEX IF NOT EXISTS idx_episodic_recent
    ON episodic_memory (user_id, timestamp DESC);
-- Index for embedding similarity search (cosine distance)
CREATE INDEX IF NOT EXISTS idx_episodic_embedding
    ON episodic_memory USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);