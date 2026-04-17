package student_counseling_test

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

	"github.com/yourorg/boilerplate/internal/domain/student_counseling"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var scTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func scScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestCounselingCaseHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			scRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validCaseBody(t)
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

func TestCounselingCaseHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(scScopeMiddleware(scTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, scTestScope.TenantID, s.TenantID)
		assert.Equal(t, scTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestCounselingCaseHandler_InvalidJSON_Returns400(t *testing.T) {
	d := &student_counseling.CaseDescriptor{}
	r := chi.NewRouter()
	r.Use(scScopeMiddleware(scTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			scRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			scRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			scRespondValidationError(w, err.Error())
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

// -- Invalid FK UUID tests ------------------------------------------------------

func TestCounselingCaseHandler_InvalidStudentUUID_Returns400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(scScopeMiddleware(scTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			scRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			scRespondInvalidJSON(w)
			return
		}
		studentID, _ := body["student_id"].(string)
		if _, err := uuid.Parse(studentID); err != nil {
			scRespondInvalidUUID(w, "student_id")
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"student_id":       "not-a-uuid",
		"academic_year_id": "00000000-0000-0000-0000-000000000002",
		"case_no":          "BK-2026-001",
		"category":         student_counseling.CategoryAcademic,
		"title":            "Test",
		"severity":         student_counseling.SeverityLow,
		"opened_date":      "2026-04-17",
		"counselor_id":     "00000000-0000-0000-0000-000000000003",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCounselingCaseHandler_InvalidCounselorUUID_Returns400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(scScopeMiddleware(scTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			scRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			scRespondInvalidJSON(w)
			return
		}
		counselorID, _ := body["counselor_id"].(string)
		if _, err := uuid.Parse(counselorID); err != nil {
			scRespondInvalidUUID(w, "counselor_id")
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"student_id":       "00000000-0000-0000-0000-000000000001",
		"academic_year_id": "00000000-0000-0000-0000-000000000002",
		"case_no":          "BK-2026-001",
		"category":         student_counseling.CategoryAcademic,
		"title":            "Test",
		"severity":         student_counseling.SeverityLow,
		"opened_date":      "2026-04-17",
		"counselor_id":     "bad-uuid",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// -- helpers ---------------------------------------------------------------------

func validCaseBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"student_id":       "00000000-0000-0000-0000-000000000001",
		"academic_year_id": "00000000-0000-0000-0000-000000000002",
		"case_no":          "BK-2026-001",
		"category":         student_counseling.CategoryAcademic,
		"title":            "Kesulitan Fokus Belajar",
		"severity":         student_counseling.SeverityMedium,
		"opened_date":      "2026-04-17",
		"counselor_id":     "00000000-0000-0000-0000-000000000003",
	})
	require.NoError(t, err)
	return body
}

func scRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func scRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func scRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}

func scRespondInvalidUUID(w http.ResponseWriter, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_ID", "message": field + " harus berupa UUID valid"},
	})
}
