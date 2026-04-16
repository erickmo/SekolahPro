package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_cdd"
	createamlcdd "github.com/yourorg/boilerplate/internal/command/create_aml_cdd"
	updateamlcdd "github.com/yourorg/boilerplate/internal/command/update_aml_cdd"
	updateriskamlcdd "github.com/yourorg/boilerplate/internal/command/update_aml_cdd_risk"
	getamlcdd "github.com/yourorg/boilerplate/internal/query/get_aml_cdd_by_id"
	listamlcdds "github.com/yourorg/boilerplate/internal/query/list_aml_cdds"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// AMLCDDHandler menangani HTTP request untuk domain AML CDD.
type AMLCDDHandler struct {
	BaseHandler
}

var amlCDDSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at":  "created_at",
		"risk_level":  "risk_level",
		"risk_score":  "risk_score",
		"status":      "status",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewAMLCDDHandler membuat instance AMLCDDHandler baru.
func NewAMLCDDHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *AMLCDDHandler {
	return &AMLCDDHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain AML CDD.
func (h *AMLCDDHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/aml/cdds", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Put("/{id}/risk", h.UpdateRisk)
	})
}

// List godoc
// @Summary      List AML CDD records
// @Tags         aml-cdd
// @Produce      json
// @Param        risk_level  query  string  false  "Filter by risk level"
// @Param        cdd_level   query  string  false  "Filter by CDD level"
// @Param        status      query  string  false  "Filter by status"
// @Param        is_pep      query  bool    false  "Filter by PEP status"
// @Success      200  {object}  listamlcdds.Result
// @Router       /api/v1/aml/cdds [get]
func (h *AMLCDDHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, amlCDDSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listamlcdds.Result](r.Context(), h.queryBus, listamlcdds.Query{
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
// @Summary      Get AML CDD by ID
// @Tags         aml-cdd
// @Produce      json
// @Param        id   path      string  true  "CDD Record ID"
// @Success      200  {object}  getamlcdd.Result
// @Router       /api/v1/aml/cdds/{id} [get]
func (h *AMLCDDHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getamlcdd.Result](r.Context(), h.queryBus, getamlcdd.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create AML CDD record
// @Tags         aml-cdd
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/aml/cdds [post]
func (h *AMLCDDHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NasabahID              string           `json:"nasabah_id"`
		RiskLevel              string           `json:"risk_level"`
		RiskScore              int              `json:"risk_score"`
		RiskCategory           json.RawMessage  `json:"risk_category"`
		CDDLevel               string           `json:"cdd_level"`
		CDDPurpose             string           `json:"cdd_purpose"`
		SourceOfFunds          *string          `json:"source_of_funds,omitempty"`
		SourceOfWealth         *string          `json:"source_of_wealth,omitempty"`
		IsPEP                  bool             `json:"is_pep"`
		PEPType                string           `json:"pep_type"`
		PEPPosition            *string          `json:"pep_position,omitempty"`
		PEPCountry             *string          `json:"pep_country,omitempty"`
		PEPRelationship        string           `json:"pep_relationship"`
		BeneficialOwnerName    *string          `json:"beneficial_owner_name,omitempty"`
		BeneficialOwnerIDNo    *string          `json:"beneficial_owner_id_no,omitempty"`
		BeneficialOwnershipPct *float64         `json:"beneficial_ownership_pct,omitempty"`
		BOVerified             bool             `json:"bo_verified"`
		CreatedBy              string           `json:"created_by"`
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
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "created_by tidak valid")
		return
	}

	cmd := createamlcdd.Command{
		NasabahID:              nasabahID,
		RiskLevel:              req.RiskLevel,
		RiskScore:              req.RiskScore,
		RiskCategory:           req.RiskCategory,
		CDDLevel:               req.CDDLevel,
		CDDPurpose:             req.CDDPurpose,
		SourceOfFunds:          req.SourceOfFunds,
		SourceOfWealth:         req.SourceOfWealth,
		IsPEP:                  req.IsPEP,
		PEPType:                req.PEPType,
		PEPPosition:            req.PEPPosition,
		PEPCountry:             req.PEPCountry,
		PEPRelationship:        req.PEPRelationship,
		BeneficialOwnerName:    req.BeneficialOwnerName,
		BeneficialOwnerIDNo:    req.BeneficialOwnerIDNo,
		BeneficialOwnershipPct: req.BeneficialOwnershipPct,
		BOVerified:             req.BOVerified,
		CreatedBy:              createdBy,
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Update godoc
// @Summary      Update AML CDD record
// @Tags         aml-cdd
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "CDD Record ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/aml/cdds/{id} [put]
func (h *AMLCDDHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		CDDPurpose             string  `json:"cdd_purpose"`
		SourceOfFunds          *string `json:"source_of_funds,omitempty"`
		SourceOfWealth         *string `json:"source_of_wealth,omitempty"`
		IsPEP                  bool    `json:"is_pep"`
		PEPType                string  `json:"pep_type"`
		PEPPosition            *string `json:"pep_position,omitempty"`
		PEPCountry             *string `json:"pep_country,omitempty"`
		PEPRelationship        string  `json:"pep_relationship"`
		BeneficialOwnerName    *string `json:"beneficial_owner_name,omitempty"`
		BeneficialOwnerIDNo    *string `json:"beneficial_owner_id_no,omitempty"`
		BeneficialOwnershipPct *float64 `json:"beneficial_ownership_pct,omitempty"`
		BOVerified             bool    `json:"bo_verified"`
		Status                 string  `json:"status"`
		RestrictionReason      *string `json:"restriction_reason,omitempty"`
		ExitReason             *string `json:"exit_reason,omitempty"`
		UpdatedBy              string  `json:"updated_by"`
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
	if err := h.commandBus.Dispatch(r.Context(), updateamlcdd.Command{
		ID:                     id,
		CDDPurpose:             req.CDDPurpose,
		SourceOfFunds:          req.SourceOfFunds,
		SourceOfWealth:         req.SourceOfWealth,
		IsPEP:                  req.IsPEP,
		PEPType:                req.PEPType,
		PEPPosition:            req.PEPPosition,
		PEPCountry:             req.PEPCountry,
		PEPRelationship:        req.PEPRelationship,
		BeneficialOwnerName:    req.BeneficialOwnerName,
		BeneficialOwnerIDNo:    req.BeneficialOwnerIDNo,
		BeneficialOwnershipPct: req.BeneficialOwnershipPct,
		BOVerified:             req.BOVerified,
		Status:                 req.Status,
		RestrictionReason:      req.RestrictionReason,
		ExitReason:             req.ExitReason,
		UpdatedBy:              updatedBy,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// UpdateRisk godoc
// @Summary      Update AML CDD risk assessment
// @Tags         aml-cdd
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "CDD Record ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/aml/cdds/{id}/risk [put]
func (h *AMLCDDHandler) UpdateRisk(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		RiskLevel      string          `json:"risk_level"`
		RiskScore      int             `json:"risk_score"`
		RiskCategory   json.RawMessage `json:"risk_category"`
		CDDLevel       string          `json:"cdd_level"`
		NextReviewDate *string         `json:"next_review_date,omitempty"`
		UpdatedBy      string          `json:"updated_by"`
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

	cmd := updateriskamlcdd.Command{
		ID:           id,
		RiskLevel:    req.RiskLevel,
		RiskScore:    req.RiskScore,
		RiskCategory: req.RiskCategory,
		CDDLevel:     req.CDDLevel,
		UpdatedBy:    updatedBy,
	}
	if req.NextReviewDate != nil {
		nrd, parseErr := time.Parse("2006-01-02", *req.NextReviewDate)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "next_review_date tidak valid")
			return
		}
		cmd.NextReviewDate = &nrd
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// parseListFilter membaca filter dari query parameters.
func (h *AMLCDDHandler) parseListFilter(r *http.Request) aml_cdd.ListFilter {
	var filter aml_cdd.ListFilter
	q := r.URL.Query()

	if rl := q.Get("risk_level"); rl != "" {
		filter.RiskLevel = &rl
	}
	if cl := q.Get("cdd_level"); cl != "" {
		filter.CDDLevel = &cl
	}
	if st := q.Get("status"); st != "" {
		filter.Status = &st
	}
	if q.Get("is_pep") != "" {
		isPEP := q.Get("is_pep") == "true"
		filter.IsPEP = &isPEP
	}

	return filter
}
