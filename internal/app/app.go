package app

import (
	"fmt"
	"strings"
	"sync"

	"github.com/danielmesquitta/openfinance/config"
	"github.com/danielmesquitta/openfinance/internal/service/meupluggyapi"
)

func Run() error {
	e := config.LoadEnv()

	meupluggyAPIClient := meupluggyapi.NewClient(
		e.MeuPluggyClientID,
		e.MeuPluggyClientSecret,
	)

	if err := meupluggyAPIClient.Authenticate(); err != nil {
		return fmt.Errorf("error authenticating: %v", err)
	}

	ch := make(chan *meupluggyapi.ListTransactionsResponse, 2)
	mu := sync.Mutex{}
	errs := []error{}

	go func() {
		bankTransactions, err := meupluggyAPIClient.ListTransactions(
			e.MeuPluggyBankAccountID,
			nil,
			nil,
		)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
		ch <- bankTransactions
	}()

	go func() {
		creditCardTransactions, err := meupluggyAPIClient.ListTransactions(
			e.MeuPluggyCreditAccountID,
			nil,
			nil,
		)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
		ch <- creditCardTransactions
	}()

	for transactions := range ch {
		if transactions == nil {
			continue
		}
	}

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
