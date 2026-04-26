-- name: CreateChirp :one
INSERT INTO chirps (id, user_id, body, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps WHERE (sqlc.narg('user_id')::uuid is NULL or user_id = sqlc.narg('user_id')::uuid) ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1;
