package app

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/danielmesquitta/asyncloop"

	"github.com/danielmesquitta/openfinance/config"
	"github.com/danielmesquitta/openfinance/internal/service/meupluggyapi"
)

func Run() error {
	e := config.LoadEnv()

	fmt.Printf("%+v", *e)

	meupluggyAPIClient := meupluggyapi.NewClient(
		e.MeuPluggyClientID,
		e.MeuPluggyClientSecret,
	)

	if err := meupluggyAPIClient.Authenticate(); err != nil {
		return fmt.Errorf("error authenticating: %v", err)
	}

	now := time.Now()
	startOfMonth := time.Date(
		now.Year(),
		now.Month(),
		1,
		0,
		0,
		0,
		0,
		now.Location(),
	)

	mu := sync.Mutex{}
	errs := []error{}

	requestsData := []string{
		e.MeuPluggyBankAccountID,
		e.MeuPluggyCreditAccountID,
	}

	asyncloop.Loop(requestsData, func(i int, v string) {
		ts, err := meupluggyAPIClient.ListTransactions(
			v,
			&startOfMonth,
			nil,
		)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
		}
	})

	if len(errs) > 0 {
		fmtErrs := []string{}
		for _, err := range errs {
			fmtErrs = append(fmtErrs, err.Error())
		}
		return fmt.Errorf(
			"error fetching transactions:\n%v",
			strings.Join(fmtErrs, "\n\n"),
		)
	}

	return nil
}
