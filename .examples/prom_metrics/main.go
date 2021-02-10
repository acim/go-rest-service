package main

import (
	"net/http"

	"github.com/acim/arc/pkg/middleware"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.PromMetrics("my-service", nil))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	// Run another HTTP servers just for metrics scraping
	go func() {
		s := &http.Server{Addr: ":3001", Handler: promhttp.Handler()}
		s.ListenAndServe()
	}()

	// Run the main server for which we collect metrics
	http.ListenAndServe(":3000", r)
}
