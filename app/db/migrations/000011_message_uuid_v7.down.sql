drop index if exists messages_session_prompts;
alter table messages alter column id set default gen_random_uuid();
drop function if exists uuid_generate_v7();
