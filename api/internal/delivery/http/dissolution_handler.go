package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	canceldissolution "github.com/yourorg/boilerplate/internal/command/cancel_dissolution"
	completedissolution "github.com/yourorg/boilerplate/internal/command/complete_dissolution"
	createdissolution "github.com/yourorg/boilerplate/internal/command/create_dissolution"
	deletedissolution "github.com/yourorg/boilerplate/internal/command/delete_dissolution"
	updatedissolutionstage "github.com/yourorg/boilerplate/internal/command/update_dissolution_stage"
	getdissolution "github.com/yourorg/boilerplate/internal/query/get_dissolution_by_id"
	listdissolutions "github.com/yourorg/boilerplate/internal/query/list_dissolutions"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// DissolutionHandler menangani HTTP request untuk domain Dissolution.
type DissolutionHandler struct {
	BaseHandler
}

var dissolutionSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at":     "created_at",
		"effective_date": "effective_date",
		"stage":          "stage",
		"status":         "status",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewDissolutionHandler membuat instance DissolutionHandler baru.
func NewDissolutionHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *DissolutionHandler {
	return &DissolutionHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Dissolution.
func (h *DissolutionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/dissolution", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}/stage", h.UpdateStage)
		r.Put("/{id}/complete", h.Complete)
		r.Put("/{id}/cancel", h.Cancel)
		r.Delete("/{id}", h.Delete)
	})
}

// List godoc
// @Summary      List dissolution processes
// @Tags         dissolution
// @Produce      json
// @Param        dissolution_type  query  string  false  "Filter by dissolution type"
// @Param        status            query  string  false  "Filter by status"
// @Param        stage             query  string  false  "Filter by stage"
// @Success      200  {object}  listdissolutions.Result
// @Router       /api/v1/dissolution [get]
func (h *DissolutionHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, dissolutionSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listdissolutions.Result](r.Context(), h.queryBus, listdissolutions.Query{
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
// @Summary      Get dissolution by ID
// @Tags         dissolution
// @Produce      json
// @Param        id   path      string  true  "Dissolution ID"
// @Success      200  {object}  getdissolution.Result
// @Router       /api/v1/dissolution/{id} [get]
func (h *DissolutionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getdissolution.Result](r.Context(), h.queryBus, getdissolution.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create dissolution process
// @Tags         dissolution
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/dissolution [post]
func (h *DissolutionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DissolutionType string     `json:"dissolution_type"`
		RatMeetingID    *string    `json:"rat_meeting_id,omitempty"`
		Reason          string     `json:"reason"`
		EffectiveDate   string     `json:"effective_date"`
		LiquidatorIDs   []string   `json:"liquidator_ids"`
		SupervisorID    *string    `json:"supervisor_id,omitempty"`
		ClaimDeadline   *string    `json:"claim_deadline,omitempty"`
		InitiatedBy     string     `json:"initiated_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	cmd, err := h.buildCreateCommand(req)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// UpdateStage godoc
// @Summary      Update dissolution stage
// @Tags         dissolution
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Dissolution ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/dissolution/{id}/stage [put]
func (h *DissolutionHandler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Stage string `json:"stage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), updatedissolutionstage.Command{
		ID:       id,
		NewStage: dissolution.Stage(req.Stage),
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Complete godoc
// @Summary      Complete dissolution process
// @Tags         dissolution
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Dissolution ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/dissolution/{id}/complete [put]
func (h *DissolutionHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		TotalAssets      int64   `json:"total_assets"`
		TotalLiabilities int64   `json:"total_liabilities"`
		MemberCount      int     `json:"member_count"`
		FinalReportDocID *string `json:"final_report_doc_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	cmd := completedissolution.Command{
		ID:               id,
		TotalAssets:      req.TotalAssets,
		TotalLiabilities: req.TotalLiabilities,
		MemberCount:      req.MemberCount,
	}
	if req.FinalReportDocID != nil {
		docID, parseErr := uuid.Parse(*req.FinalReportDocID)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "final_report_doc_id tidak valid")
			return
		}
		cmd.FinalReportDocID = &docID
	}
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Cancel godoc
// @Summary      Cancel dissolution process
// @Tags         dissolution
// @Param        id   path      string  true  "Dissolution ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/dissolution/{id}/cancel [put]
func (h *DissolutionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), canceldissolution.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete dissolution process
// @Tags         dissolution
// @Param        id   path      string  true  "Dissolution ID"
// @Success      204
// @Router       /api/v1/dissolution/{id} [delete]
func (h *DissolutionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletedissolution.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseListFilter membaca filter dari query parameters.
func (h *DissolutionHandler) parseListFilter(r *http.Request) dissolution.ListFilter {
	var filter dissolution.ListFilter
	q := r.URL.Query()

	if dt := q.Get("dissolution_type"); dt != "" {
		d := dissolution.DissolutionType(dt)
		filter.DissolutionType = &d
	}
	if st := q.Get("status"); st != "" {
		s := dissolution.Status(st)
		filter.Status = &s
	}
	if sg := q.Get("stage"); sg != "" {
		s := dissolution.Stage(sg)
		filter.Stage = &s
	}

	return filter
}

// buildCreateCommand mengubah HTTP request menjadi create_dissolution.Command.
func (h *DissolutionHandler) buildCreateCommand(req struct {
	DissolutionType string     `json:"dissolution_type"`
	RatMeetingID    *string    `json:"rat_meeting_id,omitempty"`
	Reason          string     `json:"reason"`
	EffectiveDate   string     `json:"effective_date"`
	LiquidatorIDs   []string   `json:"liquidator_ids"`
	SupervisorID    *string    `json:"supervisor_id,omitempty"`
	ClaimDeadline   *string    `json:"claim_deadline,omitempty"`
	InitiatedBy     string     `json:"initiated_by"`
}) (createdissolution.Command, error) {
	initiatedBy, err := uuid.Parse(req.InitiatedBy)
	if err != nil {
		return createdissolution.Command{}, err
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return createdissolution.Command{}, err
	}

	cmd := createdissolution.Command{
		DissolutionType: dissolution.DissolutionType(req.DissolutionType),
		Reason:          req.Reason,
		EffectiveDate:   effectiveDate,
		InitiatedBy:     initiatedBy,
	}

	if req.RatMeetingID != nil {
		id, parseErr := uuid.Parse(*req.RatMeetingID)
		if parseErr != nil {
			return createdissolution.Command{}, parseErr
		}
		cmd.RatMeetingID = &id
	}

	if req.SupervisorID != nil {
		id, parseErr := uuid.Parse(*req.SupervisorID)
		if parseErr != nil {
			return createdissolution.Command{}, parseErr
		}
		cmd.SupervisorID = &id
	}

	if req.ClaimDeadline != nil {
		d, parseErr := time.Parse("2006-01-02", *req.ClaimDeadline)
		if parseErr != nil {
			return createdissolution.Command{}, parseErr
		}
		cmd.ClaimDeadline = &d
	}

	if len(req.LiquidatorIDs) > 0 {
		cmd.LiquidatorIDs = make([]uuid.UUID, len(req.LiquidatorIDs))
		for i, idStr := range req.LiquidatorIDs {
			id, parseErr := uuid.Parse(idStr)
			if parseErr != nil {
				return createdissolution.Command{}, parseErr
			}
			cmd.LiquidatorIDs[i] = id
		}
	}

	return cmd, nil
}
