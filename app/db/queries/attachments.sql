-- name: CreateAttachment :one
INSERT INTO attachments (owner_id, name, media_type, size, data) VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, media_type, size;

-- name: ClaimAttachments :many
UPDATE attachments SET message_id = @message_id
WHERE owner_id = @owner_id AND message_id IS NULL AND id = ANY(@ids::uuid[])
RETURNING id, name, media_type, size;

-- name: GetAttachmentContent :one
SELECT media_type, data FROM attachments WHERE id = $1 AND owner_id = $2;

-- name: ListAttachmentData :many
SELECT id, name, media_type, data FROM attachments
WHERE owner_id = @owner_id AND id = ANY(@ids::uuid[])
ORDER BY array_position(@ids::uuid[], id);

-- name: DeleteStaleAttachments :execrows
DELETE FROM attachments WHERE message_id IS NULL AND created_at < now() - interval '1 day';
