package root

import (
	"embed"
)

//go:embed .env*
var EnvFile embed.FS

//go:embed config/sync_profiles.json
var SyncProfilesFile embed.FS
