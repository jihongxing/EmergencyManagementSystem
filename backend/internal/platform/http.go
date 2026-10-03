package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func NewHandler(checks ...func(context.Context) error) http.Handler {
	return NewHandlerWithRoutes(nil, checks...)
}

func NewHandlerWithRoutes(routes map[string]http.Handler, checks ...func(context.Context) error) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
		}{Status: "alive"})
	})
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		status := "not_ready"
		if len(checks) == 1 && checks[0] != nil && checks[0](ctx) == nil {
			status = "ready"
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	})
	for pattern, handler := range routes {
		mux.Handle(pattern, handler)
	}
	return mux
}
