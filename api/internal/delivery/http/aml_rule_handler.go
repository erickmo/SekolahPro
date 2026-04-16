package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_rule"
	createamlrule "github.com/yourorg/boilerplate/internal/command/create_aml_rule"
	toggleamlrule "github.com/yourorg/boilerplate/internal/command/toggle_aml_rule"
	updateamlrule "github.com/yourorg/boilerplate/internal/command/update_aml_rule"
	getamlrule "github.com/yourorg/boilerplate/internal/query/get_aml_rule_by_id"
	listamlrules "github.com/yourorg/boilerplate/internal/query/list_aml_rules"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// AMLRuleHandler menangani HTTP request untuk domain AML Rule.
type AMLRuleHandler struct {
	BaseHandler
}

var amlRuleSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at": "created_at",
		"rule_code":  "rule_code",
		"rule_type":  "rule_type",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewAMLRuleHandler membuat instance AMLRuleHandler baru.
func NewAMLRuleHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *AMLRuleHandler {
	return &AMLRuleHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain AML Rule.
func (h *AMLRuleHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/aml/rules", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Put("/{id}/toggle", h.Toggle)
	})
}

// List godoc
// @Summary      List AML monitoring rules
// @Tags         aml-rule
// @Produce      json
// @Param        rule_type  query  string  false  "Filter by rule type"
// @Param        is_active  query  bool    false  "Filter by active status"
// @Success      200  {object}  listamlrules.Result
// @Router       /api/v1/aml/rules [get]
func (h *AMLRuleHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, amlRuleSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listamlrules.Result](r.Context(), h.queryBus, listamlrules.Query{
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
// @Summary      Get AML rule by ID
// @Tags         aml-rule
// @Produce      json
// @Param        id   path      string  true  "Rule ID"
// @Success      200  {object}  getamlrule.Result
// @Router       /api/v1/aml/rules/{id} [get]
func (h *AMLRuleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getamlrule.Result](r.Context(), h.queryBus, getamlrule.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create AML monitoring rule
// @Tags         aml-rule
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/aml/rules [post]
func (h *AMLRuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RuleCode    string          `json:"rule_code"`
		RuleName    string          `json:"rule_name"`
		RuleType    string          `json:"rule_type"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
		IsActive    bool            `json:"is_active"`
		AppliesTo   string          `json:"applies_to"`
		AutoAlert   bool            `json:"auto_alert"`
		AutoBlock   bool            `json:"auto_block"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	cmd := createamlrule.Command{
		RuleCode:    req.RuleCode,
		RuleName:    req.RuleName,
		RuleType:    req.RuleType,
		Description: req.Description,
		Parameters:  req.Parameters,
		IsActive:    req.IsActive,
		AppliesTo:   req.AppliesTo,
		AutoAlert:   req.AutoAlert,
		AutoBlock:   req.AutoBlock,
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Update godoc
// @Summary      Update AML monitoring rule
// @Tags         aml-rule
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Rule ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/aml/rules/{id} [put]
func (h *AMLRuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		RuleName    string          `json:"rule_name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
		AppliesTo   string          `json:"applies_to"`
		AutoAlert   bool            `json:"auto_alert"`
		AutoBlock   bool            `json:"auto_block"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), updateamlrule.Command{
		ID:          id,
		RuleName:    req.RuleName,
		Description: req.Description,
		Parameters:  req.Parameters,
		AppliesTo:   req.AppliesTo,
		AutoAlert:   req.AutoAlert,
		AutoBlock:   req.AutoBlock,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Toggle godoc
// @Summary      Toggle AML monitoring rule active status
// @Tags         aml-rule
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Rule ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/aml/rules/{id}/toggle [put]
func (h *AMLRuleHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), toggleamlrule.Command{
		ID:       id,
		IsActive: req.IsActive,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// parseListFilter membaca filter dari query parameters.
func (h *AMLRuleHandler) parseListFilter(r *http.Request) aml_rule.ListFilter {
	var filter aml_rule.ListFilter
	q := r.URL.Query()

	if rt := q.Get("rule_type"); rt != "" {
		filter.RuleType = &rt
	}
	if q.Get("is_active") != "" {
		isActive := q.Get("is_active") == "true"
		filter.IsActive = &isActive
	}

	return filter
}
