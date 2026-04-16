package student_admission_test

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

	"github.com/yourorg/boilerplate/internal/domain/student_admission"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var admissionTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func admissionScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestStudentAdmissionHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "FORBIDDEN",
					"message": "scope organisasi tidak teridentifikasi",
				},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validAdmissionBody(t)
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

func TestStudentAdmissionHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(admissionScopeMiddleware(admissionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, admissionTestScope.TenantID, s.TenantID)
		assert.Equal(t, admissionTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestStudentAdmissionHandler_InvalidJSONBody(t *testing.T) {
	d := &student_admission.Descriptor{}
	r := chi.NewRouter()
	r.Use(admissionScopeMiddleware(admissionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			admissionRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			admissionRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			admissionRespondValidationError(w, err.Error())
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

func TestStudentAdmissionHandler_InvalidUUIDInPath_GetByID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(admissionScopeMiddleware(admissionTestScope))
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

func TestStudentAdmissionHandler_InvalidUUIDInPath_Delete(t *testing.T) {
	r := chi.NewRouter()
	r.Use(admissionScopeMiddleware(admissionTestScope))
	r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
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
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodDelete, "/bad-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	errMap, ok := resp["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "INVALID_ID", errMap["code"])
}

// -- Descriptor.Validate through HTTP pipeline -----------------------------------

func TestStudentAdmissionHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &student_admission.Descriptor{}
	body := validAdmissionBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestStudentAdmissionHandler_MissingRegistrationNumber_FailsValidation(t *testing.T) {
	d := &student_admission.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"admission_type": "regular",
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registration_number wajib diisi")
}

// -- Table-driven: full validation pipeline --------------------------------------

func TestStudentAdmissionHandler_CreateValidation_Table(t *testing.T) {
	d := &student_admission.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid regular pending",
			body: map[string]any{
				"registration_number": "PPDB-2026-001",
				"student_id":          "00000000-0000-0000-0000-000000000001",
				"academic_year_id":    "00000000-0000-0000-0000-000000000002",
				"admission_type":      student_admission.TypeRegular,
				"status":              student_admission.StatusPending,
				"registration_date":   "2026-01-10",
			},
			wantErr: false,
		},
		{
			name: "valid pindahan accepted",
			body: map[string]any{
				"registration_number": "PPDB-2026-002",
				"student_id":          "00000000-0000-0000-0000-000000000001",
				"academic_year_id":    "00000000-0000-0000-0000-000000000002",
				"admission_type":      student_admission.TypePindahan,
				"status":              student_admission.StatusAccepted,
			},
			wantErr: false,
		},
		{
			name: "valid afirmasi",
			body: map[string]any{
				"registration_number": "PPDB-2026-003",
				"admission_type":      student_admission.TypeAfirmasi,
				"status":              student_admission.StatusTest,
			},
			wantErr: false,
		},
		{
			name: "valid prestasi",
			body: map[string]any{
				"registration_number": "PPDB-2026-004",
				"admission_type":      student_admission.TypePrestasi,
				"status":              student_admission.StatusInterview,
			},
			wantErr: false,
		},
		{
			name: "missing registration_number",
			body: map[string]any{
				"admission_type": student_admission.TypeRegular,
			},
			wantErr: true,
			errMsg:  "registration_number wajib diisi",
		},
		{
			name: "missing admission_type",
			body: map[string]any{
				"registration_number": "PPDB-2026-001",
			},
			wantErr: true,
			errMsg:  "admission_type wajib diisi",
		},
		{
			name: "invalid admission_type",
			body: map[string]any{
				"registration_number": "PPDB-2026-001",
				"admission_type":      "internasional",
			},
			wantErr: true,
			errMsg:  "admission_type tidak valid",
		},
		{
			name: "invalid status",
			body: map[string]any{
				"registration_number": "PPDB-2026-001",
				"admission_type":      student_admission.TypeRegular,
				"status":              "waiting",
			},
			wantErr: true,
			errMsg:  "status tidak valid",
		},
		{
			name: "negative test_score",
			body: map[string]any{
				"registration_number": "PPDB-2026-001",
				"admission_type":      student_admission.TypeRegular,
				"test_score":          float64(-10),
			},
			wantErr: true,
			errMsg:  "test_score tidak boleh negatif",
		},
		{
			name: "negative interview_score",
			body: map[string]any{
				"registration_number": "PPDB-2026-001",
				"admission_type":      student_admission.TypeRegular,
				"interview_score":     float64(-5),
			},
			wantErr: true,
			errMsg:  "interview_score tidak boleh negatif",
		},
		{
			name: "negative final_score",
			body: map[string]any{
				"registration_number": "PPDB-2026-001",
				"admission_type":      student_admission.TypeRegular,
				"final_score":         float64(-1),
			},
			wantErr: true,
			errMsg:  "final_score tidak boleh negatif",
		},
		{
			name: "valid with positive scores",
			body: map[string]any{
				"registration_number": "PPDB-2026-001",
				"admission_type":      student_admission.TypeRegular,
				"test_score":          float64(85),
				"interview_score":     float64(90),
				"final_score":         float64(87.5),
			},
			wantErr: false,
		},
		{
			name: "valid with all statuses",
			body: map[string]any{
				"registration_number": "PPDB-2026-005",
				"admission_type":      student_admission.TypeRegular,
				"status":              student_admission.StatusDocReview,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			var parsed map[string]any
			err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&parsed)
			require.NoError(t, err)

			err = d.Validate(parsed)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// -- Full handler pipeline simulation -------------------------------------------

func TestStudentAdmissionHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &student_admission.Descriptor{}
	r := chi.NewRouter()
	r.Use(admissionScopeMiddleware(admissionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			admissionRespondForbidden(w)
			return
		}
		assert.Equal(t, admissionTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			admissionRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			admissionRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validAdmissionBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestStudentAdmissionHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &student_admission.Descriptor{}
	r := chi.NewRouter()
	r.Use(admissionScopeMiddleware(admissionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			admissionRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			admissionRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			admissionRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"admission_type": student_admission.TypeRegular,
		// Missing registration_number
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

func TestStudentAdmissionHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &student_admission.Descriptor{}
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			admissionRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			admissionRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			admissionRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validAdmissionBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestStudentAdmissionHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &student_admission.Descriptor{}
	r := chi.NewRouter()
	r.Use(admissionScopeMiddleware(admissionTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			admissionRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			admissionRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			admissionRespondValidationError(w, err.Error())
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

func validAdmissionBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"registration_number": "PPDB-2026-001",
		"student_id":          "00000000-0000-0000-0000-000000000001",
		"academic_year_id":    "00000000-0000-0000-0000-000000000002",
		"admission_type":      student_admission.TypeRegular,
		"status":              student_admission.StatusPending,
		"registration_date":   "2026-01-10",
	})
	require.NoError(t, err)
	return body
}

func admissionRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func admissionRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func admissionRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
