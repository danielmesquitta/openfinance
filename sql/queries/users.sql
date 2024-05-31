-- name: GetUserByEmail :one
SELECT
  *
FROM
  users
WHERE
  email = $1;


-- name: GetUserByID :one
SELECT
  *
FROM
  users
WHERE
  id = $1;


-- name: GetUserWithSettingByID :one
SELECT
  users.*,
  settings.*
FROM
  users
  LEFT JOIN settings ON users.id = settings.user_id
WHERE
  users.id = $1;


-- name: CreateUser :exec
INSERT INTO
  users (id, email, updated_at)
VALUES
  ($1, $2, $3);
