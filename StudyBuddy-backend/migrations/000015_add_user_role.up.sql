ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'student';

ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('student', 'admin'));

UPDATE users SET role = 'student' WHERE role IS NULL;
