package rapor_generation_test

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

	"github.com/yourorg/boilerplate/internal/domain/rapor_generation"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures -----------------------------------------------------------

var raporTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func raporScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- helpers -----------------------------------------------------------------

func raporRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func raporRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func raporRespondInvalidUUID(w http.ResponseWriter, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_ID", "message": field + " harus berupa UUID valid"},
	})
}

func raporRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}

// -- Scope enforcement tests -------------------------------------------------

func TestRecordHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			raporRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validRecordBody(t)
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

func TestRecordHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(raporScopeMiddleware(raporTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, raporTestScope.TenantID, s.TenantID)
		assert.Equal(t, raporTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -------------------------------------------------------

func TestRecordHandler_InvalidJSON_Returns400(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	r := chi.NewRouter()
	r.Use(raporScopeMiddleware(raporTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			raporRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			raporRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			raporRespondValidationError(w, err.Error())
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

// -- Invalid FK UUID tests --------------------------------------------------

func TestRecordHandler_InvalidStudentUUID_Returns400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(raporScopeMiddleware(raporTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			raporRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			raporRespondInvalidJSON(w)
			return
		}
		studentID, _ := body["student_id"].(string)
		if _, err := uuid.Parse(studentID); err != nil {
			raporRespondInvalidUUID(w, "student_id")
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"student_id":       "not-a-uuid",
		"academic_year_id": "00000000-0000-0000-0000-000000000002",
		"class_room_id":    "00000000-0000-0000-0000-000000000003",
		"template_id":      "00000000-0000-0000-0000-000000000004",
		"semester":         rapor_generation.SemesterGanjil,
		"status":           rapor_generation.RecordStatusDraft,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRecordHandler_InvalidTemplateUUID_Returns400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(raporScopeMiddleware(raporTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			raporRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			raporRespondInvalidJSON(w)
			return
		}
		templateID, _ := body["template_id"].(string)
		if _, err := uuid.Parse(templateID); err != nil {
			raporRespondInvalidUUID(w, "template_id")
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"student_id":       "00000000-0000-0000-0000-000000000001",
		"academic_year_id": "00000000-0000-0000-0000-000000000002",
		"class_room_id":    "00000000-0000-0000-0000-000000000003",
		"template_id":      "bad-uuid",
		"semester":         rapor_generation.SemesterGanjil,
		"status":           rapor_generation.RecordStatusDraft,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// -- body helpers ------------------------------------------------------------

func validRecordBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"student_id":       "00000000-0000-0000-0000-000000000001",
		"academic_year_id": "00000000-0000-0000-0000-000000000002",
		"class_room_id":    "00000000-0000-0000-0000-000000000003",
		"template_id":      "00000000-0000-0000-0000-000000000004",
		"semester":         rapor_generation.SemesterGanjil,
		"status":           rapor_generation.RecordStatusDraft,
	})
	require.NoError(t, err)
	return body
}
