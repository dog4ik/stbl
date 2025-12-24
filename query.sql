-- name: CreateMapping :one
INSERT INTO gateway_id_mapping (token, merchant_private_key, gateway_id) VALUES (?, ?, ?) RETURNING *;

-- name: GetMapping :one
SELECT * FROM gateway_id_mapping
WHERE gateway_id = ? LIMIT 1;

-- name: UpsertTokenCache :exec
INSERT INTO token_cache (
    credentials_hash,
    access_token,
    refresh_token,
    access_refreshed_at,
    refresh_refreshed_at
)
VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT(credentials_hash) DO UPDATE SET
    access_token        = excluded.access_token,
    refresh_token       = excluded.refresh_token,
    access_refreshed_at = CURRENT_TIMESTAMP,
    refresh_refreshed_at = CURRENT_TIMESTAMP;

-- name: GetTokenCache :one
SELECT
    access_token,
    refresh_token,
    access_refreshed_at,
    refresh_refreshed_at
FROM token_cache
WHERE credentials_hash = ?;
