package pgcontainer

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielmesquitta/openfinance/test/integration/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type options struct {
	seeds []Seed
}

type Option func(*options)

func WithSeeds(seeds ...Seed) Option {
	return func(o *options) {
		o.seeds = seeds
	}
}

func NewPgContainer(
	ctx context.Context,
	opts ...Option,
) (dbConnURL string, terminate func()) {
	containerOpts := &options{}

	for _, opt := range opts {
		opt(containerOpts)
	}

	dbName := "testdb"
	dbUser := "test"
	dbPassword := "test"

	migrationFiles, err := getMigrationFiles()
	if err != nil {
		log.Fatalf("failed to get migration files: %s", err)
	}

	seedFiles, err := getSeedFiles(containerOpts.seeds...)
	if err != nil {
		log.Fatalf("failed to get seed files: %s", err)
	}

	scriptFiles := []string{}
	scriptFiles = append(scriptFiles, migrationFiles...)
	scriptFiles = append(scriptFiles, seedFiles...)

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithInitScripts(scriptFiles...),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	terminate = func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}

	dbConnURL, err = postgresContainer.ConnectionString(ctx)
	if err != nil {
		defer terminate()
		log.Fatalf("failed to get connection string: %s", err)
	}

	dbConnURL += "sslmode=disable"

	return dbConnURL, terminate
}

func getSeedFiles(seeds ...Seed) ([]string, error) {
	var files []string

	for _, s := range seeds {
		seed := string(s)
		if !strings.Contains(seed, ".sql") {
			seed = seed + ".sql"
		}
		if err := filepath.Walk(
			filepath.Join(container.CWD, "sql", "seeds", seed),
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && info.Name() == seed {
					files = append(files, path)
				}
				return nil
			},
		); err != nil {
			return nil, err
		}
	}

	return files, nil
}

func getMigrationFiles() ([]string, error) {
	var files []string

	if err := filepath.Walk(
		filepath.Join(container.CWD, "sql", "migrations"),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == "migration.sql" {
				files = append(files, path)
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	return files, nil
}
