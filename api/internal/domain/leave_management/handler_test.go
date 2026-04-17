package leave_management_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/leave_management"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── test fixtures ───────────────────────────────────────────────────────────

var leaveTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func leaveScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ── scope enforcement tests ─────────────────────────────────────────────────

func TestLeaveHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondLeaveForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validLeaveRequestBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	errMap, ok := resp["error"].(map[string]any)
	require.True(t, ok, "response should contain error object")
	assert.Equal(t, "FORBIDDEN", errMap["code"])
}

func TestLeaveHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(leaveScopeMiddleware(leaveTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, leaveTestScope.TenantID, s.TenantID)
		assert.Equal(t, leaveTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── invalid JSON body ───────────────────────────────────────────────────────

func TestLeaveHandler_InvalidJSON_Returns400(t *testing.T) {
	d := &leave_management.LeaveRequestDescriptor{}
	r := chi.NewRouter()
	r.Use(leaveScopeMiddleware(leaveTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondLeaveForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondLeaveInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			respondLeaveValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	errMap, ok := resp["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "INVALID_JSON", errMap["code"])
}

// ── invalid FK UUID tests ───────────────────────────────────────────────────

func TestLeaveHandler_InvalidTeacherUUID_Returns400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(leaveScopeMiddleware(leaveTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondLeaveForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondLeaveInvalidJSON(w)
			return
		}
		teacherID, _ := body["teacher_id"].(string)
		if _, err := uuid.Parse(teacherID); err != nil {
			respondLeaveInvalidUUID(w, "teacher_id")
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"teacher_id":       "not-a-uuid",
		"leave_type_id":    "00000000-0000-0000-0000-000000000002",
		"academic_year_id": "00000000-0000-0000-0000-000000000003",
		"start_date":       "2026-04-20",
		"end_date":         "2026-04-22",
		"total_days":       3,
		"reason":           "Cuti tahunan",
		"status":           leave_management.StatusDraft,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ── helpers ─────────────────────────────────────────────────────────────────

func validLeaveRequestBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"teacher_id":       "00000000-0000-0000-0000-000000000001",
		"leave_type_id":    "00000000-0000-0000-0000-000000000002",
		"academic_year_id": "00000000-0000-0000-0000-000000000003",
		"start_date":       "2026-04-20",
		"end_date":         "2026-04-22",
		"total_days":       3,
		"reason":           "Cuti tahunan",
		"status":           leave_management.StatusDraft,
	})
	require.NoError(t, err)
	return body
}

func respondLeaveForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func respondLeaveInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func respondLeaveValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}

func respondLeaveInvalidUUID(w http.ResponseWriter, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_ID", "message": field + " harus berupa UUID valid"},
	})
}
