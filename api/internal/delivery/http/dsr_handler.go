package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dsr"
	completedsr "github.com/yourorg/boilerplate/internal/command/complete_dsr"
	createdsr "github.com/yourorg/boilerplate/internal/command/create_dsr"
	updatedsrstatus "github.com/yourorg/boilerplate/internal/command/update_dsr_status"
	getdsr "github.com/yourorg/boilerplate/internal/query/get_dsr_by_id"
	listdsrs "github.com/yourorg/boilerplate/internal/query/list_dsrs"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// DSRHandler menangani HTTP request untuk domain Data Subject Request.
type DSRHandler struct {
	BaseHandler
}

var dsrSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at": "created_at",
		"due_date":   "due_date",
		"status":     "status",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewDSRHandler membuat instance DSRHandler baru.
func NewDSRHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *DSRHandler {
	return &DSRHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain DSR.
func (h *DSRHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/data-subject-requests", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}/status", h.UpdateStatus)
		r.Put("/{id}/complete", h.Complete)
	})
}

// List godoc
// @Summary      List data subject requests
// @Tags         data-subject-requests
// @Produce      json
// @Param        request_type  query  string  false  "Filter by request type"
// @Param        status        query  string  false  "Filter by status"
// @Success      200  {object}  listdsrs.Result
// @Router       /api/v1/data-subject-requests [get]
func (h *DSRHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, dsrSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listdsrs.Result](r.Context(), h.queryBus, listdsrs.Query{
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
// @Summary      Get data subject request by ID
// @Tags         data-subject-requests
// @Produce      json
// @Param        id   path      string  true  "DSR ID"
// @Success      200  {object}  getdsr.Result
// @Router       /api/v1/data-subject-requests/{id} [get]
func (h *DSRHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getdsr.Result](r.Context(), h.queryBus, getdsr.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create data subject request
// @Tags         data-subject-requests
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/data-subject-requests [post]
func (h *DSRHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RequestType    string `json:"request_type"`
		RequestorName  string `json:"requestor_name"`
		RequestorEmail string `json:"requestor_email"`
		SubjectID      string `json:"subject_id"`
		SubjectType    string `json:"subject_type"`
		Description    string `json:"description"`
		DueDate        string `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	subjectID, err := uuid.Parse(req.SubjectID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "subject_id tidak valid")
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "due_date tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createdsr.Command{
		RequestType:    dsr.RequestType(req.RequestType),
		RequestorName:  req.RequestorName,
		RequestorEmail: req.RequestorEmail,
		SubjectID:      subjectID,
		SubjectType:    req.SubjectType,
		Description:    req.Description,
		DueDate:        dueDate,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// UpdateStatus godoc
// @Summary      Update data subject request status
// @Tags         data-subject-requests
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "DSR ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/data-subject-requests/{id}/status [put]
func (h *DSRHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Status          string  `json:"status"`
		VerifiedBy      *string `json:"verified_by,omitempty"`
		RejectionReason string  `json:"rejection_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	cmd := updatedsrstatus.Command{
		ID:              id,
		Status:          dsr.RequestStatus(req.Status),
		RejectionReason: req.RejectionReason,
	}
	if req.VerifiedBy != nil {
		vb, parseErr := uuid.Parse(*req.VerifiedBy)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "verified_by tidak valid")
			return
		}
		cmd.VerifiedBy = &vb
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Complete godoc
// @Summary      Complete data subject request
// @Tags         data-subject-requests
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "DSR ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/data-subject-requests/{id}/complete [put]
func (h *DSRHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		CompletedBy  string `json:"completed_by"`
		ResponseData string `json:"response_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	completedBy, err := uuid.Parse(req.CompletedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "completed_by tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), completedsr.Command{
		ID:           id,
		CompletedBy:  completedBy,
		ResponseData: req.ResponseData,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// parseListFilter membaca filter dari query parameters.
func (h *DSRHandler) parseListFilter(r *http.Request) dsr.DSRFilter {
	var filter dsr.DSRFilter
	q := r.URL.Query()

	if rt := q.Get("request_type"); rt != "" {
		t := dsr.RequestType(rt)
		filter.RequestType = &t
	}
	if st := q.Get("status"); st != "" {
		s := dsr.RequestStatus(st)
		filter.Status = &s
	}
	if sid := q.Get("subject_id"); sid != "" {
		id, err := uuid.Parse(sid)
		if err == nil {
			filter.SubjectID = &id
		}
	}

	return filter
}
