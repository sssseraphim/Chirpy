-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
gen_random_uuid(),
Now(),
Now(),
$1,
$2
) RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: ChangeUserInfo :one
UPDATE users
SET email = $1,
hashed_password = $2,
updated_at = Now()
WHERE id = $3
RETURNING *;

-- name: MakeUserRed :exec
UPDATE users
SET chirpy_red = true,
updated_at = Now()
WHERE id = $1;

