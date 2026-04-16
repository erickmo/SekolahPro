package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_alert"
	createamlalert "github.com/yourorg/boilerplate/internal/command/create_aml_alert"
	investigateamlalert "github.com/yourorg/boilerplate/internal/command/investigate_aml_alert"
	resolveamlalert "github.com/yourorg/boilerplate/internal/command/resolve_aml_alert"
	getamlalert "github.com/yourorg/boilerplate/internal/query/get_aml_alert_by_id"
	listamlalerts "github.com/yourorg/boilerplate/internal/query/list_aml_alerts"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// AMLAlertHandler menangani HTTP request untuk domain AML Alert.
type AMLAlertHandler struct {
	BaseHandler
}

var amlAlertSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at":  "created_at",
		"detected_at": "detected_at",
		"severity":    "severity",
		"status":      "status",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewAMLAlertHandler membuat instance AMLAlertHandler baru.
func NewAMLAlertHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *AMLAlertHandler {
	return &AMLAlertHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain AML Alert.
func (h *AMLAlertHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/aml/alerts", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}/investigate", h.Investigate)
		r.Put("/{id}/resolve", h.Resolve)
	})
}

// List godoc
// @Summary      List AML transaction alerts
// @Tags         aml-alert
// @Produce      json
// @Param        status      query  string  false  "Filter by status"
// @Param        severity    query  string  false  "Filter by severity"
// @Param        alert_type  query  string  false  "Filter by alert type"
// @Param        rule_type   query  string  false  "Filter by rule type"
// @Success      200  {object}  listamlalerts.Result
// @Router       /api/v1/aml/alerts [get]
func (h *AMLAlertHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, amlAlertSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listamlalerts.Result](r.Context(), h.queryBus, listamlalerts.Query{
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
// @Summary      Get AML alert by ID
// @Tags         aml-alert
// @Produce      json
// @Param        id   path      string  true  "Alert ID"
// @Success      200  {object}  getamlalert.Result
// @Router       /api/v1/aml/alerts/{id} [get]
func (h *AMLAlertHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getamlalert.Result](r.Context(), h.queryBus, getamlalert.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create AML transaction alert
// @Tags         aml-alert
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/aml/alerts [post]
func (h *AMLAlertHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NasabahID          string          `json:"nasabah_id"`
		TransaksiID        *string         `json:"transaksi_id,omitempty"`
		RuleID             string          `json:"rule_id"`
		RuleType           string          `json:"rule_type"`
		AlertType          string          `json:"alert_type"`
		Severity           string          `json:"severity"`
		Description        string          `json:"description"`
		TransactionDetails json.RawMessage `json:"transaction_details"`
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
	ruleID, err := uuid.Parse(req.RuleID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "rule_id tidak valid")
		return
	}

	cmd := createamlalert.Command{
		NasabahID:          nasabahID,
		RuleID:             ruleID,
		RuleType:           req.RuleType,
		AlertType:          req.AlertType,
		Severity:           req.Severity,
		Description:        req.Description,
		TransactionDetails: req.TransactionDetails,
	}
	if req.TransaksiID != nil {
		tid, parseErr := uuid.Parse(*req.TransaksiID)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "transaksi_id tidak valid")
			return
		}
		cmd.TransaksiID = &tid
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Investigate godoc
// @Summary      Start investigating AML alert
// @Tags         aml-alert
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Alert ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/aml/alerts/{id}/investigate [put]
func (h *AMLAlertHandler) Investigate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		InvestigatedBy     string `json:"investigated_by"`
		InvestigationNotes string `json:"investigation_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	investigatedBy, parseErr := uuid.Parse(req.InvestigatedBy)
	if parseErr != nil {
		h.RespondError(w, http.StatusBadRequest, "investigated_by tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), investigateamlalert.Command{
		ID:                 id,
		InvestigatedBy:     investigatedBy,
		InvestigationNotes: req.InvestigationNotes,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Resolve godoc
// @Summary      Resolve AML alert
// @Tags         aml-alert
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Alert ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/aml/alerts/{id}/resolve [put]
func (h *AMLAlertHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		InvestigationOutcome string  `json:"investigation_outcome"`
		InvestigationNotes   string  `json:"investigation_notes"`
		LTKMFiled           bool    `json:"ltkm_filed"`
		LTKMReference        *string `json:"ltkm_reference,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), resolveamlalert.Command{
		ID:                   id,
		InvestigationOutcome: req.InvestigationOutcome,
		InvestigationNotes:   req.InvestigationNotes,
		LTKMFiled:           req.LTKMFiled,
		LTKMReference:        req.LTKMReference,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// parseListFilter membaca filter dari query parameters.
func (h *AMLAlertHandler) parseListFilter(r *http.Request) aml_alert.ListFilter {
	var filter aml_alert.ListFilter
	q := r.URL.Query()

	if st := q.Get("status"); st != "" {
		filter.Status = &st
	}
	if sv := q.Get("severity"); sv != "" {
		filter.Severity = &sv
	}
	if at := q.Get("alert_type"); at != "" {
		filter.AlertType = &at
	}
	if rt := q.Get("rule_type"); rt != "" {
		filter.RuleType = &rt
	}
	if nid := q.Get("nasabah_id"); nid != "" {
		id, err := uuid.Parse(nid)
		if err == nil {
			filter.NasabahID = &id
		}
	}

	return filter
}
