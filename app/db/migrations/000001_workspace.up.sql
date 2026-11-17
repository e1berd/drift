CREATE TABLE projects (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    text        NOT NULL,
    name        text        NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
    path        text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX projects_owner_created ON projects (owner_id, created_at, id);

CREATE TABLE sessions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  uuid        NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    owner_id    text        NOT NULL,
    title       text        NOT NULL CHECK (length(title) BETWEEN 1 AND 200),
    agent       text        NOT NULL CHECK (agent IN ('claude', 'codex', 'antigravity')),
    status      text        NOT NULL DEFAULT 'idle' CHECK (status IN ('idle', 'running')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX sessions_project_updated ON sessions (project_id, updated_at DESC, id DESC);

CREATE TABLE messages (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  uuid        NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    role        text        NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    kind        text        NOT NULL CHECK (kind IN ('text', 'thinking', 'tool', 'note')),
    body        jsonb       NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX messages_session_created ON messages (session_id, created_at DESC, id DESC);

CREATE TABLE drafts (
    owner_id    text        NOT NULL,
    key         text        NOT NULL,
    text        text        NOT NULL,
    device_id   text        NOT NULL,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (owner_id, key)
);
