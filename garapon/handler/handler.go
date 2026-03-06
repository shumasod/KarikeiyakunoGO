// Package handler provides the HTTP presentation layer.
// It translates HTTP requests into service calls and service results into HTTP responses.
package handler

import (
	"encoding/json"
	"net/http"

	"garapon/model"
	"garapon/service"
)

// Handler holds a reference to the LotteryService and exposes HTTP methods.
type Handler struct {
	svc service.LotteryService
}

// New constructs a Handler with the provided LotteryService.
func New(svc service.LotteryService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all API and UI routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.Home)
	mux.HandleFunc("/api/draw", h.Draw)
	mux.HandleFunc("/api/history", h.History)
	mux.HandleFunc("/api/stats", h.Stats)
	mux.HandleFunc("/api/prizes", h.Prizes)
}

// writeJSON encodes v as JSON and writes it with the given status code.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// At this point the header is already sent; log only.
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}

// writeError sends a JSON error body with the given HTTP status.
func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, model.ErrorResponse{Error: msg})
}

// requireMethod checks that r.Method matches method; on mismatch it writes 405
// and returns false so the caller can return early.
func (h *Handler) requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	h.writeError(w, http.StatusMethodNotAllowed,
		"このエンドポイントは "+method+" のみ受け付けます")
	return false
}

// Home serves the main HTML page.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.writeError(w, http.StatusNotFound, "ページが見つかりません")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(indexHTML)) //nolint:errcheck
}

// Draw handles GET /api/draw — performs one lottery draw.
func (h *Handler) Draw(w http.ResponseWriter, r *http.Request) {
	if !h.requireMethod(w, r, http.MethodGet) {
		return
	}
	result, err := h.svc.Draw()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, result)
}

// History handles GET /api/history — returns the draw history.
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	if !h.requireMethod(w, r, http.MethodGet) {
		return
	}
	h.writeJSON(w, http.StatusOK, h.svc.History())
}

// Stats handles GET /api/stats — returns aggregate statistics.
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if !h.requireMethod(w, r, http.MethodGet) {
		return
	}
	h.writeJSON(w, http.StatusOK, h.svc.Stats())
}

// Prizes handles GET /api/prizes — returns current prize table with rotation info.
func (h *Handler) Prizes(w http.ResponseWriter, r *http.Request) {
	if !h.requireMethod(w, r, http.MethodGet) {
		return
	}
	h.writeJSON(w, http.StatusOK, h.svc.Prizes())
}
