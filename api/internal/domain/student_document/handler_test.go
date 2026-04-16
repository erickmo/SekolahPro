package student_document_test

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

	"github.com/yourorg/boilerplate/internal/domain/student_document"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var docTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func docScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestStudentDocumentHandler_MissingScope_Returns403(t *testing.T) {
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

	body := validDocumentBody(t)
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

func TestStudentDocumentHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(docScopeMiddleware(docTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, docTestScope.TenantID, s.TenantID)
		assert.Equal(t, docTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestStudentDocumentHandler_InvalidJSONBody(t *testing.T) {
	d := &student_document.Descriptor{}
	r := chi.NewRouter()
	r.Use(docScopeMiddleware(docTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			docRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			docRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			docRespondValidationError(w, err.Error())
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

func TestStudentDocumentHandler_InvalidUUIDInPath_GetByID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(docScopeMiddleware(docTestScope))
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

func TestStudentDocumentHandler_InvalidUUIDInPath_Delete(t *testing.T) {
	r := chi.NewRouter()
	r.Use(docScopeMiddleware(docTestScope))
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

func TestStudentDocumentHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &student_document.Descriptor{}
	body := validDocumentBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestStudentDocumentHandler_MissingStudentID_FailsValidation(t *testing.T) {
	d := &student_document.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"document_type": "akta_kelahiran",
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "student_id wajib diisi")
}

// -- Table-driven: full validation pipeline --------------------------------------

func TestStudentDocumentHandler_CreateValidation_Table(t *testing.T) {
	d := &student_document.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid akta_kelahiran",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  student_document.TypeAkta,
				"title":          "Akta Kelahiran Budi",
				"file_name":      "akta.pdf",
				"mime_type":      "application/pdf",
			},
			wantErr: false,
		},
		{
			name: "valid kartu_keluarga",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  student_document.TypeKK,
			},
			wantErr: false,
		},
		{
			name: "valid ijazah",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  student_document.TypeIjazah,
			},
			wantErr: false,
		},
		{
			name: "valid skhun",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  student_document.TypeSKHUN,
			},
			wantErr: false,
		},
		{
			name: "valid pass_photo",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  student_document.TypePhoto,
			},
			wantErr: false,
		},
		{
			name: "valid ktp_ortu",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  student_document.TypeKTPOrtu,
			},
			wantErr: false,
		},
		{
			name: "missing student_id",
			body: map[string]any{
				"document_type": student_document.TypeAkta,
			},
			wantErr: true,
			errMsg:  "student_id wajib diisi",
		},
		{
			name: "missing document_type",
			body: map[string]any{
				"student_id": "00000000-0000-0000-0000-000000000001",
			},
			wantErr: true,
			errMsg:  "document_type wajib diisi",
		},
		{
			name: "invalid document_type",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  "ktp_siswa",
			},
			wantErr: true,
			errMsg:  "document_type tidak valid",
		},
		{
			name: "invalid status",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  student_document.TypeAkta,
				"status":         "approved",
			},
			wantErr: true,
			errMsg:  "status tidak valid",
		},
		{
			name: "valid with status verified",
			body: map[string]any{
				"student_id":     "00000000-0000-0000-0000-000000000001",
				"document_type":  student_document.TypeAkta,
				"status":         student_document.StatusVerified,
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

func TestStudentDocumentHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &student_document.Descriptor{}
	r := chi.NewRouter()
	r.Use(docScopeMiddleware(docTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			docRespondForbidden(w)
			return
		}
		assert.Equal(t, docTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			docRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			docRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validDocumentBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestStudentDocumentHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &student_document.Descriptor{}
	r := chi.NewRouter()
	r.Use(docScopeMiddleware(docTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			docRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			docRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			docRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"document_type": student_document.TypeAkta,
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

func TestStudentDocumentHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &student_document.Descriptor{}
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			docRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			docRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			docRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validDocumentBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestStudentDocumentHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &student_document.Descriptor{}
	r := chi.NewRouter()
	r.Use(docScopeMiddleware(docTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			docRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			docRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			docRespondValidationError(w, err.Error())
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

func validDocumentBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"student_id":     "00000000-0000-0000-0000-000000000001",
		"document_type":  student_document.TypeAkta,
		"title":          "Akta Kelahiran Budi",
		"file_name":      "akta.pdf",
		"mime_type":      "application/pdf",
	})
	require.NoError(t, err)
	return body
}

func docRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func docRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func docRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
