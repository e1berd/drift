create table subscriptions (
    id          uuid primary key default gen_random_uuid(),
    owner_id    text        not null,
    agent       text        not null check (agent in ('claude', 'codex', 'antigravity')),
    method      text        not null check (method in ('login', 'token')),
    state       text        not null default 'pending' check (state in ('pending', 'active', 'failed')),
    account     text        not null default '',
    plan        text        not null default '',
    error       text        not null default '',
    is_current  boolean     not null default false,
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()
);

create index subscriptions_owner_created on subscriptions (owner_id, created_at, id);
create unique index subscriptions_one_current on subscriptions (owner_id, agent) where is_current;
