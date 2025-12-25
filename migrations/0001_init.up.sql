CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       email TEXT NOT NULL UNIQUE,
                       password_hash TEXT NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE projects (
                          id UUID PRIMARY KEY,
                          user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          name TEXT NOT NULL,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                          updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX projects_user_id_idx ON projects(user_id);

CREATE TABLE tasks (
                       id UUID PRIMARY KEY,
                       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                       project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                       title TEXT NOT NULL,
                       description TEXT,
                       status TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo','doing','done')),
                       due_date TIMESTAMPTZ,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX tasks_user_id_idx ON tasks(user_id);
CREATE INDEX tasks_project_id_idx ON tasks(project_id);
CREATE INDEX tasks_created_id_idx ON tasks(created_at, id);
