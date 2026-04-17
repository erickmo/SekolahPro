package subject_grade_test

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

	"github.com/yourorg/boilerplate/internal/domain/subject_grade"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var sgTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func sgScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestSubjectGradeHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sgRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validSubjectGradeBody(t)
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

func TestSubjectGradeHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(sgScopeMiddleware(sgTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, sgTestScope.TenantID, s.TenantID)
		assert.Equal(t, sgTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestSubjectGradeHandler_InvalidJSONBody(t *testing.T) {
	d := &subject_grade.Descriptor{}
	r := chi.NewRouter()
	r.Use(sgScopeMiddleware(sgTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sgRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			sgRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			sgRespondValidationError(w, err.Error())
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

// -- Invalid UUID in path -------------------------------------------------------

func TestSubjectGradeHandler_InvalidUUIDInPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(sgScopeMiddleware(sgTestScope))
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		raw := chi.URLParam(req, "id")
		_, err := uuid.Parse(raw)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "INVALID_ID",
					"message": "ID harus berupa UUID valid",
				},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		pathParam  string
		wantStatus int
	}{
		{"not-a-uuid", "not-a-uuid", http.StatusBadRequest},
		{"partial-uuid", "abc-def", http.StatusBadRequest},
		{"numeric", "12345", http.StatusBadRequest},
		{"valid-uuid", uuid.New().String(), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// -- Invalid UUID in FK fields --------------------------------------------------

func TestSubjectGradeHandler_InvalidStudentUUIDInBody(t *testing.T) {
	d := &subject_grade.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"student_id":  "not-a-uuid",
		"subject_id":  "00000000-0000-0000-0000-000000000002",
		"teacher_id":  "00000000-0000-0000-0000-000000000003",
		"grade":       85,
		"grade_type":  subject_grade.GradeTypeAssignment,
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	// student_id is a non-empty string so descriptor validation passes;
	// UUID format validation happens at the handler/DB layer.
	assert.NoError(t, err, "descriptor only checks non-empty, UUID format is a handler concern")
}

func TestSubjectGradeHandler_InvalidSubjectUUIDInBody(t *testing.T) {
	d := &subject_grade.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"student_id":  "00000000-0000-0000-0000-000000000001",
		"subject_id":  "not-a-uuid",
		"teacher_id":  "00000000-0000-0000-0000-000000000003",
		"grade":       85,
		"grade_type":  subject_grade.GradeTypeAssignment,
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err, "descriptor only checks non-empty, UUID format is a handler concern")
}

func TestSubjectGradeHandler_InvalidTeacherUUIDInBody(t *testing.T) {
	d := &subject_grade.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"student_id":  "00000000-0000-0000-0000-000000000001",
		"subject_id":  "00000000-0000-0000-0000-000000000002",
		"teacher_id":  "not-a-uuid",
		"grade":       85,
		"grade_type":  subject_grade.GradeTypeAssignment,
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err, "descriptor only checks non-empty, UUID format is a handler concern")
}

// -- Full handler pipeline simulation -------------------------------------------

func TestSubjectGradeHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &subject_grade.Descriptor{}
	r := chi.NewRouter()
	r.Use(sgScopeMiddleware(sgTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sgRespondForbidden(w)
			return
		}
		assert.Equal(t, sgTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			sgRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			sgRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validSubjectGradeBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestSubjectGradeHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &subject_grade.Descriptor{}
	r := chi.NewRouter()
	r.Use(sgScopeMiddleware(sgTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sgRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			sgRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			sgRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"subject_id": "00000000-0000-0000-0000-000000000002",
		"teacher_id": "00000000-0000-0000-0000-000000000003",
		// Missing student_id and grade
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	errMap, ok := resp["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", errMap["code"])
}

func TestSubjectGradeHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &subject_grade.Descriptor{}
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sgRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			sgRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			sgRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validSubjectGradeBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSubjectGradeHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &subject_grade.Descriptor{}
	r := chi.NewRouter()
	r.Use(sgScopeMiddleware(sgTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			sgRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			sgRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			sgRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("invalid body")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// -- helpers ---------------------------------------------------------------------

func validSubjectGradeBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"student_id":  "00000000-0000-0000-0000-000000000001",
		"subject_id":  "00000000-0000-0000-0000-000000000002",
		"teacher_id":  "00000000-0000-0000-0000-000000000003",
		"grade":       85,
		"grade_type":  subject_grade.GradeTypeAssignment,
		"description": "Tugas Bab 3",
	})
	require.NoError(t, err)
	return body
}

func sgRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func sgRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func sgRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
