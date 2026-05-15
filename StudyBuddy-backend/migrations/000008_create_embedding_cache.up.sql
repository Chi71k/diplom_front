CREATE TABLE user_embedding_cache (
    user_id     UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    embedding   FLOAT8[] NOT NULL,
    profile_hash TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
