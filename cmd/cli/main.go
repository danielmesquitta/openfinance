package main

import "github.com/danielmesquitta/openfinance/internal/app/cli"

func main() {
	if err := cli.Run(); err != nil {
		panic(err)
	}
}
