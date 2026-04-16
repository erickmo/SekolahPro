package student_guardian_test

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

	"github.com/yourorg/boilerplate/internal/domain/student_guardian"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var guardianTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func guardianScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestStudentGuardianHandler_MissingScope_Returns403(t *testing.T) {
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

	body := validGuardianBody(t)
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

func TestStudentGuardianHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(guardianScopeMiddleware(guardianTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, guardianTestScope.TenantID, s.TenantID)
		assert.Equal(t, guardianTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestStudentGuardianHandler_InvalidJSONBody(t *testing.T) {
	d := &student_guardian.Descriptor{}
	r := chi.NewRouter()
	r.Use(guardianScopeMiddleware(guardianTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			guardianRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			guardianRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			guardianRespondValidationError(w, err.Error())
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

func TestStudentGuardianHandler_InvalidUUIDInPath_GetByID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(guardianScopeMiddleware(guardianTestScope))
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

func TestStudentGuardianHandler_InvalidUUIDInPath_Delete(t *testing.T) {
	r := chi.NewRouter()
	r.Use(guardianScopeMiddleware(guardianTestScope))
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

func TestStudentGuardianHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &student_guardian.Descriptor{}
	body := validGuardianBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestStudentGuardianHandler_MissingStudentID_FailsValidation(t *testing.T) {
	d := &student_guardian.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"guardian_type": "ayah",
		"full_name":     "Ahmad Santoso",
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "student_id wajib diisi")
}

// -- Table-driven: full validation pipeline --------------------------------------

func TestStudentGuardianHandler_CreateValidation_Table(t *testing.T) {
	d := &student_guardian.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ayah",
			body: map[string]any{
				"student_id":    "00000000-0000-0000-0000-000000000001",
				"guardian_type": student_guardian.TypeAyah,
				"full_name":     "Ahmad Santoso",
				"nik":           "3201234567890001",
				"phone":         "081234567890",
			},
			wantErr: false,
		},
		{
			name: "valid ibu",
			body: map[string]any{
				"student_id":    "00000000-0000-0000-0000-000000000001",
				"guardian_type": student_guardian.TypeIbu,
				"full_name":     "Siti Aminah",
			},
			wantErr: false,
		},
		{
			name: "valid wali",
			body: map[string]any{
				"student_id":    "00000000-0000-0000-0000-000000000001",
				"guardian_type": student_guardian.TypeWali,
				"full_name":     "Hasan Basri",
			},
			wantErr: false,
		},
		{
			name: "missing student_id",
			body: map[string]any{
				"guardian_type": student_guardian.TypeAyah,
				"full_name":     "Ahmad",
			},
			wantErr: true,
			errMsg:  "student_id wajib diisi",
		},
		{
			name: "missing guardian_type",
			body: map[string]any{
				"student_id": "00000000-0000-0000-0000-000000000001",
				"full_name":  "Ahmad",
			},
			wantErr: true,
			errMsg:  "guardian_type wajib diisi",
		},
		{
			name: "invalid guardian_type",
			body: map[string]any{
				"student_id":    "00000000-0000-0000-0000-000000000001",
				"guardian_type": "kakek",
				"full_name":     "Ahmad",
			},
			wantErr: true,
			errMsg:  "guardian_type tidak valid",
		},
		{
			name: "missing full_name",
			body: map[string]any{
				"student_id":    "00000000-0000-0000-0000-000000000001",
				"guardian_type": student_guardian.TypeAyah,
			},
			wantErr: true,
			errMsg:  "full_name wajib diisi",
		},
		{
			name: "valid with all optional fields",
			body: map[string]any{
				"student_id":       "00000000-0000-0000-0000-000000000001",
				"guardian_type":    student_guardian.TypeAyah,
				"full_name":        "Ahmad Santoso",
				"nik":              "3201234567890001",
				"phone":            "081234567890",
				"email":            "ahmad@example.com",
				"address":          "Jl. Merdeka 1",
				"is_primary":       true,
				"education_level":  "S1",
				"occupation":       "PNS",
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

func TestStudentGuardianHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &student_guardian.Descriptor{}
	r := chi.NewRouter()
	r.Use(guardianScopeMiddleware(guardianTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			guardianRespondForbidden(w)
			return
		}
		assert.Equal(t, guardianTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			guardianRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			guardianRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validGuardianBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestStudentGuardianHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &student_guardian.Descriptor{}
	r := chi.NewRouter()
	r.Use(guardianScopeMiddleware(guardianTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			guardianRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			guardianRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			guardianRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"guardian_type": student_guardian.TypeAyah,
		"full_name":     "Ahmad Santoso",
		// Missing student_id
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

func TestStudentGuardianHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &student_guardian.Descriptor{}
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			guardianRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			guardianRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			guardianRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validGuardianBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestStudentGuardianHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &student_guardian.Descriptor{}
	r := chi.NewRouter()
	r.Use(guardianScopeMiddleware(guardianTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			guardianRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			guardianRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			guardianRespondValidationError(w, err.Error())
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

func validGuardianBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"student_id":    "00000000-0000-0000-0000-000000000001",
		"guardian_type": student_guardian.TypeAyah,
		"full_name":     "Ahmad Santoso",
		"nik":           "3201234567890001",
		"phone":         "081234567890",
	})
	require.NoError(t, err)
	return body
}

func guardianRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func guardianRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func guardianRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
