alter table sessions drop constraint sessions_agent_check;
alter table sessions add constraint sessions_agent_check check (agent in ('claude', 'codex', 'antigravity', 'kimi', 'mimo'));
