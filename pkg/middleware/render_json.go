package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/middleware"
)

type responseKey struct{}

// Response ...
type Response struct {
	statusCode int
	headers    map[string]string
	Payload    interface{} `json:"data,omitempty"`
	Errors     []string    `json:"errors,omitempty"`
}

// ResponseFromContext returns response from context.
func ResponseFromContext(ctx context.Context) *Response {
	response := ctx.Value(responseKey{}).(*Response)
	return response
}

// SetStatusCode sets status code to response. If status code is not set if will default to http.StatusOK.
func (r *Response) SetStatusCode(statusCode int) *Response {
	r.statusCode = statusCode
	return r
}

// SetHeader sets header to response.
func (r *Response) SetHeader(key, value string) *Response {
	r.headers[key] = value
	return r
}

// SetHeaders sets headers to response.
func (r *Response) SetHeaders(headers map[string]string) *Response {
	r.headers = headers
	return r
}

// SetPayload sets payload to response.
func (r *Response) SetPayload(payload interface{}) *Response {
	r.Payload = payload
	return r
}

// AddError adds error to response.
func (r *Response) AddError(err string) *Response {
	r.Errors = append(r.Errors, err)
	return r
}

// RenderJSON middleware is used to inject response object in context and later render it as JSON.
func RenderJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), responseKey{}, newResponse())

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r.WithContext(ctx))

		if resp, ok := ctx.Value(responseKey{}).(*Response); ok {
			if ww.Status() == http.StatusNotFound {
				resp.AddError(http.StatusText(http.StatusNotFound))
			}

			body, err := json.Marshal(resp)
			if err != nil {
				ww.WriteHeader(http.StatusInternalServerError)
				return
			}

			resp.headers["Content-Type"] = "application/json"
			for k, v := range resp.headers {
				ww.Header().Set(k, v)
			}

			if resp.statusCode > 0 {
				ww.WriteHeader(resp.statusCode)
			}

			ww.Write(body)
		}
	})
}

func newResponse() *Response {
	return &Response{
		headers: make(map[string]string, 1),
	}
}
