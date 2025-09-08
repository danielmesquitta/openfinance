package main

import (
	"log"

	"github.com/danielmesquitta/openfinance/internal/app/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatalf("failed to execute cli: %v", err)
	}
}
