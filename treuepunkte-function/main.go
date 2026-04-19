package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"treuepunkte/internal/config"
	httpx "treuepunkte/internal/http"
	"treuepunkte/internal/service"
	"treuepunkte/internal/storage"
)

type application struct {
	cfg     config.Config
	router  http.Handler
	adapter *httpadapter.HandlerAdapterV2
}

func bootstrap() (*application, error) {
	cfg := config.FromEnv()

	db, err := storage.OpenMySQL(cfg.AppEnv, cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	if err != nil {
		return nil, err
	}

	loyalty := service.NewLoyaltyService(db)
	handlers := &httpx.Handlers{Loyalty: loyalty}
	router := httpx.Router(handlers)
	adapter := httpadapter.NewV2(router)

	return &application{
		cfg:     cfg,
		router:  router,
		adapter: adapter,
	}, nil
}

func main() {
	app, err := bootstrap()
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	if app.cfg.AppEnv == "local" {
		log.Printf("starting local HTTP server on :%s", app.cfg.AppPort)
		if err := http.ListenAndServe(":"+app.cfg.AppPort, app.router); err != nil {
			log.Fatalf("http server failed: %v", err)
		}
		return
	}

	lambda.Start(func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return app.adapter.ProxyWithContext(ctx, event)
	})
}