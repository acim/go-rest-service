package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

type responseKey struct{}

// Response ...
type Response struct {
	statusCode int
	headers    http.Header
	payload    interface{}
	errors     []string
}

// ResponseFromContext returns response from context.
func ResponseFromContext(ctx context.Context) *Response {
	response := ctx.Value(responseKey{}).(*Response)

	return response
}

// SetStatus sets status code to response. If status code is not set if will default to http.StatusOK.
func (r *Response) SetStatus(statusCode int) *Response {
	r.statusCode = statusCode

	return r
}

// SetStatusBadRequest sets status code to http.StatusBadRequest.
func (r *Response) SetStatusBadRequest(err string) *Response {
	r.statusCode = http.StatusBadRequest

	if err != "" {
		r.AddError(err)
	}

	return r
}

// SetStatusForbidden sets status code to http.StatusForbidden.
func (r *Response) SetStatusForbidden(err string) *Response {
	r.statusCode = http.StatusForbidden

	if err != "" {
		r.AddError(err)
	}

	return r
}

// SetStatusInternalServerError sets status code to http.StatusInternalServerError.
func (r *Response) SetStatusInternalServerError(err string) *Response {
	r.statusCode = http.StatusInternalServerError

	if err != "" {
		r.AddError(err)
	}

	return r
}

// SetStatusNotFound sets status code to http.StatusNotFound.
func (r *Response) SetStatusNotFound(err string) *Response {
	r.statusCode = http.StatusNotFound

	if err != "" {
		r.AddError(err)
	}

	return r
}

// SetStatusAccepted sets status code to http.StatusAccepted.
func (r *Response) SetStatusAccepted() *Response {
	r.statusCode = http.StatusAccepted

	return r
}

// SetHeader sets header to response.
func (r *Response) SetHeader(key, value string) *Response {
	r.headers.Set(key, value)

	return r
}

// AddHeader ads headers to response.
func (r *Response) AddHeader(key, value string) *Response {
	r.headers.Add(key, value)

	return r
}

// SetPayload sets payload to response.
func (r *Response) SetPayload(payload interface{}) *Response {
	r.payload = payload

	return r
}

// AddError adds error to response.
func (r *Response) AddError(err string) *Response {
	r.errors = append(r.errors, err)

	return r
}

// RenderJSON middleware is used to inject response object in context and later render it as JSON.
func RenderJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), responseKey{}, newResponse())

		// ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(w, r.WithContext(ctx))

		if res, ok := ctx.Value(responseKey{}).(*Response); ok { //nolint:nestif
			res.headers.Set("Content-Type", "application/json")

			var body []byte
			var err error

			if res.payload != nil || len(res.errors) > 0 {
				b := &response{
					Payload: res.payload,
					Errors:  res.errors,
				}

				body, err = json.Marshal(b)
				if err != nil {
					res.SetStatusInternalServerError("Error rendering response body")
				}
			}

			for hk, hsv := range res.headers {
				for _, hv := range hsv {
					w.Header().Add(hk, hv)
				}
			}

			if res.statusCode > 0 {
				w.WriteHeader(res.statusCode)
			}

			if res.statusCode == http.StatusNoContent {
				return
			}

			_, _ = w.Write(body)
		}
	})
}

func newResponse() *Response {
	return &Response{ //nolint:exhaustivestruct
		headers: make(http.Header, 1),
	}
}

type response struct {
	Payload interface{} `json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}
