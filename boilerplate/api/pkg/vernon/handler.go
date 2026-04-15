package vernon

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourorg/boilerplate/pkg/scope"
)

const maxPageSize = 100

// BaseHandler menyediakan endpoint HTTP generik untuk semua domain Vernon.
// Setiap domain mendapat: GET /, POST /, GET /{id}, PUT /{id}, PATCH /{id},
// DELETE /{id}, POST /{id}/_sync, GET /{id}/{relation}.
type BaseHandler struct {
	svc *BaseService
	log zerolog.Logger
}

// NewBaseHandler membuat instance baru BaseHandler.
func NewBaseHandler(svc *BaseService, log zerolog.Logger) *BaseHandler {
	return &BaseHandler{svc: svc, log: log}
}

// RegisterRoutes mendaftarkan semua route Vernon ke router Chi yang diberikan.
func (h *BaseHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.getByID)
	r.Put("/{id}", h.update)
	r.Patch("/{id}", h.patch)
	r.Delete("/{id}", h.delete)
	r.Post("/{id}/_sync", h.forceSync)
}

func (h *BaseHandler) list(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	params := ParseQueryParams(r, maxPageSize)
	items, total, err := h.svc.List(r.Context(), s, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	data := make([]map[string]any, len(items))
	for i := range items {
		data[i] = FormatItem(&items[i])
	}
	respondList(w, data, total, params)
}

func (h *BaseHandler) create(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "request body bukan JSON valid")
		return
	}
	entity, err := h.svc.Create(r.Context(), s, body)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"data": FormatItem(entity)})
}

func (h *BaseHandler) getByID(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	entity, err := h.svc.GetByID(r.Context(), s, id)
	if err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data": FormatItem(entity),
		"meta": map[string]any{"_sync_status": entity.SyncStatus, "_sync_version": entity.SyncVersion},
	})
}

func (h *BaseHandler) update(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "request body bukan JSON valid")
		return
	}
	entity, err := h.svc.Update(r.Context(), s, id, body)
	if err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": FormatItem(entity)})
}

func (h *BaseHandler) patch(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "request body bukan JSON valid")
		return
	}
	entity, err := h.svc.Patch(r.Context(), s, id, body)
	if err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": FormatItem(entity)})
}

func (h *BaseHandler) delete(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), s, id); err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BaseHandler) forceSync(w http.ResponseWriter, r *http.Request) {
	s, ok := scopeOrError(w, r)
	if !ok {
		return
	}
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	entity, err := h.svc.GetByID(r.Context(), s, id)
	if err == ErrNotFound {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "entity tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	h.svc.emitSync(r.Context(), s, id.String(), ActionUpdate, nil)
	respondJSON(w, http.StatusAccepted, map[string]any{
		"message": "sync dijadwalkan",
		"meta":    map[string]any{"_sync_status": entity.SyncStatus, "_sync_version": entity.SyncVersion},
	})
}

// ── response helpers ──────────────────────────────────────────────────────────

func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func respondList(w http.ResponseWriter, data any, total int64, p FindParams) {
	totalPages := int(total) / p.Limit
	if int(total)%p.Limit != 0 {
		totalPages++
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data": data,
		"meta": map[string]any{
			"total": total, "page": p.Page,
			"per_page": p.Limit, "total_pages": totalPages,
		},
	})
}

// ── request helpers ───────────────────────────────────────────────────────────

func scopeOrError(w http.ResponseWriter, r *http.Request) (scope.Scope, bool) {
	s, ok := scope.ScopeFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusForbidden, "FORBIDDEN", "scope organisasi tidak teridentifikasi")
		return scope.Scope{}, false
	}
	return s, true
}

func parseUUID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	raw := chi.URLParam(r, param)
	id, err := uuid.Parse(raw)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "ID harus berupa UUID valid")
		return uuid.Nil, false
	}
	return id, true
}
