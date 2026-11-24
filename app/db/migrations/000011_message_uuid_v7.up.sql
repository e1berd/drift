create or replace function uuid_generate_v7() returns uuid
language sql volatile parallel safe
as $$
    select encode(
        set_bit(
            set_bit(
                overlay(
                    uuid_send(gen_random_uuid())
                    placing substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3)
                    from 1 for 6
                ),
                52, 1
            ),
            53, 1
        ),
        'hex'
    )::uuid
$$;

alter table messages alter column id set default uuid_generate_v7();

create index if not exists messages_session_prompts on messages (session_id, created_at desc, id desc)
    where role = 'user';
