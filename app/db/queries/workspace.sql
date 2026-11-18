-- name: ListProjects :many
SELECT p.*, (SELECT count(*) FROM sessions s WHERE s.project_id = p.id) AS session_count
FROM projects p
WHERE p.owner_id = $1
  AND (sqlc.narg(after_created)::timestamptz IS NULL
       OR (p.created_at, p.id) > (sqlc.narg(after_created)::timestamptz, sqlc.narg(after_id)::uuid))
ORDER BY p.created_at, p.id
LIMIT $2;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1 AND owner_id = $2;

-- name: CreateProject :one
INSERT INTO projects (owner_id, name, path) VALUES ($1, $2, $3) RETURNING *;

-- name: RenameProject :one
UPDATE projects SET name = $3, updated_at = now() WHERE id = $1 AND owner_id = $2 RETURNING *;

-- name: DeleteProject :execrows
DELETE FROM projects WHERE id = $1 AND owner_id = $2;

-- name: CountSessions :one
SELECT count(*) FROM sessions WHERE project_id = $1;

-- name: ListSessions :many
SELECT * FROM sessions
WHERE project_id = $1 AND owner_id = $2
  AND (sqlc.narg(before_updated)::timestamptz IS NULL
       OR (updated_at, id) < (sqlc.narg(before_updated)::timestamptz, sqlc.narg(before_id)::uuid))
ORDER BY updated_at DESC, id DESC
LIMIT $3;

-- name: ListStandaloneSessions :many
SELECT * FROM sessions
WHERE owner_id = $1 AND project_id IS NULL
  AND (sqlc.narg(before_updated)::timestamptz IS NULL
       OR (updated_at, id) < (sqlc.narg(before_updated)::timestamptz, sqlc.narg(before_id)::uuid))
ORDER BY updated_at DESC, id DESC
LIMIT $2;

-- name: GetSession :one
SELECT * FROM sessions WHERE id = $1 AND owner_id = $2;

-- name: CreateSession :one
INSERT INTO sessions (project_id, owner_id, title, agent) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateSession :one
UPDATE sessions
SET title = coalesce(sqlc.narg(title), title),
    agent = coalesce(sqlc.narg(agent), agent),
    agent_session_id = CASE WHEN coalesce(sqlc.narg(agent), agent) = agent THEN agent_session_id ELSE '' END,
    updated_at = now()
WHERE id = $1 AND owner_id = $2
RETURNING *;

-- name: TouchSession :exec
UPDATE sessions SET updated_at = now() WHERE id = $1;

-- name: DeleteSession :execrows
DELETE FROM sessions WHERE id = $1 AND owner_id = $2;

-- name: ListMessages :many
SELECT * FROM messages
WHERE session_id = $1
  AND (sqlc.narg(before_created)::timestamptz IS NULL
       OR (created_at, id) < (sqlc.narg(before_created)::timestamptz, sqlc.narg(before_id)::uuid))
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: CreateMessage :one
INSERT INTO messages (session_id, role, kind, body) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListDrafts :many
SELECT * FROM drafts WHERE owner_id = $1 ORDER BY updated_at DESC;

-- name: UpsertDraft :one
INSERT INTO drafts (owner_id, key, text, device_id) VALUES ($1, $2, $3, $4)
ON CONFLICT (owner_id, key) DO UPDATE
SET text = excluded.text, device_id = excluded.device_id, updated_at = now()
RETURNING *;

-- name: DeleteDraft :execrows
DELETE FROM drafts WHERE owner_id = $1 AND key = $2;

-- name: DeleteDraftsForProject :exec
DELETE FROM drafts WHERE owner_id = $1 AND key = 'new:' || sqlc.arg(project_id)::text;

-- name: DeleteDraftsForSession :exec
DELETE FROM drafts WHERE owner_id = $1 AND key = 'session:' || sqlc.arg(session_id)::text;

-- name: SetSessionStatus :one
UPDATE sessions SET status = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: SetAgentSessionID :exec
UPDATE sessions SET agent_session_id = $2 WHERE id = $1;

-- name: ResetRunningSessions :exec
UPDATE sessions SET status = 'idle' WHERE status = 'running';

-- name: GetRunContext :one
SELECT s.*, p.path AS project_path FROM sessions s LEFT JOIN projects p ON p.id = s.project_id
WHERE s.id = $1 AND s.owner_id = $2;

-- name: UpdateMessageBody :one
UPDATE messages SET body = $2 WHERE id = $1 RETURNING *;
