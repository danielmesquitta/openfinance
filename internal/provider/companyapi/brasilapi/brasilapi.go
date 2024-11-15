package brasilapi

import (
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	*resty.Client
}

func NewClient() *Client {
	c := resty.New().SetBaseURL("https://brasilapi.com.br")

	return &Client{
		Client: c,
	}
}

var _ companyapi.APIProvider = (*Client)(nil)
