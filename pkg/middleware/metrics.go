package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// PromMetrics returns middleware with Prometheus metrics.
func PromMetrics(serviceName string, buckets []float64) func(next http.Handler) http.Handler {
	if len(buckets) == 0 {
		buckets = []float64{100, 200, 500}
	}

	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{ //nolint:exhaustivestruct
			Name:        "requests_total",
			Help:        "Number of completed requests partitioned by status code, method and URI.",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"code", "method", "uri"},
	)
	prometheus.MustRegister(requests)

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{ //nolint:exhaustivestruct
		Name:        "request_duration_seconds",
		Help:        "Duration of requests completion partitioned by status code, method and URI.",
		ConstLabels: prometheus.Labels{"service": serviceName},
		Buckets:     buckets,
	},
		[]string{"code", "method", "uri"},
	)
	prometheus.MustRegister(duration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				requests.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.RequestURI).Inc()
				duration.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.RequestURI).
					Observe(time.Since(start).Seconds()) //nolint:gomnd
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
