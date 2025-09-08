package root

import (
	"embed"
)

//go:embed .env*
var EnvFile embed.FS

//go:embed users.json
var UsersFile embed.FS
