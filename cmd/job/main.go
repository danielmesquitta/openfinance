package main

import "github.com/danielmesquitta/openfinance/internal/app"

func main() {
	err := app.Run()
	if err != nil {
		panic(err)
	}
}
