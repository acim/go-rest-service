package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	abmiddleware "github.com/acim/go-rest-service/pkg/middleware"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/valve"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	v := valve.New()
	baseCtx := v.Context()

	logger, err := logger(c)
	if err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	r := router(c, logger)

	go func() {
		srv := http.Server{Addr: ":" + strconv.Itoa(c.MetricsPort), Handler: promhttp.Handler()}
		logger.Info("metrics server", zap.String("name", c.ServiceName), zap.Int("port", c.MetricsPort))
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("metrics server", zap.Error(err))
		}
	}()

	srv := http.Server{Addr: ":" + strconv.Itoa(c.ServerPort), Handler: chi.ServerBaseContext(baseCtx, r)}

	go shutdown(&srv, v, logger)

	logger.Info("server", zap.String("name", c.ServiceName), zap.Int("port", c.ServerPort))

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server", zap.Error(err))
	}
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

func shutdown(srv *http.Server, v *valve.Valve, logger *zap.Logger) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	logger.Info("shutdown activated")

	if err := v.Shutdown(20 * time.Second); err != nil {
		logger.Error("shutdown", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown", zap.Error(err))
	}

	select {
	case <-time.After(21 * time.Second):
		logger.Info("some connections not finished")
	case <-ctx.Done():
	}
}
