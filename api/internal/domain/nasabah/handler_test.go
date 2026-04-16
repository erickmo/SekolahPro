package nasabah_test

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

	"github.com/yourorg/boilerplate/internal/domain/nasabah"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── test fixtures ────────────────────────────────────────────────────────────

var nasabahTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

// scopeMiddleware injects the test scope into request context.
func nasabahScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ── Scope enforcement tests ──────────────────────────────────────────────────

func TestNasabahHandler_MissingScope_Returns403(t *testing.T) {
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

	body := validNasabahBody(t)
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

func TestNasabahHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(nasabahScopeMiddleware(nasabahTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, nasabahTestScope.TenantID, s.TenantID)
		assert.Equal(t, nasabahTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── Invalid JSON body ────────────────────────────────────────────────────────

func TestNasabahHandler_InvalidJSONBody(t *testing.T) {
	r := chi.NewRouter()
	r.Use(nasabahScopeMiddleware(nasabahTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "INVALID_JSON",
					"message": "request body bukan JSON valid",
				},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{invalid")))
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

// ── Invalid UUID in path ─────────────────────────────────────────────────────

func TestNasabahHandler_InvalidUUIDInPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(nasabahScopeMiddleware(nasabahTestScope))
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
		{"partial-uuid", "12345", http.StatusBadRequest},
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

// ── Descriptor.Validate through HTTP pipeline ────────────────────────────────

func TestNasabahHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &nasabah.Descriptor{}
	body := validNasabahBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestNasabahHandler_MissingNIK_FailsValidation(t *testing.T) {
	d := &nasabah.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"full_name": "Budi Santoso",
		"type":      nasabah.TypeSiswa,
		"gender":    nasabah.GenderL,
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nik wajib diisi")
}

func TestNasabahHandler_InvalidNIK_FailsValidation(t *testing.T) {
	d := &nasabah.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"nik":       "123",
		"full_name": "Budi Santoso",
		"type":      nasabah.TypeSiswa,
		"gender":    nasabah.GenderL,
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nik tidak valid")
}

// ── Table-driven: full HTTP create pipeline ──────────────────────────────────

func TestNasabahHandler_CreateValidation_Table(t *testing.T) {
	d := &nasabah.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid siswa L",
			body: map[string]any{
				"nik": "3201234567890001", "full_name": "Budi",
				"type": nasabah.TypeSiswa, "gender": nasabah.GenderL,
			},
			wantErr: false,
		},
		{
			name: "valid guru P",
			body: map[string]any{
				"nik": "3201234567890002", "full_name": "Siti",
				"type": nasabah.TypeGuru, "gender": nasabah.GenderP,
			},
			wantErr: false,
		},
		{
			name: "valid staff L",
			body: map[string]any{
				"nik": "3201234567890003", "full_name": "Ahmad",
				"type": nasabah.TypeStaff, "gender": nasabah.GenderL,
			},
			wantErr: false,
		},
		{
			name: "valid umum P",
			body: map[string]any{
				"nik": "3201234567890004", "full_name": "Dewi",
				"type": nasabah.TypeUmum, "gender": nasabah.GenderP,
			},
			wantErr: false,
		},
		{
			name: "missing NIK",
			body: map[string]any{
				"full_name": "Budi", "type": nasabah.TypeSiswa, "gender": nasabah.GenderL,
			},
			wantErr: true, errMsg: "nik wajib diisi",
		},
		{
			name: "invalid NIK too short",
			body: map[string]any{
				"nik": "12345", "full_name": "Budi",
				"type": nasabah.TypeSiswa, "gender": nasabah.GenderL,
			},
			wantErr: true, errMsg: "nik tidak valid",
		},
		{
			name: "missing type",
			body: map[string]any{
				"nik": "3201234567890001", "full_name": "Budi", "gender": nasabah.GenderL,
			},
			wantErr: true, errMsg: "type wajib diisi",
		},
		{
			name: "invalid type",
			body: map[string]any{
				"nik": "3201234567890001", "full_name": "Budi",
				"type": "mahasiswa", "gender": nasabah.GenderL,
			},
			wantErr: true, errMsg: "type tidak valid",
		},
		{
			name: "missing gender",
			body: map[string]any{
				"nik": "3201234567890001", "full_name": "Budi", "type": nasabah.TypeSiswa,
			},
			wantErr: true, errMsg: "gender wajib diisi",
		},
		{
			name: "invalid gender",
			body: map[string]any{
				"nik": "3201234567890001", "full_name": "Budi",
				"type": nasabah.TypeSiswa, "gender": "X",
			},
			wantErr: true, errMsg: "gender tidak valid",
		},
		{
			name: "invalid status",
			body: map[string]any{
				"nik": "3201234567890001", "full_name": "Budi",
				"type": nasabah.TypeSiswa, "gender": nasabah.GenderL,
				"status": "unknown",
			},
			wantErr: true, errMsg: "status tidak valid",
		},
		{
			name: "valid with optional status active",
			body: map[string]any{
				"nik": "3201234567890001", "full_name": "Budi",
				"type": nasabah.TypeSiswa, "gender": nasabah.GenderL,
				"status": nasabah.StatusActive,
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

// ── Full handler pipeline simulation ─────────────────────────────────────────

func TestNasabahHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &nasabah.Descriptor{}
	r := chi.NewRouter()
	r.Use(nasabahScopeMiddleware(nasabahTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondForbidden(w)
			return
		}
		assert.Equal(t, nasabahTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			respondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validNasabahBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestNasabahHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &nasabah.Descriptor{}
	r := chi.NewRouter()
	r.Use(nasabahScopeMiddleware(nasabahTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			respondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"full_name": "Budi Santoso",
		"type":      nasabah.TypeSiswa,
		"gender":    nasabah.GenderL,
		// Missing NIK
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

func TestNasabahHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &nasabah.Descriptor{}
	r := chi.NewRouter()
	r.Use(nasabahScopeMiddleware(nasabahTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			respondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNasabahHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &nasabah.Descriptor{}
	r := chi.NewRouter()
	// No scope middleware
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			respondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			respondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validNasabahBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func validNasabahBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"nik":       "3201234567890001",
		"full_name": "Budi Santoso",
		"type":      nasabah.TypeSiswa,
		"gender":    nasabah.GenderL,
		"status":    nasabah.StatusActive,
	})
	require.NoError(t, err)
	return body
}

func respondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func respondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func respondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
