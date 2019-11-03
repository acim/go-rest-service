package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	abmiddleware "github.com/acim/go-rest-service/pkg/middleware"
	"github.com/acim/go-rest-service/pkg/rest"
	"github.com/go-chi/valve"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type config struct {
	ServiceName string `default:"rest-service"`
	ServerPort  int    `default:"3000"`
	MetricsPort int    `default:"3001"`
	Environment string `default:"dev"`
}

func main() {
	c := &config{}
	if err := envconfig.Process("", c); err != nil {
		log.Fatalf("failed parsing environment variables: %v", err)
	}

	logger, err := logger(c)
	if err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	router := rest.DefaultRouter(c.ServiceName, logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		res := abmiddleware.ResponseFromContext(r.Context())
		res.SetPayload("hello world")
	})

	router.Get("/heavy", func(w http.ResponseWriter, r *http.Request) {
		err := valve.Lever(r.Context()).Open()
		if err != nil {
			logger.Error("open valve lever", zap.Error(err))
		}
		defer valve.Lever(r.Context()).Close()

		select {
		case <-valve.Lever(r.Context()).Stop():
			logger.Info("valve closed, finishing")
		case <-time.After(2 * time.Second):
			// Do some heave lifting
			time.Sleep(2 * time.Second)
		}

		res := abmiddleware.ResponseFromContext(r.Context())
		res.SetPayload("all done")
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	rest.NewServer(c.ServiceName, c.ServerPort, c.MetricsPort, router, logger).Run()
}

func logger(c *config) (*zap.Logger, error) {
	var logger *zap.Logger

	var err error

	switch c.Environment {
	case "prod":
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		logger, err = config.Build()
	case "dev":
		logger, err = zap.NewDevelopment()
	default:
		return nil, fmt.Errorf("logger: unknown environment: '%s'", c.Environment)
	}

	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}

	return logger, nil
}
