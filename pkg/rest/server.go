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

// Server ...
type Server struct {
	serviceName string
	serverPort  string
	metricsPort string
	router      *chi.Mux
	srv         *http.Server
	valve       *valve.Valve
	logger      *zap.Logger
}

// NewServer creates new server.
func NewServer(serviceName string, serverPort, metricsPort int, router *chi.Mux, logger *zap.Logger) *Server {
	s := &Server{
		serviceName: serviceName,
		serverPort:  ":" + strconv.Itoa(serverPort),
		metricsPort: ":" + strconv.Itoa(metricsPort),
		router:      router,
		logger:      logger,
		valve:       valve.New(),
	}
	return s
}

// Run ...
func (s *Server) Run() {
	go func() {
		srv := http.Server{Addr: s.metricsPort, Handler: promhttp.Handler()}
		s.logger.Info("metrics server", zap.String("name", s.serviceName), zap.String("port", s.metricsPort))
		if err := srv.ListenAndServe(); err != nil {
			s.logger.Error("metrics server", zap.Error(err))
		}
	}()

	s.srv = &http.Server{Addr: s.serverPort, Handler: chi.ServerBaseContext(s.valve.Context(), s.router)}

	go s.shutdown()

	s.logger.Info("server", zap.String("name", s.serviceName), zap.String("port", s.serverPort))

	if err := s.srv.ListenAndServe(); err != nil {
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

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error("shutdown", zap.Error(err))
	}

	select {
	case <-time.After(21 * time.Second):
		s.logger.Info("some connections not finished")
	case <-ctx.Done():
	}
}
