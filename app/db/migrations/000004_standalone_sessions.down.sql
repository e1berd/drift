drop index sessions_owner_standalone;

delete from sessions where project_id is null;

alter table sessions alter column project_id set not null;
