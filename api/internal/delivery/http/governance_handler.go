package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/governance"
	creategovernanceposition "github.com/yourorg/boilerplate/internal/command/create_governance_position"
	deletegovernanceposition "github.com/yourorg/boilerplate/internal/command/delete_governance_position"
	updategovernanceposition "github.com/yourorg/boilerplate/internal/command/update_governance_position"
	updategovernancestatus "github.com/yourorg/boilerplate/internal/command/update_governance_status"
	getgovernanceposition "github.com/yourorg/boilerplate/internal/query/get_governance_position_by_id"
	listgovernancepositions "github.com/yourorg/boilerplate/internal/query/list_governance_positions"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// GovernanceHandler menangani HTTP request untuk domain Governance.
type GovernanceHandler struct {
	BaseHandler
}

var governanceSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at":      "created_at",
		"position_type":   "position_type",
		"position_level":  "position_level",
		"status":          "status",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewGovernanceHandler membuat instance GovernanceHandler baru.
func NewGovernanceHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *GovernanceHandler {
	return &GovernanceHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Governance.
func (h *GovernanceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/governance-positions", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Put("/{id}/status", h.UpdateStatus)
		r.Delete("/{id}", h.Delete)
	})
}

// List godoc
// @Summary      List governance positions
// @Tags         governance
// @Produce      json
// @Param        position_type  query  string  false  "Filter by position type"
// @Param        status         query  string  false  "Filter by status"
// @Param        position_level query  string  false  "Filter by position level"
// @Success      200  {object}  listgovernancepositions.Result
// @Router       /api/v1/governance-positions [get]
func (h *GovernanceHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, governanceSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listgovernancepositions.Result](r.Context(), h.queryBus, listgovernancepositions.Query{
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
// @Summary      Get governance position by ID
// @Tags         governance
// @Produce      json
// @Param        id   path      string  true  "Governance Position ID"
// @Success      200  {object}  getgovernanceposition.Result
// @Router       /api/v1/governance-positions/{id} [get]
func (h *GovernanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getgovernanceposition.Result](r.Context(), h.queryBus, getgovernanceposition.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create governance position
// @Tags         governance
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/governance-positions [post]
func (h *GovernanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NasabahID         string     `json:"nasabah_id"`
		PositionType      string     `json:"position_type"`
		PositionLevel     string     `json:"position_level"`
		TermStart         string     `json:"term_start"`
		TermEnd           string     `json:"term_end"`
		TermNumber        int        `json:"term_number"`
		AppointedBy       string     `json:"appointed_by"`
		AppointmentDocID  *string    `json:"appointment_doc_id,omitempty"`
		MaxApprovalAmount int64      `json:"max_approval_amount"`
		CanDisburse       bool       `json:"can_disburse"`
		CanReverse        bool       `json:"can_reverse"`
		CanWaivePenalty   bool       `json:"can_waive_penalty"`
		CanWriteOff       bool       `json:"can_write_off"`
		CreatedBy         string     `json:"created_by"`
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

// Update godoc
// @Summary      Update governance position
// @Tags         governance
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Governance Position ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/governance-positions/{id} [put]
func (h *GovernanceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		PositionType      string  `json:"position_type"`
		PositionLevel     string  `json:"position_level"`
		TermStart         string  `json:"term_start"`
		TermEnd           string  `json:"term_end"`
		TermNumber        int     `json:"term_number"`
		AppointedBy       string  `json:"appointed_by"`
		AppointmentDocID  *string `json:"appointment_doc_id,omitempty"`
		MaxApprovalAmount int64   `json:"max_approval_amount"`
		CanDisburse       bool    `json:"can_disburse"`
		CanReverse        bool    `json:"can_reverse"`
		CanWaivePenalty   bool    `json:"can_waive_penalty"`
		CanWriteOff       bool    `json:"can_write_off"`
		UpdatedBy         string  `json:"updated_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	cmd, parseErr := h.buildUpdateCommand(id, req)
	if parseErr != nil {
		h.RespondError(w, http.StatusBadRequest, parseErr.Error())
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// UpdateStatus godoc
// @Summary      Update governance position status
// @Tags         governance
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Governance Position ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/governance-positions/{id}/status [put]
func (h *GovernanceHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Status    string `json:"status"`
		UpdatedBy string `json:"updated_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	updatedBy, parseErr := uuid.Parse(req.UpdatedBy)
	if parseErr != nil {
		h.RespondError(w, http.StatusBadRequest, "updated_by tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), updategovernancestatus.Command{
		ID: id, Status: req.Status, UpdatedBy: updatedBy,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete governance position
// @Tags         governance
// @Param        id   path      string  true  "Governance Position ID"
// @Success      204
// @Router       /api/v1/governance-positions/{id} [delete]
func (h *GovernanceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletegovernanceposition.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseListFilter membaca filter dari query parameters.
func (h *GovernanceHandler) parseListFilter(r *http.Request) governance.ListFilter {
	var filter governance.ListFilter
	q := r.URL.Query()

	if pt := q.Get("position_type"); pt != "" {
		filter.PositionType = &pt
	}
	if st := q.Get("status"); st != "" {
		filter.Status = &st
	}
	if pl := q.Get("position_level"); pl != "" {
		filter.PositionLevel = &pl
	}

	return filter
}

// buildCreateCommand mengubah HTTP request menjadi create_governance_position.Command.
func (h *GovernanceHandler) buildCreateCommand(req struct {
	NasabahID         string     `json:"nasabah_id"`
	PositionType      string     `json:"position_type"`
	PositionLevel     string     `json:"position_level"`
	TermStart         string     `json:"term_start"`
	TermEnd           string     `json:"term_end"`
	TermNumber        int        `json:"term_number"`
	AppointedBy       string     `json:"appointed_by"`
	AppointmentDocID  *string    `json:"appointment_doc_id,omitempty"`
	MaxApprovalAmount int64      `json:"max_approval_amount"`
	CanDisburse       bool       `json:"can_disburse"`
	CanReverse        bool       `json:"can_reverse"`
	CanWaivePenalty   bool       `json:"can_waive_penalty"`
	CanWriteOff       bool       `json:"can_write_off"`
	CreatedBy         string     `json:"created_by"`
}) (creategovernanceposition.Command, error) {
	nasabahID, err := uuid.Parse(req.NasabahID)
	if err != nil {
		return creategovernanceposition.Command{}, err
	}
	termStart, err := time.Parse("2006-01-02", req.TermStart)
	if err != nil {
		return creategovernanceposition.Command{}, err
	}
	termEnd, err := time.Parse("2006-01-02", req.TermEnd)
	if err != nil {
		return creategovernanceposition.Command{}, err
	}
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return creategovernanceposition.Command{}, err
	}

	cmd := creategovernanceposition.Command{
		NasabahID:         nasabahID,
		PositionType:      req.PositionType,
		PositionLevel:     req.PositionLevel,
		TermStart:         termStart,
		TermEnd:           termEnd,
		TermNumber:        req.TermNumber,
		AppointedBy:       req.AppointedBy,
		MaxApprovalAmount: req.MaxApprovalAmount,
		CanDisburse:       req.CanDisburse,
		CanReverse:        req.CanReverse,
		CanWaivePenalty:   req.CanWaivePenalty,
		CanWriteOff:       req.CanWriteOff,
		CreatedBy:         createdBy,
	}

	if req.AppointmentDocID != nil {
		id, parseErr := uuid.Parse(*req.AppointmentDocID)
		if parseErr != nil {
			return creategovernanceposition.Command{}, parseErr
		}
		cmd.AppointmentDocID = &id
	}

	return cmd, nil
}

// buildUpdateCommand mengubah HTTP request menjadi update_governance_position.Command.
func (h *GovernanceHandler) buildUpdateCommand(id uuid.UUID, req struct {
	PositionType      string  `json:"position_type"`
	PositionLevel     string  `json:"position_level"`
	TermStart         string  `json:"term_start"`
	TermEnd           string  `json:"term_end"`
	TermNumber        int     `json:"term_number"`
	AppointedBy       string  `json:"appointed_by"`
	AppointmentDocID  *string `json:"appointment_doc_id,omitempty"`
	MaxApprovalAmount int64   `json:"max_approval_amount"`
	CanDisburse       bool    `json:"can_disburse"`
	CanReverse        bool    `json:"can_reverse"`
	CanWaivePenalty   bool    `json:"can_waive_penalty"`
	CanWriteOff       bool    `json:"can_write_off"`
	UpdatedBy         string  `json:"updated_by"`
}) (updategovernanceposition.Command, error) {
	updatedBy, err := uuid.Parse(req.UpdatedBy)
	if err != nil {
		return updategovernanceposition.Command{}, err
	}

	cmd := updategovernanceposition.Command{
		ID:                id,
		PositionType:      req.PositionType,
		PositionLevel:     req.PositionLevel,
		TermNumber:        req.TermNumber,
		AppointedBy:       req.AppointedBy,
		MaxApprovalAmount: req.MaxApprovalAmount,
		CanDisburse:       req.CanDisburse,
		CanReverse:        req.CanReverse,
		CanWaivePenalty:   req.CanWaivePenalty,
		CanWriteOff:       req.CanWriteOff,
		UpdatedBy:         updatedBy,
	}

	if req.TermStart != "" {
		ts, parseErr := time.Parse("2006-01-02", req.TermStart)
		if parseErr != nil {
			return updategovernanceposition.Command{}, parseErr
		}
		cmd.TermStart = ts
	}
	if req.TermEnd != "" {
		te, parseErr := time.Parse("2006-01-02", req.TermEnd)
		if parseErr != nil {
			return updategovernanceposition.Command{}, parseErr
		}
		cmd.TermEnd = te
	}
	if req.AppointmentDocID != nil {
		aid, parseErr := uuid.Parse(*req.AppointmentDocID)
		if parseErr != nil {
			return updategovernanceposition.Command{}, parseErr
		}
		cmd.AppointmentDocID = &aid
	}

	return cmd, nil
}
