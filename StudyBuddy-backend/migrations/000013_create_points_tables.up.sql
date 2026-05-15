CREATE TABLE point_transactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount      INT NOT NULL,
    reason      TEXT NOT NULL,
    source_key  TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON point_transactions(user_id);

CREATE UNIQUE INDEX ux_point_transactions_dedupe
ON point_transactions (user_id, reason, source_key)
WHERE source_key <> '';

CREATE MATERIALIZED VIEW leaderboard_points AS
SELECT user_id, SUM(amount)::bigint AS total_points
FROM point_transactions
GROUP BY user_id;

CREATE UNIQUE INDEX leaderboard_points_user_uidx ON leaderboard_points(user_id);
