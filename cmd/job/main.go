package main

import (
	"github.com/danielmesquitta/openfinance/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
