package main

import (
	"net/http"

	"github.com/acim/arc/pkg/middleware"
	"github.com/go-chi/chi"
)

func handler(w http.ResponseWriter, r *http.Request) {
	res := middleware.ResponseFromContext(r.Context())
	payload := &struct {
		Name     string
		Language string
	}{"example", "golang"}
	res.SetPayload(payload)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RenderJSON)

	r.Get("/", handler)

	http.ListenAndServe(":3000", r)
}
