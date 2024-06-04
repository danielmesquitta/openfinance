-- name: GetUserByEmail :one
SELECT
  *
FROM
  users
WHERE
  email = $1;


-- name: GetFullUserByID :one
SELECT
  users.*,
  settings.*
FROM
  users
  LEFT JOIN settings ON users.id = settings.user_id
WHERE
  users.id = $1;


-- name: CreateUser :one
INSERT INTO
  users (id, email, updated_at)
VALUES
  ($1, $2, $3)
RETURNING
  *;