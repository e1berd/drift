-- name: ListSubscriptions :many
SELECT * FROM subscriptions WHERE owner_id = $1 ORDER BY created_at, id;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = $1 AND owner_id = $2;

-- name: CreateSubscription :one
INSERT INTO subscriptions (owner_id, agent, method) VALUES ($1, $2, $3) RETURNING *;

-- name: MarkSubscriptionPending :one
UPDATE subscriptions SET state = 'pending', error = '', updated_at = now()
WHERE id = $1 AND owner_id = $2
RETURNING *;

-- name: MarkSubscriptionActive :one
UPDATE subscriptions
SET state = 'active', error = '', account = sqlc.arg(account), plan = sqlc.arg(plan), updated_at = now(),
    is_current = subscriptions.is_current OR NOT EXISTS (
        SELECT 1 FROM subscriptions other
        WHERE other.owner_id = sqlc.arg(owner_id) AND other.agent = subscriptions.agent AND other.is_current
    )
WHERE subscriptions.id = sqlc.arg(id) AND subscriptions.owner_id = sqlc.arg(owner_id)
RETURNING *;

-- name: MarkSubscriptionFailed :one
UPDATE subscriptions SET state = 'failed', error = $3, updated_at = now()
WHERE id = $1 AND owner_id = $2
RETURNING *;

-- name: FailInterruptedLogins :exec
UPDATE subscriptions SET state = 'failed', error = 'interrupted by a server restart', updated_at = now()
WHERE state = 'pending';

-- name: ClearCurrentSubscription :exec
UPDATE subscriptions SET is_current = false, updated_at = now()
WHERE owner_id = $1 AND agent = $2 AND is_current;

-- name: SetCurrentSubscription :one
UPDATE subscriptions SET is_current = true, updated_at = now()
WHERE id = $1 AND owner_id = $2 AND state = 'active'
RETURNING *;

-- name: PromoteNextSubscription :one
UPDATE subscriptions SET is_current = true, updated_at = now()
WHERE id = (
    SELECT next.id FROM subscriptions next
    WHERE next.owner_id = $1 AND next.agent = $2 AND next.state = 'active'
    ORDER BY next.created_at DESC
    LIMIT 1
)
RETURNING *;

-- name: DeleteSubscription :one
DELETE FROM subscriptions WHERE id = $1 AND owner_id = $2 RETURNING *;
