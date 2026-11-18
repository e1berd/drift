alter table messages drop constraint messages_kind_check;
alter table messages add constraint messages_kind_check
    check (kind in ('text', 'thinking', 'tool', 'note', 'question', 'limit'));

create index messages_scheduled_continuations on messages (session_id)
    where kind = 'limit' and body ->> 'status' = 'scheduled';
