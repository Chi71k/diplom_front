-- Enable pgvector if not already enabled
CREATE EXTENSION IF NOT EXISTS vector;

-- Add 768-dimensional embedding column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS
  embedding_vector vector(768);

-- Create an IVFFlat index for fast cosine similarity search
-- lists=100 is appropriate for up to ~1M rows
CREATE INDEX IF NOT EXISTS users_embedding_ivfflat_idx
  ON users
  USING ivfflat (embedding_vector vector_cosine_ops)
  WITH (lists = 100);
