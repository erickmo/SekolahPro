package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	enrollbiometric "github.com/yourorg/boilerplate/internal/command/enroll_biometric"
	verifybiometric "github.com/yourorg/boilerplate/internal/command/verify_biometric"
	deactivatebiometric "github.com/yourorg/boilerplate/internal/command/deactivate_biometric"
	createverificationlog "github.com/yourorg/boilerplate/internal/command/create_verification_log"
	"github.com/yourorg/boilerplate/internal/domain/biometric"
	"github.com/yourorg/boilerplate/internal/domain/biometric_log"
	getbiometric "github.com/yourorg/boilerplate/internal/query/get_biometric_by_id"
	listbiometrics "github.com/yourorg/boilerplate/internal/query/list_biometrics"
	getbiometriclog "github.com/yourorg/boilerplate/internal/query/get_biometric_log_by_id"
	listverificationlogs "github.com/yourorg/boilerplate/internal/query/list_verification_logs"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// BiometricHandler menangani HTTP request untuk domain Biometric Enrollment dan Verification Log.
type BiometricHandler struct {
	BaseHandler
}

var biometricSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at":     "created_at",
		"nasabah_id":     "nasabah_id",
		"biometric_type": "biometric_type",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

var logSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"created_at":          "created_at",
		"attempted_at":        "attempted_at",
		"verification_result": "verification_result",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewBiometricHandler membuat instance BiometricHandler baru.
func NewBiometricHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *BiometricHandler {
	return &BiometricHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Biometric.
func (h *BiometricHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/biometric", func(r chi.Router) {
		// Enrollment routes
		r.Route("/enrollments", func(r chi.Router) {
			r.Get("/", h.ListEnrollments)
			r.Post("/", h.Enroll)
			r.Get("/{id}", h.GetEnrollmentByID)
			r.Put("/{id}/verify", h.Verify)
			r.Put("/{id}/deactivate", h.Deactivate)
			r.Delete("/{id}", h.DeleteEnrollment)
		})
		// Verification log routes
		r.Route("/logs", func(r chi.Router) {
			r.Get("/", h.ListLogs)
			r.Post("/", h.CreateLog)
			r.Get("/{id}", h.GetLogByID)
		})
	})
}

// ── Enrollment Handlers ───────────────────────────────────────────────────────

// ListEnrollments godoc
// @Summary      List biometric enrollments
// @Tags         biometric
// @Produce      json
// @Param        nasabah_id      query    string  false  "Filter by Nasabah ID"
// @Param        biometric_type  query    string  false  "Filter by type (fingerprint, face, voice)"
// @Param        is_active       query    bool    false  "Filter by active status"
// @Success      200  {object}  listbiometrics.Result
// @Router       /api/v1/biometric/enrollments [get]
func (h *BiometricHandler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, biometricSortConfig)
	filter := listbiometrics.BuildFilter(
		r.URL.Query().Get("nasabah_id"),
		r.URL.Query().Get("biometric_type"),
		r.URL.Query().Get("is_active"),
	)
	result, err := querybus.Dispatch[*listbiometrics.Result](r.Context(), h.queryBus, listbiometrics.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetEnrollmentByID godoc
// @Summary      Get biometric enrollment by ID
// @Tags         biometric
// @Produce      json
// @Param        id   path      string  true  "Enrollment ID"
// @Success      200  {object}  getbiometric.Result
// @Router       /api/v1/biometric/enrollments/{id} [get]
func (h *BiometricHandler) GetEnrollmentByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getbiometric.Result](r.Context(), h.queryBus, getbiometric.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// Enroll godoc
// @Summary      Enroll biometric
// @Tags         biometric
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/biometric/enrollments [post]
func (h *BiometricHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NasabahID     string `json:"nasabah_id"`
		BiometricType string `json:"biometric_type"`
		DeviceInfo    string `json:"device_info"`
		TemplateHash  string `json:"template_hash"`
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
	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, enrollbiometric.Command{
		NasabahID:     nasabahID,
		BiometricType: biometric.BiometricType(req.BiometricType),
		DeviceInfo:    req.DeviceInfo,
		TemplateHash:  req.TemplateHash,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// Verify godoc
// @Summary      Verify biometric
// @Tags         biometric
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Enrollment ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/biometric/enrollments/{id}/verify [put]
func (h *BiometricHandler) Verify(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		IsVerified bool   `json:"is_verified"`
		VerifiedBy string `json:"verified_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	verifiedBy, err := uuid.Parse(req.VerifiedBy)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "verified_by tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), verifybiometric.Command{
		ID:         id,
		IsVerified: req.IsVerified,
		VerifiedBy: verifiedBy,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// Deactivate godoc
// @Summary      Deactivate biometric enrollment
// @Tags         biometric
// @Produce      json
// @Param        id   path      string  true  "Enrollment ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/biometric/enrollments/{id}/deactivate [put]
func (h *BiometricHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deactivatebiometric.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// DeleteEnrollment godoc
// @Summary      Delete biometric enrollment (soft delete)
// @Tags         biometric
// @Param        id   path      string  true  "Enrollment ID"
// @Success      204
// @Router       /api/v1/biometric/enrollments/{id} [delete]
func (h *BiometricHandler) DeleteEnrollment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deactivatebiometric.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Verification Log Handlers ─────────────────────────────────────────────────

// ListLogs godoc
// @Summary      List verification logs
// @Tags         biometric
// @Produce      json
// @Param        enrollment_id        query    string  false  "Filter by Enrollment ID"
// @Param        nasabah_id           query    string  false  "Filter by Nasabah ID"
// @Param        verification_result  query    string  false  "Filter by result (success, failed, fallback)"
// @Success      200  {object}  listverificationlogs.Result
// @Router       /api/v1/biometric/logs [get]
func (h *BiometricHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, logSortConfig)
	filter := listverificationlogs.BuildFilter(
		r.URL.Query().Get("enrollment_id"),
		r.URL.Query().Get("nasabah_id"),
		r.URL.Query().Get("verification_result"),
	)
	result, err := querybus.Dispatch[*listverificationlogs.Result](r.Context(), h.queryBus, listverificationlogs.Query{
		Params: params,
		Filter: filter,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetLogByID godoc
// @Summary      Get verification log by ID
// @Tags         biometric
// @Produce      json
// @Param        id   path      string  true  "Log ID"
// @Success      200  {object}  getbiometriclog.Result
// @Router       /api/v1/biometric/logs/{id} [get]
func (h *BiometricHandler) GetLogByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getbiometriclog.Result](r.Context(), h.queryBus, getbiometriclog.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateLog godoc
// @Summary      Create verification log
// @Tags         biometric
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/biometric/logs [post]
func (h *BiometricHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EnrollmentID       string  `json:"enrollment_id"`
		NasabahID          string  `json:"nasabah_id"`
		VerificationResult string  `json:"verification_result"`
		FallbackMethod     *string `json:"fallback_method"`
		DeviceInfo         string  `json:"device_info"`
		IPAddress          string  `json:"ip_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	enrollmentID, err := uuid.Parse(req.EnrollmentID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "enrollment_id tidak valid")
		return
	}
	nasabahID, err := uuid.Parse(req.NasabahID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "nasabah_id tidak valid")
		return
	}

	var cmdFallbackMethod *biometric_log.FallbackMethod
	if req.FallbackMethod != nil {
		fm := biometric_log.FallbackMethod(*req.FallbackMethod)
		cmdFallbackMethod = &fm
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createverificationlog.Command{
		EnrollmentID:       enrollmentID,
		NasabahID:          nasabahID,
		VerificationResult: biometric_log.VerificationResult(req.VerificationResult),
		FallbackMethod:     cmdFallbackMethod,
		DeviceInfo:         req.DeviceInfo,
		IPAddress:          req.IPAddress,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}
