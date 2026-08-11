package root

import (
	"embed"
)

//go:embed .env*
var EnvFile embed.FS

//go:embed config/ingest_profiles.json
var IngestProfilesFile embed.FS
