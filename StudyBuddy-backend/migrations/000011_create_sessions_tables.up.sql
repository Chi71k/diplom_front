CREATE TABLE study_sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title        TEXT NOT NULL,
    organizer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id    UUID,
    group_id     UUID,
    start_time   TIMESTAMPTZ NOT NULL,
    end_time     TIMESTAMPTZ NOT NULL,
    timezone     TEXT NOT NULL DEFAULT 'UTC',
    status       TEXT NOT NULL DEFAULT 'proposed',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_study_sessions_organizer_id ON study_sessions(organizer_id);
CREATE INDEX idx_study_sessions_start_time ON study_sessions(start_time);

CREATE TABLE session_participants (
    session_id    UUID NOT NULL REFERENCES study_sessions(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    confirmed     BOOLEAN NOT NULL DEFAULT FALSE,
    gcal_event_id TEXT,
    PRIMARY KEY (session_id, user_id)
);

CREATE INDEX idx_session_participants_user_id ON session_participants(user_id);
CREATE INDEX ON session_participants(session_id);
