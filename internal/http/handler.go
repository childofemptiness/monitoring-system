package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"url-monitor/internal/monitor"
)

type Handler struct {
	service *monitor.Service
}

type CreateMonitorRequest struct {
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
}

type CreateMonitorResponse struct {
	ID              int64
	URL             string
	IntervalSeconds int
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	LastCheckAt     *time.Time
	NextCheckAt     *time.Time
}

type MonitorItem struct {
	ID              int64      `json:"id"`
	URL             string     `json:"url"`
	IntervalSeconds int        `json:"interval_seconds"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	LastCheckAt     *time.Time `json:"last_check_at"`
	NextCheckAt     *time.Time `json:"next_check_at"`
}

type ListMonitorsMeta struct {
	Total int `json:"total"`
}

type ListMonitorsResponse struct {
	Data []MonitorItem    `json:"data"`
	Meta ListMonitorsMeta `json:"meta"`
}

func NewHandler(service *monitor.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) CreateMonitor(w http.ResponseWriter, r *http.Request) {
	var req CreateMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	input := monitor.CreateMonitorInput{
		URL:             req.URL,
		IntervalSeconds: req.IntervalSeconds,
	}

	created, err := h.service.CreateMonitor(r.Context(), input)
	if err != nil {
		writeError(w, statusFromError(err), err.Error())
		return
	}

	response := CreateMonitorResponse{
		ID:              created.ID,
		URL:             created.URL,
		IntervalSeconds: created.IntervalSeconds,
		CreatedAt:       created.CreatedAt,
		UpdatedAt:       created.UpdatedAt,
		LastCheckAt:     created.LastCheckAt,
		NextCheckAt:     created.NextCheckAt,
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) ListMonitors(w http.ResponseWriter, r *http.Request) {
	monitors, err := h.service.ListMonitors(r.Context())
	if err != nil {
		writeError(w, statusFromError(err), err.Error())
		return
	}

	items := make([]MonitorItem, 0, len(monitors))
	for _, m := range monitors {
		items = append(items, MonitorItem{
			ID:              m.ID,
			URL:             m.URL,
			IntervalSeconds: m.IntervalSeconds,
			CreatedAt:       m.CreatedAt,
			UpdatedAt:       m.UpdatedAt,
			LastCheckAt:     m.LastCheckAt,
			NextCheckAt:     m.NextCheckAt,
		})
	}

	writeJSON(w, http.StatusOK, ListMonitorsResponse{
		Data: items,
		Meta: ListMonitorsMeta{
			Total: len(monitors),
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func statusFromError(err error) int {
	switch {
	case errors.Is(err, monitor.ErrInvalidURL):
		return http.StatusBadRequest
	case errors.Is(err, monitor.ErrInvalidInterval):
		return http.StatusBadRequest
	case errors.Is(err, monitor.ErrMonitorAlreadyExists):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
