package main

import (
	"log"

	"github.com/danielmesquitta/openfinance/internal/app/cli"
	"github.com/danielmesquitta/openfinance/internal/domain/errs"
)

func main() {
	if err := cli.Run(); err != nil {
		switch v := err.(type) {
		case *errs.Err:
			log.Fatalln(v, v.StackTrace)
		default:
			log.Fatalln(err)
		}
	}
}
