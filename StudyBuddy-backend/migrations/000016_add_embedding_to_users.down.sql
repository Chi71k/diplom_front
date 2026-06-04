DROP INDEX IF EXISTS users_embedding_ivfflat_idx;
ALTER TABLE users DROP COLUMN IF EXISTS embedding_vector;
