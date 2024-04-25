package main

import (
	"github.com/danielmesquitta/openfinance/internal/app/rest"
)

func main() {
	if err := rest.Start(); err != nil {
		panic(err)
	}
}
