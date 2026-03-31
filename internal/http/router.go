package http

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /monitors", h.ListMonitors)
	mux.HandleFunc("POST /monitors", h.CreateMonitor)
	mux.Handle("/metrics", h.metrics.Handler())

	return mux
}
