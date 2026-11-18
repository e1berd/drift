create table attachments (
    id          uuid primary key default gen_random_uuid(),
    owner_id    text        not null,
    message_id  uuid references messages (id) on delete cascade,
    name        text        not null check (length(name) between 1 and 255),
    media_type  text        not null,
    size        integer     not null check (size > 0),
    data        bytea       not null,
    created_at  timestamptz not null default now()
);

create index attachments_message on attachments (message_id) where message_id is not null;
create index attachments_unclaimed on attachments (created_at) where message_id is null;
