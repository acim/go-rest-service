package rest

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/valve"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server runs REST and metrics servers.
type Server struct {
	serviceName   string
	router        *chi.Mux
	server        *http.Server
	metricsServer *http.Server
	valve         *valve.Valve
	logger        *zap.Logger
}

// NewServer creates new server.
func NewServer(serviceName string, serverPort, metricsPort int, router *chi.Mux, logger *zap.Logger) *Server {
	s := &Server{
		serviceName: serviceName,
		router:      router,
		valve:       valve.New(),
		logger:      logger,
	}
	s.server = &http.Server{Addr: ":" + strconv.Itoa(serverPort),
		Handler: chi.ServerBaseContext(s.valve.Context(), s.router)}
	s.metricsServer = &http.Server{Addr: ":" + strconv.Itoa(metricsPort), Handler: promhttp.Handler()}

	return s
}

// Run starts REST and metrics servers.
func (s *Server) Run() {
	go func() {
		s.logger.Info("metrics server", zap.String("name", s.serviceName), zap.String("port", s.metricsServer.Addr))
		if err := s.metricsServer.ListenAndServe(); err != nil {
			s.logger.Error("metrics server", zap.Error(err))
		}
	}()

	go s.shutdown()

	s.logger.Info("rest server", zap.String("name", s.serviceName), zap.String("port", s.server.Addr))

	if err := s.server.ListenAndServe(); err != nil {
		s.logger.Error("server", zap.Error(err))
	}
}

func (s *Server) shutdown() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	s.logger.Info("shutdown activated")

	if err := s.valve.Shutdown(20 * time.Second); err != nil {
		s.logger.Error("shutdown", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("shutdown", zap.Error(err))
	}

	select {
	case <-time.After(21 * time.Second):
		s.logger.Info("some connections not finished")
	case <-ctx.Done():
	}
}
