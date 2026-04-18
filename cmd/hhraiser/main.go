package main

import (
	"context"
	"log"
	"time"

	"github.com/rycln/hhraiser/internal/config"
	"github.com/rycln/hhraiser/internal/domain"
	"github.com/rycln/hhraiser/internal/infrastructure/gateways"
	"github.com/rycln/hhraiser/internal/infrastructure/httpclient"
	"github.com/rycln/hhraiser/internal/usecases"
)

const (
	baseURL       = "https://hh.ru"
	reqTimeout    = 10 * time.Second
	clientTimeout = 30 * time.Second
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	config.InitLogger(cfg.LogLevel)

	client, err := httpclient.New(clientTimeout)
	if err != nil {
		log.Fatal(err)
	}

	gateway := gateways.NewGateway(client, baseURL)
	auth := usecases.NewAuth(gateway, reqTimeout)
	raise := usecases.NewRaise(gateway, reqTimeout)

	session, err := auth.Authenticate(context.Background(), domain.NewCredentials(cfg.Phone, cfg.Password))
	if err != nil {
		log.Fatal(err)
	}

	err = raise.RaiseResume(context.Background(), domain.NewResume(cfg.ResumeID, cfg.ResumeTitle), session)
	if err != nil {
		log.Fatal(err)
	}
}
