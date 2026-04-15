package http

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// SortConfig mendefinisikan konfigurasi sorting yang diizinkan per domain handler.
type SortConfig struct {
	AllowedFields map[string]string
	DefaultField  string
	DefaultOrder  string
}

// BaseHandler adalah struct yang di-embed oleh semua domain handler.
type BaseHandler struct {
	commandBus *commandbus.CommandBus
	queryBus   *querybus.QueryBus
}

// NewBaseHandler membuat instance BaseHandler baru.
func NewBaseHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) BaseHandler {
	return BaseHandler{commandBus: cb, queryBus: qb}
}

// ParsePagination membaca parameter pagination dan memvalidasi sort field terhadap whitelist.
func (h *BaseHandler) ParsePagination(r *http.Request, cfg SortConfig) pagination.ListParams {
	p := pagination.ParseFromRequest(r)

	if col, ok := cfg.AllowedFields[p.SortBy]; ok {
		p.SortBy = col
	} else {
		p.SortBy = cfg.DefaultField
	}

	if r.URL.Query().Get("order") == "" && cfg.DefaultOrder != "" {
		p.Order = cfg.DefaultOrder
	}

	return p
}

// RespondJSON menulis response JSON dengan status code yang diberikan.
func (h *BaseHandler) RespondJSON(w http.ResponseWriter, status int, body any) {
	respondJSON(w, status, body)
}

// RespondError menulis error response dengan format {"error": "pesan"}.
func (h *BaseHandler) RespondError(w http.ResponseWriter, status int, msg string) {
	respondError(w, status, msg)
}

func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
