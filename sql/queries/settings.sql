-- name: CreateSetting :one
INSERT INTO
  settings (
    id,
    notion_token,
    notion_page_id,
    meu_pluggy_client_id,
    meu_pluggy_client_secret,
    meu_pluggy_account_ids,
    user_id,
    updated_at
  )
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
  *;


-- name: UpdateSetting :one
UPDATE
  settings
SET
  notion_token = $2,
  notion_page_id = $3,
  meu_pluggy_client_id = $4,
  meu_pluggy_client_secret = $5,
  meu_pluggy_account_ids = $6,
  updated_at = $7
WHERE
  id = $1
RETURNING
  *;


-- name: ListSettings :many
SELECT
  *
FROM
  settings;