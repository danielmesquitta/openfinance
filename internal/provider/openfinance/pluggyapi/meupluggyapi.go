package pluggyapi

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance"
)

type conn struct {
	accessToken string
	accountIDs  []string
}

type Client struct {
	client                  *resty.Client
	conns                   map[string]conn
	accountSlots            chan struct{}
	maxConcurrentOperations int
}

func NewClient(env *config.Env) (*Client, error) {
	client := resty.New().SetBaseURL("https://api.pluggy.ai")

	c := &Client{
		client:                  client,
		accountSlots:            make(chan struct{}, env.MaxConcurrentOperations),
		maxConcurrentOperations: env.MaxConcurrentOperations,
	}

	mu := sync.Mutex{}
	conns := map[string]conn{}
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(env.MaxConcurrentOperations)

	for _, ingestProfile := range env.IngestProfiles {
		g.Go(func() error {
			token, err := c.authenticate(
				ctx,
				ingestProfile.PluggyClientID,
				ingestProfile.PluggyClientSecret,
			)
			if err != nil {
				return err
			}

			mu.Lock()
			conns[ingestProfile.ID] = conn{
				accessToken: token,
				accountIDs:  ingestProfile.PluggyAccountIDs,
			}
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("authenticate Pluggy ingest profiles: %w", err)
	}

	c.conns = conns

	return c, nil
}

var _ openfinance.APIProvider = (*Client)(nil)
