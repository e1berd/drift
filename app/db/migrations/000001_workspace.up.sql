create table projects (
    id          uuid primary key default gen_random_uuid(),
    owner_id    text        not null,
    name        text        not null check (length(name) between 1 and 120),
    path        text        not null,
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()
);

create index projects_owner_created on projects (owner_id, created_at, id);

create table sessions (
    id          uuid primary key default gen_random_uuid(),
    project_id  uuid        not null references projects (id) on delete cascade,
    owner_id    text        not null,
    title       text        not null check (length(title) between 1 and 200),
    agent       text        not null check (agent in ('claude', 'codex', 'antigravity')),
    status      text        not null default 'idle' check (status in ('idle', 'running')),
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()
);

create index sessions_project_updated on sessions (project_id, updated_at desc, id desc);

create table messages (
    id          uuid primary key default gen_random_uuid(),
    session_id  uuid        not null references sessions (id) on delete cascade,
    role        text        not null check (role in ('user', 'assistant', 'system')),
    kind        text        not null check (kind in ('text', 'thinking', 'tool', 'note')),
    body        jsonb       not null,
    created_at  timestamptz not null default now()
);

create index messages_session_created on messages (session_id, created_at desc, id desc);

create table drafts (
    owner_id    text        not null,
    key         text        not null,
    text        text        not null,
    device_id   text        not null,
    updated_at  timestamptz not null default now(),
    primary key (owner_id, key)
);
