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
		addr := ":" + strconv.Itoa(c.MetricsPort)
		srv := http.Server{Addr: addr, Handler: promhttp.Handler()}
		logger.Info("metrics server", zap.String("name", c.ServiceName), zap.Int("port", c.MetricsPort))
		if err := srv.ListenAndServe(); err != nil {
		logger.Error("metrics server", zap.Error(err))
	}
	}()

	addr := ":" + strconv.Itoa(c.ServerPort)
	srv := http.Server{Addr: addr, Handler: chi.ServerBaseContext(baseCtx, r)}

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

		valve.Lever(r.Context()).Open()
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

	v.Shutdown(20 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	select {
	case <-time.After(21 * time.Second):
		logger.Info("some connections not finished")
	case <-ctx.Done():
	}
}
