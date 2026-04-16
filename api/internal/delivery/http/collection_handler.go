package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	createcase "github.com/yourorg/boilerplate/internal/command/create_collection_case"
	updatebucket "github.com/yourorg/boilerplate/internal/command/update_aging_bucket"
	assigncollector "github.com/yourorg/boilerplate/internal/command/assign_collector"
	resolvecase "github.com/yourorg/boilerplate/internal/command/resolve_collection_case"
	createactivity "github.com/yourorg/boilerplate/internal/command/create_collection_activity"
	recordperf "github.com/yourorg/boilerplate/internal/command/record_collector_performance"
	getcase "github.com/yourorg/boilerplate/internal/query/get_collection_case_by_id"
	listcases "github.com/yourorg/boilerplate/internal/query/list_collection_cases"
	listactivities "github.com/yourorg/boilerplate/internal/query/list_collection_activities"
	listperformances "github.com/yourorg/boilerplate/internal/query/list_collector_performances"
	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// CollectionHandler menangani HTTP request untuk domain Collection.
type CollectionHandler struct {
	BaseHandler
}

var collectionCaseSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"case_number":  "case_number",
		"current_dpd":  "current_dpd",
		"aging_bucket": "aging_bucket",
		"status":       "status",
		"opened_at":    "opened_at",
		"created_at":   "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewCollectionHandler membuat instance CollectionHandler baru.
func NewCollectionHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *CollectionHandler {
	return &CollectionHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Collection.
func (h *CollectionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/collection", func(r chi.Router) {
		// Collection Cases
		r.Route("/cases", func(r chi.Router) {
			r.Get("/", h.ListCases)
			r.Post("/", h.CreateCase)
			r.Get("/{id}", h.GetCaseByID)
			r.Put("/{id}/aging", h.UpdateAgingBucket)
			r.Put("/{id}/assign", h.AssignCollector)
			r.Put("/{id}/resolve", h.ResolveCase)
		})

		// Collection Activities
		r.Route("/cases/{caseId}/activities", func(r chi.Router) {
			r.Get("/", h.ListActivities)
			r.Post("/", h.CreateActivity)
		})

		// Collector Performances
		r.Route("/performances", func(r chi.Router) {
			r.Get("/", h.ListPerformances)
			r.Post("/", h.RecordPerformance)
		})
	})
}

// ── Collection Cases ──────────────────────────────────────────────────────────

// ListCases godoc
// @Summary      List collection cases
// @Tags         collection
// @Produce      json
// @Param        status        query    string  false  "Filter by status"
// @Param        aging_bucket  query    string  false  "Filter by aging bucket"
// @Param        collector_id  query    string  false  "Filter by collector ID"
// @Param        nasabah_id    query    string  false  "Filter by nasabah ID"
// @Success      200  {object}  listcases.Result
// @Router       /api/v1/collection/cases [get]
func (h *CollectionHandler) ListCases(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, collectionCaseSortConfig)

	qry := listcases.Query{Params: params}

	if s := r.URL.Query().Get("status"); s != "" {
		status := collection.CaseStatus(s)
		qry.Status = &status
	}
	if s := r.URL.Query().Get("aging_bucket"); s != "" {
		bucket := collection.AgingBucket(s)
		qry.AgingBucket = &bucket
	}
	if s := r.URL.Query().Get("collector_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			qry.CollectorID = &id
		}
	}
	if s := r.URL.Query().Get("nasabah_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			qry.NasabahID = &id
		}
	}

	result, err := querybus.Dispatch[*listcases.Result](r.Context(), h.queryBus, qry)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data collection cases")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetCaseByID godoc
// @Summary      Get collection case by ID
// @Tags         collection
// @Produce      json
// @Param        id   path      string  true  "Case ID"
// @Success      200  {object}  getcase.Result
// @Router       /api/v1/collection/cases/{id} [get]
func (h *CollectionHandler) GetCaseByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getcase.Result](r.Context(), h.queryBus, getcase.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateCase godoc
// @Summary      Create collection case
// @Tags         collection
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/collection/cases [post]
func (h *CollectionHandler) CreateCase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PinjamanID               string `json:"pinjaman_id"`
		NasabahID                string `json:"nasabah_id"`
		CaseNumber               string `json:"case_number"`
		CurrentDPD               int    `json:"current_dpd"`
		TotalOverdueAmount       int64  `json:"total_overdue_amount"`
		TotalOverdueInstallments int    `json:"total_overdue_installments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	pinjamanID, err := uuid.Parse(req.PinjamanID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "pinjaman_id tidak valid")
		return
	}
	nasabahID, err := uuid.Parse(req.NasabahID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "nasabah_id tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createcase.Command{
		PinjamanID:               pinjamanID,
		NasabahID:                nasabahID,
		CaseNumber:               req.CaseNumber,
		CurrentDPD:               req.CurrentDPD,
		TotalOverdueAmount:       req.TotalOverdueAmount,
		TotalOverdueInstallments: req.TotalOverdueInstallments,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// UpdateAgingBucket godoc
// @Summary      Update aging bucket
// @Tags         collection
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Case ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/collection/cases/{id}/aging [put]
func (h *CollectionHandler) UpdateAgingBucket(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		NewDPD             int   `json:"current_dpd"`
		TotalOverdueAmount int64 `json:"total_overdue_amount"`
		OverdueInstallments int  `json:"total_overdue_installments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), updatebucket.Command{
		ID:                  id,
		NewDPD:              req.NewDPD,
		TotalOverdueAmount:  req.TotalOverdueAmount,
		OverdueInstallments: req.OverdueInstallments,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// AssignCollector godoc
// @Summary      Assign collector to case
// @Tags         collection
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Case ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/collection/cases/{id}/assign [put]
func (h *CollectionHandler) AssignCollector(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		CollectorID string `json:"collector_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	collectorID, err := uuid.Parse(req.CollectorID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "collector_id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), assigncollector.Command{
		ID:          id,
		CollectorID: collectorID,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ResolveCase godoc
// @Summary      Resolve collection case
// @Tags         collection
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Case ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/collection/cases/{id}/resolve [put]
func (h *CollectionHandler) ResolveCase(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		ResolutionType string `json:"resolution_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), resolvecase.Command{
		ID:             id,
		ResolutionType: collection.ResolutionType(req.ResolutionType),
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ── Collection Activities ────────────────────────────────────────────────────

// ListActivities godoc
// @Summary      List collection activities
// @Tags         collection
// @Produce      json
// @Param        caseId         path     string  true  "Case ID"
// @Param        activity_type  query    string  false  "Filter by activity type"
// @Success      200  {object}  listactivities.Result
// @Router       /api/v1/collection/cases/{caseId}/activities [get]
func (h *CollectionHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	caseID, err := uuid.Parse(chi.URLParam(r, "caseId"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "case_id tidak valid")
		return
	}

	qry := listactivities.Query{CaseID: caseID}
	if s := r.URL.Query().Get("activity_type"); s != "" {
		at := collection.ActivityType(s)
		qry.ActivityType = &at
	}
	qry.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	qry.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	result, err := querybus.Dispatch[*listactivities.Result](r.Context(), h.queryBus, qry)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data aktivitas")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateActivity godoc
// @Summary      Create collection activity
// @Tags         collection
// @Accept       json
// @Produce      json
// @Param        caseId  path      string  true  "Case ID"
// @Success      201  {object}  map[string]string
// @Router       /api/v1/collection/cases/{caseId}/activities [post]
func (h *CollectionHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	caseID, err := uuid.Parse(chi.URLParam(r, "caseId"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "case_id tidak valid")
		return
	}
	var req struct {
		ActivityType     string  `json:"activity_type"`
		ActivityDate     string  `json:"activity_date"`
		PerformedBy      string  `json:"performed_by"`
		ContactResult    string  `json:"contact_result"`
		Notes            string  `json:"notes"`
		NasabahResponse  *string `json:"nasabah_response"`
		PromiseAmount    *int64  `json:"promise_amount"`
		PromiseDate      *string `json:"promise_date"`
		FollowupRequired bool    `json:"followup_required"`
		FollowupDate     *string `json:"followup_date"`
		FollowupType     *string `json:"followup_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	performedBy, err := uuid.Parse(req.PerformedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "performed_by tidak valid")
		return
	}

	activityDate, err := time.Parse("2006-01-02", req.ActivityDate)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "activity_date tidak valid (format: YYYY-MM-DD)")
		return
	}

	var nasabahResp *collection.NasabahResponse
	if req.NasabahResponse != nil {
		nr := collection.NasabahResponse(*req.NasabahResponse)
		nasabahResp = &nr
	}

	var promiseDate *time.Time
	if req.PromiseDate != nil {
		pd, err := time.Parse("2006-01-02", *req.PromiseDate)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "promise_date tidak valid (format: YYYY-MM-DD)")
			return
		}
		promiseDate = &pd
	}

	var followupDate *time.Time
	if req.FollowupDate != nil {
		fd, err := time.Parse("2006-01-02", *req.FollowupDate)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "followup_date tidak valid (format: YYYY-MM-DD)")
			return
		}
		followupDate = &fd
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createactivity.Command{
		CaseID:           caseID,
		ActivityType:     collection.ActivityType(req.ActivityType),
		ActivityDate:     activityDate,
		PerformedBy:      performedBy,
		ContactResult:    collection.ContactResult(req.ContactResult),
		Notes:            req.Notes,
		NasabahResponse:  nasabahResp,
		PromiseAmount:    req.PromiseAmount,
		PromiseDate:      promiseDate,
		FollowupRequired: req.FollowupRequired,
		FollowupDate:     followupDate,
		FollowupType:     req.FollowupType,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// ── Collector Performances ───────────────────────────────────────────────────

// ListPerformances godoc
// @Summary      List collector performances
// @Tags         collection
// @Produce      json
// @Param        collector_id   query    string  false  "Filter by collector ID"
// @Param        period_month   query    string  false  "Filter by period month (YYYY-MM)"
// @Success      200  {object}  listperformances.Result
// @Router       /api/v1/collection/performances [get]
func (h *CollectionHandler) ListPerformances(w http.ResponseWriter, r *http.Request) {
	qry := listperformances.Query{}

	if s := r.URL.Query().Get("collector_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			qry.CollectorID = &id
		}
	}
	qry.PeriodMonth = r.URL.Query().Get("period_month")
	qry.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	qry.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	result, err := querybus.Dispatch[*listperformances.Result](r.Context(), h.queryBus, qry)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data performa")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// RecordPerformance godoc
// @Summary      Record collector performance
// @Tags         collection
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/collection/performances [post]
func (h *CollectionHandler) RecordPerformance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CollectorID          string  `json:"collector_id"`
		PeriodMonth          string  `json:"period_month"`
		TotalCalls           int     `json:"total_calls"`
		SuccessfulContacts   int     `json:"successful_contacts"`
		TotalVisits          int     `json:"total_visits"`
		SuccessfulVisits     int     `json:"successful_visits"`
		CasesHandled         int     `json:"cases_handled"`
		CasesResolved        int     `json:"cases_resolved"`
		TotalAmountCollected int64   `json:"total_amount_collected"`
		PromiseToPayCount    int     `json:"promise_to_pay_count"`
		PromiseKeptCount     int     `json:"promise_kept_count"`
		ContactRatePct       float64 `json:"contact_rate_pct"`
		ResolutionRatePct    float64 `json:"resolution_rate_pct"`
		CollectionRatePct    float64 `json:"collection_rate_pct"`
		PromiseKeptRatePct   float64 `json:"promise_kept_rate_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	collectorID, err := uuid.Parse(req.CollectorID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "collector_id tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, recordperf.Command{
		CollectorID:          collectorID,
		PeriodMonth:          req.PeriodMonth,
		TotalCalls:           req.TotalCalls,
		SuccessfulContacts:   req.SuccessfulContacts,
		TotalVisits:          req.TotalVisits,
		SuccessfulVisits:     req.SuccessfulVisits,
		CasesHandled:         req.CasesHandled,
		CasesResolved:        req.CasesResolved,
		TotalAmountCollected: req.TotalAmountCollected,
		PromiseToPayCount:    req.PromiseToPayCount,
		PromiseKeptCount:     req.PromiseKeptCount,
		ContactRatePct:       req.ContactRatePct,
		ResolutionRatePct:    req.ResolutionRatePct,
		CollectionRatePct:    req.CollectionRatePct,
		PromiseKeptRatePct:   req.PromiseKeptRatePct,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}
