package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	approveinsuranceclaim "github.com/yourorg/boilerplate/internal/command/approve_insurance_claim"
	cancelinsurancepolicy "github.com/yourorg/boilerplate/internal/command/cancel_insurance_policy"
	createinsuranceclaim "github.com/yourorg/boilerplate/internal/command/create_insurance_claim"
	createinsurancepolicy "github.com/yourorg/boilerplate/internal/command/create_insurance_policy"
	createinsuranceproduct "github.com/yourorg/boilerplate/internal/command/create_insurance_product"
	rejectinsuranceclaim "github.com/yourorg/boilerplate/internal/command/reject_insurance_claim"
	reviewinsuranceclaim "github.com/yourorg/boilerplate/internal/command/review_insurance_claim"
	settleinsuranceclaim "github.com/yourorg/boilerplate/internal/command/settle_insurance_claim"
	updateinsuranceproduct "github.com/yourorg/boilerplate/internal/command/update_insurance_product"
	getinsurancepolicy "github.com/yourorg/boilerplate/internal/query/get_insurance_policy_by_id"
	getinsuranceproduct "github.com/yourorg/boilerplate/internal/query/get_insurance_product_by_id"
	listinsuranceclaims "github.com/yourorg/boilerplate/internal/query/list_insurance_claims"
	listinsurancepolicies "github.com/yourorg/boilerplate/internal/query/list_insurance_policies"
	listinsuranceproducts "github.com/yourorg/boilerplate/internal/query/list_insurance_products"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// InsuranceHandler menangani HTTP request untuk domain Insurance.
type InsuranceHandler struct {
	BaseHandler
}

var insuranceSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"name":       "name",
		"code":       "code",
		"created_at": "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

var insurancePolicySortConfig = SortConfig{
	AllowedFields: map[string]string{
		"policy_no":  "policy_no",
		"status":     "status",
		"created_at": "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

var insuranceClaimSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"status":       "status",
		"submitted_at": "submitted_at",
		"created_at":   "created_at",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewInsuranceHandler membuat instance InsuranceHandler baru.
func NewInsuranceHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *InsuranceHandler {
	return &InsuranceHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Insurance.
func (h *InsuranceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/insurance", func(r chi.Router) {
		// Products
		r.Get("/products", h.ListProducts)
		r.Post("/products", h.CreateProduct)
		r.Get("/products/{id}", h.GetProductByID)
		r.Put("/products/{id}", h.UpdateProduct)

		// Policies
		r.Get("/policies", h.ListPolicies)
		r.Post("/policies", h.CreatePolicy)
		r.Get("/policies/{id}", h.GetPolicyByID)
		r.Put("/policies/{id}/cancel", h.CancelPolicy)

		// Claims
		r.Get("/claims", h.ListClaims)
		r.Post("/claims", h.CreateClaim)
		r.Put("/claims/{id}/review", h.ReviewClaim)
		r.Put("/claims/{id}/approve", h.ApproveClaim)
		r.Put("/claims/{id}/reject", h.RejectClaim)
		r.Put("/claims/{id}/settle", h.SettleClaim)
	})
}

// ── Products ──────────────────────────────────────────────────────────────────

// ListProducts godoc
// @Summary      List insurance products
// @Tags         insurance
// @Produce      json
// @Param        product_type  query  string  false  "Filter by product type"
// @Param        status        query  string  false  "Filter by status"
// @Success      200  {object}  listinsuranceproducts.Result
// @Router       /api/v1/insurance/products [get]
func (h *InsuranceHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, insuranceSortConfig)
	filter := h.parseProductFilter(r)
	result, err := querybus.Dispatch[*listinsuranceproducts.Result](r.Context(), h.queryBus, listinsuranceproducts.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetProductByID godoc
// @Summary      Get insurance product by ID
// @Tags         insurance
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  getinsuranceproduct.Result
// @Router       /api/v1/insurance/products/{id} [get]
func (h *InsuranceHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getinsuranceproduct.Result](r.Context(), h.queryBus, getinsuranceproduct.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateProduct godoc
// @Summary      Create insurance product
// @Tags         insurance
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/insurance/products [post]
func (h *InsuranceHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string `json:"name"`
		Code              string `json:"code"`
		ProductType       string `json:"product_type"`
		ProviderName      string `json:"provider_name"`
		PremiumAmount     int64  `json:"premium_amount"`
		CoverageAmount    int64  `json:"coverage_amount"`
		PremiumFrequency  string `json:"premium_frequency"`
		TermMonths        int    `json:"term_months"`
		Description       string `json:"description"`
		TermsConditions   string `json:"terms_conditions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createinsuranceproduct.Command{
		Name:             req.Name,
		Code:             req.Code,
		ProductType:      insurance.ProductType(req.ProductType),
		ProviderName:     req.ProviderName,
		PremiumAmount:    req.PremiumAmount,
		CoverageAmount:   req.CoverageAmount,
		PremiumFrequency: req.PremiumFrequency,
		TermMonths:       req.TermMonths,
		Description:      req.Description,
		TermsConditions:  req.TermsConditions,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// UpdateProduct godoc
// @Summary      Update insurance product
// @Tags         insurance
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/insurance/products/{id} [put]
func (h *InsuranceHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Name              string `json:"name"`
		Code              string `json:"code"`
		ProductType       string `json:"product_type"`
		Status            string `json:"status"`
		ProviderName      string `json:"provider_name"`
		PremiumAmount     int64  `json:"premium_amount"`
		CoverageAmount    int64  `json:"coverage_amount"`
		PremiumFrequency  string `json:"premium_frequency"`
		TermMonths        int    `json:"term_months"`
		Description       string `json:"description"`
		TermsConditions   string `json:"terms_conditions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), updateinsuranceproduct.Command{
		ID:               id,
		Name:             req.Name,
		Code:             req.Code,
		ProductType:      insurance.ProductType(req.ProductType),
		Status:           insurance.ProductStatus(req.Status),
		ProviderName:     req.ProviderName,
		PremiumAmount:    req.PremiumAmount,
		CoverageAmount:   req.CoverageAmount,
		PremiumFrequency: req.PremiumFrequency,
		TermMonths:       req.TermMonths,
		Description:      req.Description,
		TermsConditions:  req.TermsConditions,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ── Policies ──────────────────────────────────────────────────────────────────

// ListPolicies godoc
// @Summary      List insurance policies
// @Tags         insurance
// @Produce      json
// @Param        product_id  query  string  false  "Filter by product ID"
// @Param        nasabah_id  query  string  false  "Filter by nasabah ID"
// @Param        status      query  string  false  "Filter by status"
// @Success      200  {object}  listinsurancepolicies.Result
// @Router       /api/v1/insurance/policies [get]
func (h *InsuranceHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, insurancePolicySortConfig)
	filter := h.parsePolicyFilter(r)
	result, err := querybus.Dispatch[*listinsurancepolicies.Result](r.Context(), h.queryBus, listinsurancepolicies.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetPolicyByID godoc
// @Summary      Get insurance policy by ID
// @Tags         insurance
// @Produce      json
// @Param        id   path      string  true  "Policy ID"
// @Success      200  {object}  getinsurancepolicy.Result
// @Router       /api/v1/insurance/policies/{id} [get]
func (h *InsuranceHandler) GetPolicyByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getinsurancepolicy.Result](r.Context(), h.queryBus, getinsurancepolicy.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreatePolicy godoc
// @Summary      Create insurance policy
// @Tags         insurance
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/insurance/policies [post]
func (h *InsuranceHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID           string `json:"product_id"`
		NasabahID           string `json:"nasabah_id"`
		PolicyNo            string `json:"policy_no"`
		StartDate           string `json:"start_date"`
		EndDate             string `json:"end_date"`
		PremiumAmount       int64  `json:"premium_amount"`
		CoverageAmount      int64  `json:"coverage_amount"`
		BeneficiaryName     string `json:"beneficiary_name"`
		BeneficiaryRelation string `json:"beneficiary_relation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "product_id tidak valid")
		return
	}
	nasabahID, err := uuid.Parse(req.NasabahID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "nasabah_id tidak valid")
		return
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "start_date tidak valid")
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "end_date tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createinsurancepolicy.Command{
		ProductID:           productID,
		NasabahID:           nasabahID,
		PolicyNo:            req.PolicyNo,
		StartDate:           startDate,
		EndDate:             endDate,
		PremiumAmount:       req.PremiumAmount,
		CoverageAmount:      req.CoverageAmount,
		BeneficiaryName:     req.BeneficiaryName,
		BeneficiaryRelation: req.BeneficiaryRelation,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// CancelPolicy godoc
// @Summary      Cancel insurance policy
// @Tags         insurance
// @Param        id   path      string  true  "Policy ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/insurance/policies/{id}/cancel [put]
func (h *InsuranceHandler) CancelPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), cancelinsurancepolicy.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ── Claims ────────────────────────────────────────────────────────────────────

// ListClaims godoc
// @Summary      List insurance claims
// @Tags         insurance
// @Produce      json
// @Param        policy_id   query  string  false  "Filter by policy ID"
// @Param        claim_type  query  string  false  "Filter by claim type"
// @Param        status      query  string  false  "Filter by status"
// @Success      200  {object}  listinsuranceclaims.Result
// @Router       /api/v1/insurance/claims [get]
func (h *InsuranceHandler) ListClaims(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, insuranceClaimSortConfig)
	filter := h.parseClaimFilter(r)
	result, err := querybus.Dispatch[*listinsuranceclaims.Result](r.Context(), h.queryBus, listinsuranceclaims.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateClaim godoc
// @Summary      Create insurance claim
// @Tags         insurance
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/insurance/claims [post]
func (h *InsuranceHandler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PolicyID     string   `json:"policy_id"`
		ClaimType    string   `json:"claim_type"`
		ClaimAmount  int64    `json:"claim_amount"`
		IncidentDate string   `json:"incident_date"`
		SubmittedBy  string   `json:"submitted_by"`
		Description  string   `json:"description"`
		DocumentIDs  []string `json:"document_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	policyID, err := uuid.Parse(req.PolicyID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "policy_id tidak valid")
		return
	}
	submittedBy, err := uuid.Parse(req.SubmittedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "submitted_by tidak valid")
		return
	}
	incidentDate, err := time.Parse("2006-01-02", req.IncidentDate)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "incident_date tidak valid")
		return
	}

	docIDs := make([]uuid.UUID, 0, len(req.DocumentIDs))
	for _, d := range req.DocumentIDs {
		docID, parseErr := uuid.Parse(d)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "document_id tidak valid")
			return
		}
		docIDs = append(docIDs, docID)
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createinsuranceclaim.Command{
		PolicyID:     policyID,
		ClaimType:    insurance.ClaimType(req.ClaimType),
		ClaimAmount:  req.ClaimAmount,
		IncidentDate: incidentDate,
		SubmittedBy:  submittedBy,
		Description:  req.Description,
		DocumentIDs:  docIDs,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// ReviewClaim godoc
// @Summary      Review insurance claim
// @Tags         insurance
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Claim ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/insurance/claims/{id}/review [put]
func (h *InsuranceHandler) ReviewClaim(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		ReviewedBy string `json:"reviewed_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	reviewedBy, err := uuid.Parse(req.ReviewedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "reviewed_by tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), reviewinsuranceclaim.Command{
		ID:         id,
		ReviewedBy: reviewedBy,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ApproveClaim godoc
// @Summary      Approve insurance claim
// @Tags         insurance
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Claim ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/insurance/claims/{id}/approve [put]
func (h *InsuranceHandler) ApproveClaim(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		ApprovedAmount int64 `json:"approved_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), approveinsuranceclaim.Command{
		ID:             id,
		ApprovedAmount: req.ApprovedAmount,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// RejectClaim godoc
// @Summary      Reject insurance claim
// @Tags         insurance
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Claim ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/insurance/claims/{id}/reject [put]
func (h *InsuranceHandler) RejectClaim(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		RejectionReason string `json:"rejection_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), rejectinsuranceclaim.Command{
		ID:              id,
		RejectionReason: req.RejectionReason,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// SettleClaim godoc
// @Summary      Settle insurance claim payment
// @Tags         insurance
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Claim ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/insurance/claims/{id}/settle [put]
func (h *InsuranceHandler) SettleClaim(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		PaidBy string `json:"paid_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	paidBy, err := uuid.Parse(req.PaidBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "paid_by tidak valid")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), settleinsuranceclaim.Command{
		ID:     id,
		PaidBy: paidBy,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ── Filter Helpers ────────────────────────────────────────────────────────────

func (h *InsuranceHandler) parseProductFilter(r *http.Request) insurance.ProductFilter {
	var filter insurance.ProductFilter
	q := r.URL.Query()

	if pt := q.Get("product_type"); pt != "" {
		t := insurance.ProductType(pt)
		filter.ProductType = &t
	}
	if st := q.Get("status"); st != "" {
		s := insurance.ProductStatus(st)
		filter.Status = &s
	}

	return filter
}

func (h *InsuranceHandler) parsePolicyFilter(r *http.Request) insurance.PolicyFilter {
	var filter insurance.PolicyFilter
	q := r.URL.Query()

	if pid := q.Get("product_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err == nil {
			filter.ProductID = &id
		}
	}
	if nid := q.Get("nasabah_id"); nid != "" {
		id, err := uuid.Parse(nid)
		if err == nil {
			filter.NasabahID = &id
		}
	}
	if st := q.Get("status"); st != "" {
		s := insurance.PolicyStatus(st)
		filter.Status = &s
	}

	return filter
}

func (h *InsuranceHandler) parseClaimFilter(r *http.Request) insurance.ClaimFilter {
	var filter insurance.ClaimFilter
	q := r.URL.Query()

	if pid := q.Get("policy_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err == nil {
			filter.PolicyID = &id
		}
	}
	if ct := q.Get("claim_type"); ct != "" {
		t := insurance.ClaimType(ct)
		filter.ClaimType = &t
	}
	if st := q.Get("status"); st != "" {
		s := insurance.ClaimStatus(st)
		filter.Status = &s
	}

	return filter
}
