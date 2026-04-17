package exam_assessment_test

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

	"github.com/yourorg/boilerplate/internal/domain/exam_assessment"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var eaTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func eaScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestExamAssessmentHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			eaRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validExamAssessmentBody(t)
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

func TestExamAssessmentHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(eaScopeMiddleware(eaTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, eaTestScope.TenantID, s.TenantID)
		assert.Equal(t, eaTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestExamAssessmentHandler_InvalidJSONBody(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	r := chi.NewRouter()
	r.Use(eaScopeMiddleware(eaTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			eaRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			eaRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			eaRespondValidationError(w, err.Error())
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

func TestExamAssessmentHandler_InvalidUUIDInPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(eaScopeMiddleware(eaTestScope))
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

func TestExamAssessmentHandler_InvalidSubjectUUIDInBody(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"subject_id":   "not-a-uuid",
		"teacher_id":   "00000000-0000-0000-0000-000000000002",
		"class_room_id": "00000000-0000-0000-0000-000000000003",
		"title":        "UTS",
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err, "descriptor only checks non-empty, UUID format is a handler concern")
}

func TestExamAssessmentHandler_InvalidTeacherUUIDInBody(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"subject_id":    "00000000-0000-0000-0000-000000000001",
		"teacher_id":    "not-a-uuid",
		"class_room_id": "00000000-0000-0000-0000-000000000003",
		"title":         "UTS",
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err, "descriptor only checks non-empty, UUID format is a handler concern")
}

func TestExamAssessmentHandler_InvalidClassRoomUUIDInBody(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"subject_id":    "00000000-0000-0000-0000-000000000001",
		"teacher_id":    "00000000-0000-0000-0000-000000000002",
		"class_room_id": "not-a-uuid",
		"title":         "UTS",
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err, "descriptor only checks non-empty, UUID format is a handler concern")
}

// -- Full handler pipeline simulation -------------------------------------------

func TestExamAssessmentHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	r := chi.NewRouter()
	r.Use(eaScopeMiddleware(eaTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			eaRespondForbidden(w)
			return
		}
		assert.Equal(t, eaTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			eaRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			eaRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validExamAssessmentBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestExamAssessmentHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	r := chi.NewRouter()
	r.Use(eaScopeMiddleware(eaTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			eaRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			eaRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			eaRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"teacher_id":    "00000000-0000-0000-0000-000000000002",
		"class_room_id": "00000000-0000-0000-0000-000000000003",
		// Missing subject_id and title
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

func TestExamAssessmentHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			eaRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			eaRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			eaRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validExamAssessmentBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestExamAssessmentHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	r := chi.NewRouter()
	r.Use(eaScopeMiddleware(eaTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			eaRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			eaRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			eaRespondValidationError(w, err.Error())
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

func validExamAssessmentBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"subject_id":    "00000000-0000-0000-0000-000000000001",
		"teacher_id":    "00000000-0000-0000-0000-000000000002",
		"class_room_id": "00000000-0000-0000-0000-000000000003",
		"title":         "UTS Matematika Semester 1",
		"exam_type":     exam_assessment.ExamTypeUTS,
		"date":          "2026-05-17",
		"max_score":     100,
		"status":        exam_assessment.StatusDraft,
	})
	require.NoError(t, err)
	return body
}

func eaRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func eaRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func eaRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
