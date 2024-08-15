package meupluggyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance"
)

type _conn struct {
	accessToken string
	accountIDs  []string
}

type Client struct {
	baseURL url.URL
	conns   map[string]_conn
}

func NewClient(env *config.Env) *Client {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "api.pluggy.ai",
	}

	c := &Client{
		baseURL: baseURL,
	}

	jobsCount := len(env.Users)
	conns := make(map[string]_conn, jobsCount)
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
			conns[user.ID] = _conn{
				accessToken: token,
				accountIDs:  user.MeuPluggyAccountIDs,
			}
		}()
	}

	c.conns = conns

	return c
}

type _errorMessage struct {
	Code            int    `json:"code"`
	Message         string `json:"message"`
	CodeDescription string `json:"codeDescription"`
	Data            any    `json:"data"`
}

func parseResError(res *http.Response) error {
	jsonData := _errorMessage{}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&jsonData); err != nil {
		return entity.NewErr(fmt.Sprintf(
			"error requesting %s: %s",
			res.Request.URL,
			res.Status,
		))
	}

	return entity.NewErr(fmt.Sprintf(
		"error requesting %s: %s %v",
		res.Request.URL,
		res.Status,
		jsonData,
	))
}

var _ openfinance.APIProvider = (*Client)(nil)
