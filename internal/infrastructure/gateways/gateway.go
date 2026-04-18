package gateways

import (
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
)

type Gateway struct {
	client  *httpclient.Client
	baseURL string
}

func NewGateway(client *httpclient.Client, baseURL string) *Gateway {
	return &Gateway{
		client:  client,
		baseURL: baseURL,
	}
}
