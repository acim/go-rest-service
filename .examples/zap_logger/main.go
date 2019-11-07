package main

import (
	"net/http"

	"github.com/acim/go-rest-server/pkg/middleware"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger, err = zap.NewDevelopment()
	r := chi.NewRouter()
	r.Use(middleware.ZapLogger(logger))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	http.ListenAndServe(":3000", r)
}
