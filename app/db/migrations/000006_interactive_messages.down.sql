drop index messages_scheduled_continuations;
delete from messages where kind in ('question', 'limit');
alter table messages drop constraint messages_kind_check;
alter table messages add constraint messages_kind_check
    check (kind in ('text', 'thinking', 'tool', 'note'));
