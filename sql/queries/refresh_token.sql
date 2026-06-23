-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    NOW() + INTERVAL '60 days',
    NULL
)
RETURNING *;



-- name: UpdateRefreshTokenRefresh :one
UPDATE refresh_tokens
SET expires_at = NOW() + INTERVAL '60 days'
WHERE token = $1
RETURNING *;

-- name: UpdateRefreshTokenRevoke :one
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = $1
RETURNING *;