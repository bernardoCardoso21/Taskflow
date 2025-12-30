BEGIN;

DROP TABLE IF EXISTS tasks;

CREATE TABLE tasks (
                       id          UUID PRIMARY KEY,
                       project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                       title       TEXT NOT NULL,
                       completed   BOOLEAN NOT NULL DEFAULT FALSE,
                       created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_project_created
    ON tasks (project_id, created_at DESC, id DESC);

CREATE INDEX idx_tasks_project_completed
    ON tasks (project_id, completed);

COMMIT;
