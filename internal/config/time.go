package config

import "time"

const (
	AmericaSaoPaulo = "America/Sao_Paulo"
)

func init() {
	loc, err := time.LoadLocation(AmericaSaoPaulo)
	if err != nil {
		panic(err)
	}
	time.Local = loc
}
