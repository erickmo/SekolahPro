package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/rat"
	cancelratmeeting "github.com/yourorg/boilerplate/internal/command/cancel_rat_meeting"
	createratmeeting "github.com/yourorg/boilerplate/internal/command/create_rat_meeting"
	deleteratmeeting "github.com/yourorg/boilerplate/internal/command/delete_rat_meeting"
	updateratmeeting "github.com/yourorg/boilerplate/internal/command/update_rat_meeting"
	getratmeeting "github.com/yourorg/boilerplate/internal/query/get_rat_meeting_by_id"
	listratmeetings "github.com/yourorg/boilerplate/internal/query/list_rat_meetings"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// RATHandler menangani HTTP request untuk domain RAT.
type RATHandler struct {
	BaseHandler
}

var ratSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at":    "created_at",
		"meeting_date":  "meeting_date",
		"status":        "status",
		"meeting_type":  "meeting_type",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewRATHandler membuat instance RATHandler baru.
func NewRATHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *RATHandler {
	return &RATHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain RAT.
func (h *RATHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/rat-meetings", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Put("/{id}/cancel", h.Cancel)
		r.Delete("/{id}", h.Delete)
	})
}

// List godoc
// @Summary      List RAT meetings
// @Tags         rat
// @Produce      json
// @Param        meeting_type  query  string  false  "Filter by meeting type"
// @Param        status        query  string  false  "Filter by status"
// @Success      200  {object}  listratmeetings.Result
// @Router       /api/v1/rat-meetings [get]
func (h *RATHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, ratSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listratmeetings.Result](r.Context(), h.queryBus, listratmeetings.Query{
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
// @Summary      Get RAT meeting by ID
// @Tags         rat
// @Produce      json
// @Param        id   path      string  true  "RAT Meeting ID"
// @Success      200  {object}  getratmeeting.Result
// @Router       /api/v1/rat-meetings/{id} [get]
func (h *RATHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getratmeeting.Result](r.Context(), h.queryBus, getratmeeting.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create RAT meeting
// @Tags         rat
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/rat-meetings [post]
func (h *RATHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MeetingType          string  `json:"meeting_type"`
		MeetingNumber        string  `json:"meeting_number"`
		Title                string  `json:"title"`
		Description          string  `json:"description"`
		MeetingDate          string  `json:"meeting_date"`
		MeetingTime          string  `json:"meeting_time"`
		MeetingLocation      string  `json:"meeting_location"`
		MeetingMode          string  `json:"meeting_mode"`
		OnlineLink           *string `json:"online_link,omitempty"`
		TotalEligibleMembers int     `json:"total_eligible_members"`
		QuorumRequired       int     `json:"quorum_required"`
		AgendaDocID          *string `json:"agenda_doc_id,omitempty"`
		MinutesDocID         *string `json:"minutes_doc_id,omitempty"`
		FinancialReportID    *string `json:"financial_report_id,omitempty"`
		ShuProposalID        *string `json:"shu_proposal_id,omitempty"`
		ConvenedBy           string  `json:"convened_by"`
		SecretaryID          string  `json:"secretary_id"`
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
// @Summary      Update RAT meeting
// @Tags         rat
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "RAT Meeting ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/rat-meetings/{id} [put]
func (h *RATHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		MeetingType          string  `json:"meeting_type"`
		MeetingNumber        string  `json:"meeting_number"`
		Title                string  `json:"title"`
		Description          string  `json:"description"`
		MeetingDate          string  `json:"meeting_date"`
		MeetingTime          string  `json:"meeting_time"`
		MeetingLocation      string  `json:"meeting_location"`
		MeetingMode          string  `json:"meeting_mode"`
		OnlineLink           *string `json:"online_link,omitempty"`
		TotalEligibleMembers int     `json:"total_eligible_members"`
		QuorumRequired       int     `json:"quorum_required"`
		ActualAttendees      int     `json:"actual_attendees"`
		QuorumMet            bool    `json:"quorum_met"`
		AgendaDocID          *string `json:"agenda_doc_id,omitempty"`
		MinutesDocID         *string `json:"minutes_doc_id,omitempty"`
		FinancialReportID    *string `json:"financial_report_id,omitempty"`
		ShuProposalID        *string `json:"shu_proposal_id,omitempty"`
		SecretaryID          string  `json:"secretary_id"`
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

// Cancel godoc
// @Summary      Cancel RAT meeting
// @Tags         rat
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "RAT Meeting ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/rat-meetings/{id}/cancel [put]
func (h *RATHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		CancelledReason *string `json:"cancelled_reason,omitempty"`
		RescheduledToID *string `json:"rescheduled_to_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	cmd := cancelratmeeting.Command{ID: id, CancelledReason: req.CancelledReason}
	if req.RescheduledToID != nil {
		rid, parseErr := uuid.Parse(*req.RescheduledToID)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "rescheduled_to_id tidak valid")
			return
		}
		cmd.RescheduledToID = &rid
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Delete godoc
// @Summary      Delete RAT meeting
// @Tags         rat
// @Param        id   path      string  true  "RAT Meeting ID"
// @Success      204
// @Router       /api/v1/rat-meetings/{id} [delete]
func (h *RATHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deleteratmeeting.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseListFilter membaca filter dari query parameters.
func (h *RATHandler) parseListFilter(r *http.Request) rat.ListFilter {
	var filter rat.ListFilter
	q := r.URL.Query()

	if mt := q.Get("meeting_type"); mt != "" {
		filter.MeetingType = &mt
	}
	if st := q.Get("status"); st != "" {
		filter.Status = &st
	}

	return filter
}

// buildCreateCommand mengubah HTTP request menjadi create_rat_meeting.Command.
func (h *RATHandler) buildCreateCommand(req struct {
	MeetingType          string  `json:"meeting_type"`
	MeetingNumber        string  `json:"meeting_number"`
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	MeetingDate          string  `json:"meeting_date"`
	MeetingTime          string  `json:"meeting_time"`
	MeetingLocation      string  `json:"meeting_location"`
	MeetingMode          string  `json:"meeting_mode"`
	OnlineLink           *string `json:"online_link,omitempty"`
	TotalEligibleMembers int     `json:"total_eligible_members"`
	QuorumRequired       int     `json:"quorum_required"`
	AgendaDocID          *string `json:"agenda_doc_id,omitempty"`
	MinutesDocID         *string `json:"minutes_doc_id,omitempty"`
	FinancialReportID    *string `json:"financial_report_id,omitempty"`
	ShuProposalID        *string `json:"shu_proposal_id,omitempty"`
	ConvenedBy           string  `json:"convened_by"`
	SecretaryID          string  `json:"secretary_id"`
}) (createratmeeting.Command, error) {
	convenedBy, err := uuid.Parse(req.ConvenedBy)
	if err != nil {
		return createratmeeting.Command{}, err
	}
	secretaryID, err := uuid.Parse(req.SecretaryID)
	if err != nil {
		return createratmeeting.Command{}, err
	}
	meetingDate, err := time.Parse("2006-01-02", req.MeetingDate)
	if err != nil {
		return createratmeeting.Command{}, err
	}

	cmd := createratmeeting.Command{
		MeetingType:          req.MeetingType,
		MeetingNumber:        req.MeetingNumber,
		Title:                req.Title,
		Description:          req.Description,
		MeetingDate:          meetingDate,
		MeetingTime:          req.MeetingTime,
		MeetingLocation:      req.MeetingLocation,
		MeetingMode:          req.MeetingMode,
		OnlineLink:           req.OnlineLink,
		TotalEligibleMembers: req.TotalEligibleMembers,
		QuorumRequired:       req.QuorumRequired,
		ConvenedBy:           convenedBy,
		SecretaryID:          secretaryID,
	}

	if req.AgendaDocID != nil {
		id, parseErr := uuid.Parse(*req.AgendaDocID)
		if parseErr != nil {
			return createratmeeting.Command{}, parseErr
		}
		cmd.AgendaDocID = &id
	}
	if req.MinutesDocID != nil {
		id, parseErr := uuid.Parse(*req.MinutesDocID)
		if parseErr != nil {
			return createratmeeting.Command{}, parseErr
		}
		cmd.MinutesDocID = &id
	}
	if req.FinancialReportID != nil {
		id, parseErr := uuid.Parse(*req.FinancialReportID)
		if parseErr != nil {
			return createratmeeting.Command{}, parseErr
		}
		cmd.FinancialReportID = &id
	}
	if req.ShuProposalID != nil {
		id, parseErr := uuid.Parse(*req.ShuProposalID)
		if parseErr != nil {
			return createratmeeting.Command{}, parseErr
		}
		cmd.ShuProposalID = &id
	}

	return cmd, nil
}

// buildUpdateCommand mengubah HTTP request menjadi update_rat_meeting.Command.
func (h *RATHandler) buildUpdateCommand(id uuid.UUID, req struct {
	MeetingType          string  `json:"meeting_type"`
	MeetingNumber        string  `json:"meeting_number"`
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	MeetingDate          string  `json:"meeting_date"`
	MeetingTime          string  `json:"meeting_time"`
	MeetingLocation      string  `json:"meeting_location"`
	MeetingMode          string  `json:"meeting_mode"`
	OnlineLink           *string `json:"online_link,omitempty"`
	TotalEligibleMembers int     `json:"total_eligible_members"`
	QuorumRequired       int     `json:"quorum_required"`
	ActualAttendees      int     `json:"actual_attendees"`
	QuorumMet            bool    `json:"quorum_met"`
	AgendaDocID          *string `json:"agenda_doc_id,omitempty"`
	MinutesDocID         *string `json:"minutes_doc_id,omitempty"`
	FinancialReportID    *string `json:"financial_report_id,omitempty"`
	ShuProposalID        *string `json:"shu_proposal_id,omitempty"`
	SecretaryID          string  `json:"secretary_id"`
}) (updateratmeeting.Command, error) {
	cmd := updateratmeeting.Command{
		ID:                  id,
		MeetingType:         req.MeetingType,
		MeetingNumber:       req.MeetingNumber,
		Title:               req.Title,
		Description:         req.Description,
		MeetingTime:         req.MeetingTime,
		MeetingLocation:     req.MeetingLocation,
		MeetingMode:         req.MeetingMode,
		OnlineLink:          req.OnlineLink,
		TotalEligibleMembers: req.TotalEligibleMembers,
		QuorumRequired:      req.QuorumRequired,
		ActualAttendees:     req.ActualAttendees,
		QuorumMet:           req.QuorumMet,
	}

	if req.MeetingDate != "" {
		md, parseErr := time.Parse("2006-01-02", req.MeetingDate)
		if parseErr != nil {
			return updateratmeeting.Command{}, parseErr
		}
		cmd.MeetingDate = md
	}
	if req.SecretaryID != "" {
		sid, parseErr := uuid.Parse(req.SecretaryID)
		if parseErr != nil {
			return updateratmeeting.Command{}, parseErr
		}
		cmd.SecretaryID = sid
	}
	if req.AgendaDocID != nil {
		aid, parseErr := uuid.Parse(*req.AgendaDocID)
		if parseErr != nil {
			return updateratmeeting.Command{}, parseErr
		}
		cmd.AgendaDocID = &aid
	}
	if req.MinutesDocID != nil {
		mid, parseErr := uuid.Parse(*req.MinutesDocID)
		if parseErr != nil {
			return updateratmeeting.Command{}, parseErr
		}
		cmd.MinutesDocID = &mid
	}
	if req.FinancialReportID != nil {
		fid, parseErr := uuid.Parse(*req.FinancialReportID)
		if parseErr != nil {
			return updateratmeeting.Command{}, parseErr
		}
		cmd.FinancialReportID = &fid
	}
	if req.ShuProposalID != nil {
		sid, parseErr := uuid.Parse(*req.ShuProposalID)
		if parseErr != nil {
			return updateratmeeting.Command{}, parseErr
		}
		cmd.ShuProposalID = &sid
	}

	return cmd, nil
}
