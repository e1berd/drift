-- name: CreateAttachment :one
INSERT INTO attachments (id, owner_id, name, media_type, size, path) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, media_type, size;

-- name: ClaimAttachments :many
UPDATE attachments SET message_id = @message_id
WHERE owner_id = @owner_id AND message_id IS NULL AND id = ANY(@ids::uuid[])
RETURNING id, name, media_type, size;

-- name: GetAttachmentContent :one
SELECT name, media_type, path FROM attachments WHERE id = $1 AND owner_id = $2 AND path IS NOT NULL;

-- name: ListAttachmentData :many
SELECT id, name, media_type, size, path FROM attachments
WHERE owner_id = @owner_id AND id = ANY(@ids::uuid[])
ORDER BY array_position(@ids::uuid[], id);

-- name: DeleteStaleAttachments :execrows
DELETE FROM attachments WHERE message_id IS NULL AND created_at < now() - interval '7 days';

-- name: ListStoredAttachmentIDs :many
SELECT id FROM attachments WHERE id = ANY(@ids::uuid[]);

-- name: ListDatabaseAttachments :many
SELECT id, name, data FROM attachments WHERE path IS NULL ORDER BY created_at LIMIT 20;

-- name: MoveAttachmentToDisk :exec
UPDATE attachments SET path = @path, data = NULL WHERE id = @id;
