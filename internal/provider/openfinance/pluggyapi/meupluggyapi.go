package pluggyapi

import (
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/sourcegraph/conc/iter"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
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

	iter.ForEach(env.Users, func(user *entity.User) {
		token, err := c.authenticate(
			user.PluggyClientID,
			user.PluggyClientSecret,
		)
		if err != nil {
			panic(err)
		}
		mu.Lock()
		conns[user.ID] = conn{
			accessToken: token,
			accountIDs:  user.PluggyAccountIDs,
		}
		mu.Unlock()
	})

	c.conns = conns

	return c
}

var _ openfinance.APIProvider = (*Client)(nil)
