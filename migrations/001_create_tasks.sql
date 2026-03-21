-- +goose Up
CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    text TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS tasks;
