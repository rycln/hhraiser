package gateways

import (
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
)

const defaultBaseURL = "https://hh.ru"

type Gateway struct {
	client  *httpclient.Client
	baseURL string
}

func NewGateway(client *httpclient.Client) *Gateway {
	return NewGatewayWithURL(client, defaultBaseURL)
}

func NewGatewayWithURL(client *httpclient.Client, baseURL string) *Gateway {
	return &Gateway{client: client, baseURL: baseURL}
}
