package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	abmiddleware "github.com/acim/go-rest-service/pkg/middleware"
	"github.com/acim/go-rest-service/pkg/rest"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

	r := router(c, logger)

	srv := rest.NewServer(c.ServiceName, c.ServerPort, c.MetricsPort, r, logger)
	srv.Run()
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

func router(c *config, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(abmiddleware.ZapLogger(logger))
	r.Use(abmiddleware.PromMetrics(c.ServiceName, nil))
	r.Use(middleware.DefaultCompress)
	r.Use(abmiddleware.RenderJSON)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		res := abmiddleware.ResponseFromContext(r.Context())
		res.SetPayload("hello world")
	})
	r.Get("/heavy", func(w http.ResponseWriter, r *http.Request) {
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

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return r
}
