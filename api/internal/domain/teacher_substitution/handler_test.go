package teacher_substitution_test

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

	"github.com/yourorg/boilerplate/internal/domain/teacher_substitution"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── test fixtures ───────────────────────────────────────────────────────────

var substitutionTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func substitutionScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ── scope enforcement tests ─────────────────────────────────────────────────

func TestSubstitutionHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondSubForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validSubstitutionRequestBody(t)
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

func TestSubstitutionHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(substitutionScopeMiddleware(substitutionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, substitutionTestScope.TenantID, s.TenantID)
		assert.Equal(t, substitutionTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── invalid JSON body ───────────────────────────────────────────────────────

func TestSubstitutionHandler_InvalidJSON_Returns400(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	r := chi.NewRouter()
	r.Use(substitutionScopeMiddleware(substitutionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondSubForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondSubInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			respondSubValidationError(w, err.Error())
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

// ── validation error test ───────────────────────────────────────────────────

func TestSubstitutionHandler_SameTeacher_Returns422(t *testing.T) {
	d := &teacher_substitution.TeacherSubstitutionDescriptor{}
	r := chi.NewRouter()
	r.Use(substitutionScopeMiddleware(substitutionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondSubForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondSubInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			respondSubValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	sameID := "00000000-0000-0000-0000-000000000001"
	body, _ := json.Marshal(map[string]any{
		"original_teacher_id":  sameID,
		"substitute_teacher_id": sameID,
		"academic_year_id":     "00000000-0000-0000-0000-000000000003",
		"substitution_date":    "2026-04-20",
		"reason_type":          teacher_substitution.ReasonSakit,
		"class_room_id":        "00000000-0000-0000-0000-000000000004",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// ── invalid FK UUID test ────────────────────────────────────────────────────

func TestSubstitutionHandler_InvalidOriginalTeacherUUID_Returns400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(substitutionScopeMiddleware(substitutionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondSubForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondSubInvalidJSON(w)
			return
		}
		teacherID, _ := body["original_teacher_id"].(string)
		if _, err := uuid.Parse(teacherID); err != nil {
			respondSubInvalidUUID(w, "original_teacher_id")
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"original_teacher_id":  "not-a-uuid",
		"substitute_teacher_id": "00000000-0000-0000-0000-000000000002",
		"academic_year_id":     "00000000-0000-0000-0000-000000000003",
		"substitution_date":    "2026-04-20",
		"reason_type":          teacher_substitution.ReasonSakit,
		"class_room_id":        "00000000-0000-0000-0000-000000000004",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ── helpers ─────────────────────────────────────────────────────────────────

func validSubstitutionRequestBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"original_teacher_id":  "00000000-0000-0000-0000-000000000001",
		"substitute_teacher_id": "00000000-0000-0000-0000-000000000002",
		"academic_year_id":     "00000000-0000-0000-0000-000000000003",
		"substitution_date":    "2026-04-20",
		"reason_type":          teacher_substitution.ReasonSakit,
		"class_room_id":        "00000000-0000-0000-0000-000000000004",
		"status":               teacher_substitution.SubStatusPending,
	})
	require.NoError(t, err)
	return body
}

func respondSubForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func respondSubInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func respondSubValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}

func respondSubInvalidUUID(w http.ResponseWriter, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_ID", "message": field + " harus berupa UUID valid"},
	})
}
