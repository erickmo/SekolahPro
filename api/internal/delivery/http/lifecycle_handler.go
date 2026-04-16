package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	createlifecycle "github.com/yourorg/boilerplate/internal/command/create_lifecycle_event"
	processdeathsettlement "github.com/yourorg/boilerplate/internal/command/process_death_settlement"
	processexpulsion "github.com/yourorg/boilerplate/internal/command/process_expulsion"
	processresignation "github.com/yourorg/boilerplate/internal/command/process_resignation"
	reactivatemember "github.com/yourorg/boilerplate/internal/command/reactivate_member"
	getlifecycle "github.com/yourorg/boilerplate/internal/query/get_lifecycle_event_by_id"
	listlifecycle "github.com/yourorg/boilerplate/internal/query/list_lifecycle_events"
	getmemberstatushistory "github.com/yourorg/boilerplate/internal/query/get_member_status_history"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// LifecycleHandler menangani HTTP request untuk domain Lifecycle.
type LifecycleHandler struct {
	BaseHandler
}

var lifecycleSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at": "created_at",
		"applied_at": "applied_at",
		"current_status": "current_status",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewLifecycleHandler membuat instance LifecycleHandler baru.
func NewLifecycleHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *LifecycleHandler {
	return &LifecycleHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Lifecycle.
func (h *LifecycleHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/lifecycle", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}/resign", h.ProcessResignation)
		r.Put("/{id}/expel", h.ProcessExpulsion)
		r.Put("/{id}/death-settlement", h.ProcessDeathSettlement)
		r.Put("/{id}/reactivate", h.ReactivateMember)
		r.Get("/members/{nasabahId}/history", h.GetMemberStatusHistory)
	})
}

// List godoc
// @Summary      List lifecycle events
// @Tags         lifecycle
// @Produce      json
// @Param        current_status   query  string  false  "Filter by status"
// @Param        nasabah_id       query  string  false  "Filter by nasabah ID"
// @Param        last_transition  query  string  false  "Filter by transition type"
// @Success      200  {object}  listlifecycle.Result
// @Router       /api/v1/lifecycle [get]
func (h *LifecycleHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, lifecycleSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listlifecycle.Result](r.Context(), h.queryBus, listlifecycle.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetByID godoc
// @Summary      Get lifecycle event by ID
// @Tags         lifecycle
// @Produce      json
// @Param        id   path      string  true  "Lifecycle ID"
// @Success      200  {object}  getlifecycle.Result
// @Router       /api/v1/lifecycle/{id} [get]
func (h *LifecycleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getlifecycle.Result](r.Context(), h.queryBus, getlifecycle.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create lifecycle event
// @Tags         lifecycle
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/lifecycle [post]
func (h *LifecycleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NasabahID        string `json:"nasabah_id"`
		MembershipNo     string `json:"membership_no"`
		TransitionReason string `json:"transition_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	nasabahID, err := uuid.Parse(req.NasabahID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "nasabah_id tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createlifecycle.Command{
		NasabahID:        nasabahID,
		MembershipNo:     req.MembershipNo,
		TransitionReason: req.TransitionReason,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// ProcessResignation godoc
// @Summary      Process member resignation
// @Tags         lifecycle
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Lifecycle ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/lifecycle/{id}/resign [put]
func (h *LifecycleHandler) ProcessResignation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), processresignation.Command{
		ID:     id,
		Reason: req.Reason,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ProcessExpulsion godoc
// @Summary      Process member expulsion
// @Tags         lifecycle
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Lifecycle ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/lifecycle/{id}/expel [put]
func (h *LifecycleHandler) ProcessExpulsion(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		RevokedBy string `json:"revoked_by"`
		Reason    string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	revokedBy, err := uuid.Parse(req.RevokedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "revoked_by tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), processexpulsion.Command{
		ID:        id,
		RevokedBy: revokedBy,
		Reason:    req.Reason,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ProcessDeathSettlement godoc
// @Summary      Process death settlement
// @Tags         lifecycle
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Lifecycle ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/lifecycle/{id}/death-settlement [put]
func (h *LifecycleHandler) ProcessDeathSettlement(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), processdeathsettlement.Command{
		ID:     id,
		Reason: req.Reason,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ReactivateMember godoc
// @Summary      Reactivate member
// @Tags         lifecycle
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Lifecycle ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/lifecycle/{id}/reactivate [put]
func (h *LifecycleHandler) ReactivateMember(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), reactivatemember.Command{
		ID:     id,
		Reason: req.Reason,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// GetMemberStatusHistory godoc
// @Summary      Get member status history
// @Tags         lifecycle
// @Produce      json
// @Param        nasabahId  path      string  true  "Nasabah ID"
// @Success      200  {object}  getmemberstatushistory.Result
// @Router       /api/v1/lifecycle/members/{nasabahId}/history [get]
func (h *LifecycleHandler) GetMemberStatusHistory(w http.ResponseWriter, r *http.Request) {
	nasabahID, err := uuid.Parse(chi.URLParam(r, "nasabahId"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "nasabah_id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getmemberstatushistory.Result](r.Context(), h.queryBus, getmemberstatushistory.Query{
		NasabahID: nasabahID,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// parseListFilter membaca filter dari query parameters.
func (h *LifecycleHandler) parseListFilter(r *http.Request) lifecycle.LifecycleFilter {
	var filter lifecycle.LifecycleFilter
	q := r.URL.Query()

	if st := q.Get("current_status"); st != "" {
		s := lifecycle.MembershipStatus(st)
		filter.CurrentStatus = &s
	}
	if nid := q.Get("nasabah_id"); nid != "" {
		id, err := uuid.Parse(nid)
		if err == nil {
			filter.NasabahID = &id
		}
	}
	if lt := q.Get("last_transition"); lt != "" {
		t := lifecycle.TransitionType(lt)
		filter.LastTransition = &t
	}

	return filter
}
