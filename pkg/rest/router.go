package rest

import (
	"net/http"

	arcmw "github.com/acim/arc/pkg/middleware"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// DefaultRouter creates chi mux with default middlewares.
// Supply nil for allowedOrigins to turn off CORS middleware.
// Example string for allowedOrigins: "https://example.com".
func DefaultRouter(serviceName string, allowedOrigins []string, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(arcmw.ZapLogger(logger))
	r.Use(arcmw.PromMetrics(serviceName, nil))

	if len(allowedOrigins) > 0 {
		r.Use(getCORS(allowedOrigins).Handler)
	}
	// r.Use(middleware.DefaultCompress) compress will be done by ingress
	r.Use(arcmw.RenderJSON)
	r.Use(middleware.Recoverer)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		res := arcmw.ResponseFromContext(r.Context())
		res.SetStatusNotFound(http.StatusText(http.StatusNotFound))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		res := arcmw.ResponseFromContext(r.Context())
		res.SetStatus(http.StatusMethodNotAllowed).AddError(http.StatusText(http.StatusMethodNotAllowed))
	})

	return r
}

func getCORS(allowedOrigins []string) *cors.Cors {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	return cors.New(cors.Options{ //nolint:exhaustivestruct
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, //nolint:gomnd
	})
}
