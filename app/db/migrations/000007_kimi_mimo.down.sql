alter table subscriptions drop constraint subscriptions_agent_check;
alter table subscriptions add constraint subscriptions_agent_check check (agent in ('claude', 'codex', 'antigravity'));
