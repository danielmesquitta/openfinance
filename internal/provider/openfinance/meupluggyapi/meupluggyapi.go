package meupluggyapi

import (
	"sync"

	"github.com/go-resty/resty/v2"

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
	client := resty.New().SetBaseURL("https//api.pluggy.ai")

	c := &Client{
		client: client,
	}

	mu := sync.Mutex{}
	jobsCount := len(env.Users)
	conns := make(map[string]conn, jobsCount)
	wg := sync.WaitGroup{}
	wg.Add(jobsCount)

	for _, user := range env.Users {
		go func() {
			defer wg.Done()
			token, err := c.authenticate(
				user.MeuPluggyClientID,
				user.MeuPluggyClientSecret,
			)
			if err != nil {
				panic(err)
			}
			mu.Lock()
			conns[user.ID] = conn{
				accessToken: token,
				accountIDs:  user.MeuPluggyAccountIDs,
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	c.conns = conns

	return c
}

var _ openfinance.APIProvider = (*Client)(nil)
