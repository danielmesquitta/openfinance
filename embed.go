package root //nolint:revive

import (
	"embed"
)

//go:embed .env*
var EnvFile embed.FS

//go:embed config/*.json
var Config embed.FS
