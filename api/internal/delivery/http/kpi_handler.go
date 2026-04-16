package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	createkpi "github.com/yourorg/boilerplate/internal/command/create_kpi"
	deletekpi "github.com/yourorg/boilerplate/internal/command/delete_kpi"
	recordkpimeasurement "github.com/yourorg/boilerplate/internal/command/record_kpi_measurement"
	createstresstest "github.com/yourorg/boilerplate/internal/command/create_stress_test"
	deletestresstest "github.com/yourorg/boilerplate/internal/command/delete_stress_test"
	updatestresstestresults "github.com/yourorg/boilerplate/internal/command/update_stress_test_results"
	updatekpi "github.com/yourorg/boilerplate/internal/command/update_kpi"
	getcomposithealthscore "github.com/yourorg/boilerplate/internal/query/get_composite_health_score"
	getkpi "github.com/yourorg/boilerplate/internal/query/get_kpi_by_id"
	listkpimeasurements "github.com/yourorg/boilerplate/internal/query/list_kpi_measurements"
	listkpis "github.com/yourorg/boilerplate/internal/query/list_kpis"
	liststresstests "github.com/yourorg/boilerplate/internal/query/list_stress_tests"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// KPIHandler menangani HTTP request untuk domain KPI.
type KPIHandler struct {
	BaseHandler
}

var kpiDefSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"name":       "name",
		"code":       "code",
		"category":   "category",
		"created_at": "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

var stressTestSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"name":       "name",
		"created_at": "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewKPIHandler membuat instance KPIHandler baru.
func NewKPIHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *KPIHandler {
	return &KPIHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain KPI.
func (h *KPIHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/kpis", func(r chi.Router) {
		// KPI Definitions
		r.Get("/", h.ListKPIs)
		r.Post("/", h.CreateKPI)
		r.Get("/health-score", h.GetHealthScore)
		r.Get("/{id}", h.GetKPIByID)
		r.Put("/{id}", h.UpdateKPI)
		r.Delete("/{id}", h.DeleteKPI)

		// KPI Measurements
		r.Post("/measurements", h.RecordMeasurement)
		r.Get("/measurements", h.ListMeasurements)

		// Stress Tests
		r.Post("/stress-tests", h.CreateStressTest)
		r.Get("/stress-tests", h.ListStressTests)
		r.Put("/stress-tests/{id}/results", h.UpdateStressTestResults)
		r.Delete("/stress-tests/{id}", h.DeleteStressTest)
	})
}

// ListKPIs godoc
// @Summary      List KPI definitions
// @Tags         kpis
// @Produce      json
// @Param        category   query  string  false  "Filter by category"
// @Param        frequency  query  string  false  "Filter by frequency"
// @Success      200  {object}  listkpis.Result
// @Router       /api/v1/kpis [get]
func (h *KPIHandler) ListKPIs(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, kpiDefSortConfig)
	filter := h.parseKPIFilter(r)
	result, err := querybus.Dispatch[*listkpis.Result](r.Context(), h.queryBus, listkpis.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetKPIByID godoc
// @Summary      Get KPI definition by ID
// @Tags         kpis
// @Produce      json
// @Param        id   path      string  true  "KPI ID"
// @Success      200  {object}  getkpi.Result
// @Router       /api/v1/kpis/{id} [get]
func (h *KPIHandler) GetKPIByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getkpi.Result](r.Context(), h.queryBus, getkpi.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateKPI godoc
// @Summary      Create KPI definition
// @Tags         kpis
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/kpis [post]
func (h *KPIHandler) CreateKPI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string  `json:"name"`
		Code              string  `json:"code"`
		Category          string  `json:"category"`
		Frequency         string  `json:"frequency"`
		Unit              string  `json:"unit"`
		TargetValue       float64 `json:"target_value"`
		WarningThreshold  float64 `json:"warning_threshold"`
		CriticalThreshold float64 `json:"critical_threshold"`
		Description       string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createkpi.Command{
		Name:              req.Name,
		Code:              req.Code,
		Category:          kpi.KPICategory(req.Category),
		Frequency:         kpi.KPIFrequency(req.Frequency),
		Unit:              req.Unit,
		TargetValue:       req.TargetValue,
		WarningThreshold:  req.WarningThreshold,
		CriticalThreshold: req.CriticalThreshold,
		Description:       req.Description,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// UpdateKPI godoc
// @Summary      Update KPI definition
// @Tags         kpis
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "KPI ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/kpis/{id} [put]
func (h *KPIHandler) UpdateKPI(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Name              string  `json:"name"`
		Code              string  `json:"code"`
		Category          string  `json:"category"`
		Frequency         string  `json:"frequency"`
		Unit              string  `json:"unit"`
		TargetValue       float64 `json:"target_value"`
		WarningThreshold  float64 `json:"warning_threshold"`
		CriticalThreshold float64 `json:"critical_threshold"`
		IsActive          bool    `json:"is_active"`
		Description       string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), updatekpi.Command{
		ID:                id,
		Name:              req.Name,
		Code:              req.Code,
		Category:          kpi.KPICategory(req.Category),
		Frequency:         kpi.KPIFrequency(req.Frequency),
		Unit:              req.Unit,
		TargetValue:       req.TargetValue,
		WarningThreshold:  req.WarningThreshold,
		CriticalThreshold: req.CriticalThreshold,
		IsActive:          req.IsActive,
		Description:       req.Description,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// DeleteKPI godoc
// @Summary      Delete KPI definition
// @Tags         kpis
// @Param        id   path      string  true  "KPI ID"
// @Success      204
// @Router       /api/v1/kpis/{id} [delete]
func (h *KPIHandler) DeleteKPI(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletekpi.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RecordMeasurement godoc
// @Summary      Record KPI measurement
// @Tags         kpis
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/kpis/measurements [post]
func (h *KPIHandler) RecordMeasurement(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DefinitionID string `json:"definition_id"`
		MeasuredValue float64 `json:"measured_value"`
		MeasuredBy   string `json:"measured_by"`
		Notes        string `json:"notes"`
		PeriodStart  string `json:"period_start"`
		PeriodEnd    string `json:"period_end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	defID, err := uuid.Parse(req.DefinitionID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "definition_id tidak valid")
		return
	}
	measuredBy, err := uuid.Parse(req.MeasuredBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "measured_by tidak valid")
		return
	}
	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "period_start tidak valid")
		return
	}
	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "period_end tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, recordkpimeasurement.Command{
		DefinitionID:  defID,
		MeasuredValue: req.MeasuredValue,
		MeasuredBy:    measuredBy,
		Notes:         req.Notes,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// ListMeasurements godoc
// @Summary      List KPI measurements
// @Tags         kpis
// @Produce      json
// @Param        definition_id  query  string  false  "Filter by definition ID"
// @Param        status         query  string  false  "Filter by status"
// @Success      200  {object}  listkpimeasurements.Result
// @Router       /api/v1/kpis/measurements [get]
func (h *KPIHandler) ListMeasurements(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, kpiDefSortConfig)
	filter := h.parseMeasurementFilter(r)
	result, err := querybus.Dispatch[*listkpimeasurements.Result](r.Context(), h.queryBus, listkpimeasurements.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetHealthScore godoc
// @Summary      Get composite health score
// @Tags         kpis
// @Produce      json
// @Success      200  {object}  getcomposithealthscore.Result
// @Router       /api/v1/kpis/health-score [get]
func (h *KPIHandler) GetHealthScore(w http.ResponseWriter, r *http.Request) {
	result, err := querybus.Dispatch[*getcomposithealthscore.Result](r.Context(), h.queryBus, getcomposithealthscore.Query{})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateStressTest godoc
// @Summary      Create stress test scenario
// @Tags         kpis
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/kpis/stress-tests [post]
func (h *KPIHandler) CreateStressTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  string `json:"parameters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createstresstest.Command{
		Name:        req.Name,
		Description: req.Description,
		Parameters:  req.Parameters,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// ListStressTests godoc
// @Summary      List stress test scenarios
// @Tags         kpis
// @Produce      json
// @Param        scenario_status  query  string  false  "Filter by status"
// @Success      200  {object}  liststresstests.Result
// @Router       /api/v1/kpis/stress-tests [get]
func (h *KPIHandler) ListStressTests(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, stressTestSortConfig)
	filter := h.parseStressTestFilter(r)
	result, err := querybus.Dispatch[*liststresstests.Result](r.Context(), h.queryBus, liststresstests.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// UpdateStressTestResults godoc
// @Summary      Update stress test results
// @Tags         kpis
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Stress Test ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/kpis/stress-tests/{id}/results [put]
func (h *KPIHandler) UpdateStressTestResults(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		ScenarioStatus string  `json:"scenario_status"`
		Results        string  `json:"results"`
		SimulatedBy    *string `json:"simulated_by,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	cmd := updatestresstestresults.Command{
		ID:             id,
		ScenarioStatus: kpi.ScenarioStatus(req.ScenarioStatus),
		Results:        req.Results,
	}
	if req.SimulatedBy != nil {
		sb, parseErr := uuid.Parse(*req.SimulatedBy)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "simulated_by tidak valid")
			return
		}
		cmd.SimulatedBy = &sb
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// DeleteStressTest godoc
// @Summary      Delete stress test scenario
// @Tags         kpis
// @Param        id   path      string  true  "Stress Test ID"
// @Success      204
// @Router       /api/v1/kpis/stress-tests/{id} [delete]
func (h *KPIHandler) DeleteStressTest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletestresstest.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseKPIFilter membaca filter KPI dari query parameters.
func (h *KPIHandler) parseKPIFilter(r *http.Request) kpi.KPIDefinitionFilter {
	var filter kpi.KPIDefinitionFilter
	q := r.URL.Query()

	if cat := q.Get("category"); cat != "" {
		c := kpi.KPICategory(cat)
		filter.Category = &c
	}
	if freq := q.Get("frequency"); freq != "" {
		f := kpi.KPIFrequency(freq)
		filter.Frequency = &f
	}
	if active := q.Get("is_active"); active != "" {
		val := active == "true"
		filter.IsActive = &val
	}

	return filter
}

// parseMeasurementFilter membaca filter measurement dari query parameters.
func (h *KPIHandler) parseMeasurementFilter(r *http.Request) kpi.KPIMeasurementFilter {
	var filter kpi.KPIMeasurementFilter
	q := r.URL.Query()

	if defID := q.Get("definition_id"); defID != "" {
		id, err := uuid.Parse(defID)
		if err == nil {
			filter.DefinitionID = &id
		}
	}
	if status := q.Get("status"); status != "" {
		s := kpi.MeasurementStatus(status)
		filter.Status = &s
	}

	return filter
}

// parseStressTestFilter membaca filter stress test dari query parameters.
func (h *KPIHandler) parseStressTestFilter(r *http.Request) kpi.StressTestFilter {
	var filter kpi.StressTestFilter
	q := r.URL.Query()

	if st := q.Get("scenario_status"); st != "" {
		s := kpi.ScenarioStatus(st)
		filter.ScenarioStatus = &s
	}

	return filter
}
