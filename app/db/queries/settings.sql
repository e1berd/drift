-- name: GetInstanceSettings :one
SELECT max_upload_bytes, update_time FROM instance_settings;

-- name: UpdateInstanceSettings :one
UPDATE instance_settings SET max_upload_bytes = @max_upload_bytes, update_time = now()
RETURNING max_upload_bytes, update_time;
