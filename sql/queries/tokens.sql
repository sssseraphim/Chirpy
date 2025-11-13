-- name: AddNewRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at)
VALUES (
		$1,
		Now(),
		Now(),
		$2,
		$3
		)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * from refresh_tokens
WHERE token = $1
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = Now(),
revoked_at = Now()
WHERE token = $1;


