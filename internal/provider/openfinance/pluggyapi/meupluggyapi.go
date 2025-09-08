package pluggyapi

import (
	"context"
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
	client *resty.Client
	conns  map[string]conn
}

func NewClient(env *config.Env) *Client {
	client := resty.New().SetBaseURL("https://api.pluggy.ai")

	c := &Client{
		client: client,
	}

	mu := sync.Mutex{}
	conns := map[string]conn{}
	g, ctx := errgroup.WithContext(context.Background())

	for _, user := range env.Users {
		g.Go(func() error {
			token, err := c.authenticate(
				ctx,
				user.PluggyClientID,
				user.PluggyClientSecret,
			)
			if err != nil {
				return err
			}

			mu.Lock()
			conns[user.ID] = conn{
				accessToken: token,
				accountIDs:  user.PluggyAccountIDs,
			}
			mu.Unlock()

			return nil
		})
	}

	c.conns = conns

	return c
}

var _ openfinance.APIProvider = (*Client)(nil)
