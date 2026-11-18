alter table sessions alter column project_id drop not null;

create index sessions_owner_standalone on sessions (owner_id, updated_at desc, id desc) where project_id is null;
