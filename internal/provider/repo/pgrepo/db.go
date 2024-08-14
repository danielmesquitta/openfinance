package pgrepo

import (
	"database/sql"
	"time"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/provider/db/pgdb"
)

const timeout = 60 * time.Second

func NewPgDBConn(
	env *config.Env,
) *pgdb.Queries {
	dbConn, err := sql.Open(
		"postgres",
		env.DBConnection,
	)
	if err != nil {
		panic(err)
	}

	if err := dbConn.Ping(); err != nil {
		panic(err)
	}

	return pgdb.New(dbConn)
}
