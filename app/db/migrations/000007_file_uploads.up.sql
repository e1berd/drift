alter table attachments alter column data drop not null;
alter table attachments alter column size type bigint;
alter table attachments add column path text;
alter table attachments add constraint attachments_content check (data is not null or path is not null);

create table instance_settings (
    singleton        boolean     primary key default true check (singleton),
    max_upload_bytes bigint      not null default 2000000000 check (max_upload_bytes > 0),
    update_time      timestamptz not null default now()
);

insert into instance_settings default values;
