package student_test

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

	"github.com/yourorg/boilerplate/internal/domain/student"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var studentTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

// studentScopeMiddleware injects the test scope into request context.
func studentScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestStudentHandler_MissingScope_Returns403(t *testing.T) {
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

	body := validStudentBody(t)
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

func TestStudentHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(studentScopeMiddleware(studentTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, studentTestScope.TenantID, s.TenantID)
		assert.Equal(t, studentTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestStudentHandler_InvalidJSONBody(t *testing.T) {
	d := &student.Descriptor{}
	r := chi.NewRouter()
	r.Use(studentScopeMiddleware(studentTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			studentRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			studentRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			studentRespondValidationError(w, err.Error())
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

func TestStudentHandler_InvalidUUIDInPath_GetByID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(studentScopeMiddleware(studentTestScope))
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

func TestStudentHandler_InvalidUUIDInPath_Delete(t *testing.T) {
	r := chi.NewRouter()
	r.Use(studentScopeMiddleware(studentTestScope))
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

func TestStudentHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &student.Descriptor{}
	body := validStudentBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestStudentHandler_InvalidGender_FailsValidation(t *testing.T) {
	d := &student.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"full_name": "Budi Santoso",
		"gender":    "X",
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gender tidak valid")
}

// -- Table-driven: full validation pipeline --------------------------------------

func TestStudentHandler_CreateValidation_Table(t *testing.T) {
	d := &student.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid active student L",
			body: map[string]any{
				"nis": "S001", "full_name": "Budi Santoso",
				"gender": student.GenderL, "birth_place": "Jakarta",
				"birth_date": "2010-01-15", "religion": "islam",
				"status": student.StatusActive,
			},
			wantErr: false,
		},
		{
			name: "valid active student P",
			body: map[string]any{
				"nis": "S002", "full_name": "Siti Aminah",
				"gender": student.GenderP, "status": student.StatusActive,
			},
			wantErr: false,
		},
		{
			name: "valid graduated",
			body: map[string]any{
				"full_name": "Ahmad", "gender": student.GenderL,
				"status": student.StatusGraduated,
			},
			wantErr: false,
		},
		{
			name: "valid transferred",
			body: map[string]any{
				"full_name": "Dewi", "gender": student.GenderP,
				"status": student.StatusTransferred,
			},
			wantErr: false,
		},
		{
			name: "valid inactive",
			body: map[string]any{
				"full_name": "Roni", "gender": student.GenderL,
				"status": student.StatusInactive,
			},
			wantErr: false,
		},
		{
			name: "valid dropped_out",
			body: map[string]any{
				"full_name": "Doni", "gender": student.GenderL,
				"status": student.StatusDroppedOut,
			},
			wantErr: false,
		},
		{
			name: "missing full_name",
			body: map[string]any{
				"gender": student.GenderL,
			},
			wantErr: true,
			errMsg:  "full_name wajib diisi",
		},
		{
			name: "missing gender",
			body: map[string]any{
				"full_name": "Budi",
			},
			wantErr: true,
			errMsg:  "gender wajib diisi",
		},
		{
			name: "invalid gender",
			body: map[string]any{
				"full_name": "Budi", "gender": "X",
			},
			wantErr: true,
			errMsg:  "gender tidak valid",
		},
		{
			name: "invalid status",
			body: map[string]any{
				"full_name": "Budi", "gender": student.GenderL,
				"status": "suspended",
			},
			wantErr: true,
			errMsg:  "status tidak valid",
		},
		{
			name: "invalid blood_type",
			body: map[string]any{
				"full_name": "Budi", "gender": student.GenderL,
				"blood_type": "Z",
			},
			wantErr: true,
			errMsg:  "blood_type tidak valid",
		},
		{
			name: "invalid religion",
			body: map[string]any{
				"full_name": "Budi", "gender": student.GenderL,
				"religion": "atheism",
			},
			wantErr: true,
			errMsg:  "religion tidak valid",
		},
		{
			name: "valid with blood type O",
			body: map[string]any{
				"full_name": "Budi", "gender": student.GenderL,
				"blood_type": "O",
			},
			wantErr: false,
		},
		{
			name: "valid with religion kristen",
			body: map[string]any{
				"full_name": "Budi", "gender": student.GenderL,
				"religion": "kristen",
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

func TestStudentHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &student.Descriptor{}
	r := chi.NewRouter()
	r.Use(studentScopeMiddleware(studentTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			studentRespondForbidden(w)
			return
		}
		assert.Equal(t, studentTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			studentRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			studentRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validStudentBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestStudentHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &student.Descriptor{}
	r := chi.NewRouter()
	r.Use(studentScopeMiddleware(studentTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			studentRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			studentRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			studentRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"gender": student.GenderL,
		// Missing full_name
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

func TestStudentHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &student.Descriptor{}
	r := chi.NewRouter()
	// No scope middleware
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			studentRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			studentRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			studentRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validStudentBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestStudentHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &student.Descriptor{}
	r := chi.NewRouter()
	r.Use(studentScopeMiddleware(studentTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			studentRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			studentRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			studentRespondValidationError(w, err.Error())
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

func validStudentBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"nis":         "S001",
		"full_name":   "Budi Santoso",
		"gender":      student.GenderL,
		"birth_place": "Jakarta",
		"birth_date":  "2010-01-15",
		"religion":    "islam",
		"status":      student.StatusActive,
	})
	require.NoError(t, err)
	return body
}

func studentRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func studentRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func studentRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
