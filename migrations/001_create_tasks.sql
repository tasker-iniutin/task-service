-- +goose Up
CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    text TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL,
    expires_at TIMESTAMP NULL,
    CONSTRAINT tasks_title_not_empty CHECK (btrim(title) <> ''),
    CONSTRAINT tasks_status_valid CHECK (status IN (1, 2))
);

CREATE INDEX idx_tasks_user_id_id ON tasks (user_id, id);

-- +goose Down
DROP TABLE IF EXISTS tasks;
