package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/consent"
	createconsentrecord "github.com/yourorg/boilerplate/internal/command/create_consent_record"
	withdrawconsent "github.com/yourorg/boilerplate/internal/command/withdraw_consent"
	getconsentrecord "github.com/yourorg/boilerplate/internal/query/get_consent_record_by_id"
	listconsentrecords "github.com/yourorg/boilerplate/internal/query/list_consent_records"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// ConsentHandler menangani HTTP request untuk domain Consent.
type ConsentHandler struct {
	BaseHandler
}

var consentSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at":    "created_at",
		"consent_type":  "consent_type",
		"legal_basis":   "legal_basis",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewConsentHandler membuat instance ConsentHandler baru.
func NewConsentHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *ConsentHandler {
	return &ConsentHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Consent.
func (h *ConsentHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/consent-records", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}/withdraw", h.Withdraw)
	})
}

// List godoc
// @Summary      List consent records
// @Tags         consent
// @Produce      json
// @Param        consent_type   query  string  false  "Filter by consent type"
// @Param        legal_basis    query  string  false  "Filter by legal basis"
// @Param        consent_given  query  bool    false  "Filter by consent given"
// @Param        withdrawn      query  bool    false  "Filter by withdrawn status"
// @Param        nasabah_id     query  string  false  "Filter by nasabah ID"
// @Success      200  {object}  listconsentrecords.Result
// @Router       /api/v1/consent-records [get]
func (h *ConsentHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, consentSortConfig)
	filter := h.parseListFilter(r)
	result, err := querybus.Dispatch[*listconsentrecords.Result](r.Context(), h.queryBus, listconsentrecords.Query{
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
// @Summary      Get consent record by ID
// @Tags         consent
// @Produce      json
// @Param        id   path      string  true  "Consent Record ID"
// @Success      200  {object}  getconsentrecord.Result
// @Router       /api/v1/consent-records/{id} [get]
func (h *ConsentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getconsentrecord.Result](r.Context(), h.queryBus, getconsentrecord.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary      Create consent record
// @Tags         consent
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/consent-records [post]
func (h *ConsentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NasabahID          string  `json:"nasabah_id"`
		ConsentType        string  `json:"consent_type"`
		ConsentPurpose     string  `json:"consent_purpose"`
		LegalBasis         string  `json:"legal_basis"`
		ConsentText        string  `json:"consent_text"`
		ConsentVersion     string  `json:"consent_version"`
		ConsentGiven       bool    `json:"consent_given"`
		ConsentMethod      string  `json:"consent_method"`
		ParentID           *string `json:"parent_id,omitempty"`
		ParentRelationship *string `json:"parent_relationship,omitempty"`
		ParentConsentGiven *bool   `json:"parent_consent_given,omitempty"`
		IPAddress          *string `json:"ip_address,omitempty"`
		UserAgent          *string `json:"user_agent,omitempty"`
		WitnessID          *string `json:"witness_id,omitempty"`
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

	cmd := createconsentrecord.Command{
		NasabahID:          nasabahID,
		ConsentType:        req.ConsentType,
		ConsentPurpose:     req.ConsentPurpose,
		LegalBasis:         req.LegalBasis,
		ConsentText:        req.ConsentText,
		ConsentVersion:     req.ConsentVersion,
		ConsentGiven:       req.ConsentGiven,
		ConsentMethod:      req.ConsentMethod,
		ParentRelationship: req.ParentRelationship,
		ParentConsentGiven: req.ParentConsentGiven,
		IPAddress:          req.IPAddress,
		UserAgent:          req.UserAgent,
	}

	if req.ParentID != nil {
		pid, parseErr := uuid.Parse(*req.ParentID)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "parent_id tidak valid")
			return
		}
		cmd.ParentID = &pid
	}
	if req.WitnessID != nil {
		wid, parseErr := uuid.Parse(*req.WitnessID)
		if parseErr != nil {
			h.RespondError(w, http.StatusBadRequest, "witness_id tidak valid")
			return
		}
		cmd.WitnessID = &wid
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Withdraw godoc
// @Summary      Withdraw consent
// @Tags         consent
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Consent Record ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/consent-records/{id}/withdraw [put]
func (h *ConsentHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		WithdrawalReason *string `json:"withdrawal_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), withdrawconsent.Command{
		ID:               id,
		WithdrawalReason: req.WithdrawalReason,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// parseListFilter membaca filter dari query parameters.
func (h *ConsentHandler) parseListFilter(r *http.Request) consent.ListFilter {
	var filter consent.ListFilter
	q := r.URL.Query()

	if ct := q.Get("consent_type"); ct != "" {
		filter.ConsentType = &ct
	}
	if lb := q.Get("legal_basis"); lb != "" {
		filter.LegalBasis = &lb
	}
	if q.Get("consent_given") != "" {
		cg := q.Get("consent_given") == "true"
		filter.ConsentGiven = &cg
	}
	if q.Get("withdrawn") != "" {
		w := q.Get("withdrawn") == "true"
		filter.Withdrawn = &w
	}
	if nid := q.Get("nasabah_id"); nid != "" {
		id, err := uuid.Parse(nid)
		if err == nil {
			filter.NasabahID = &id
		}
	}

	return filter
}
