package rekening_test

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

	"github.com/yourorg/boilerplate/internal/domain/rekening"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── test fixtures ────────────────────────────────────────────────────────────

var rekeningTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

// scopeMiddleware injects the test scope into request context.
func rekeningScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ── Scope enforcement tests ──────────────────────────────────────────────────

func TestRekeningHandler_MissingScope_Returns403(t *testing.T) {
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

	body := validRekeningBody(t)
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

func TestRekeningHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(rekeningScopeMiddleware(rekeningTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, rekeningTestScope.TenantID, s.TenantID)
		assert.Equal(t, rekeningTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── Invalid JSON body ────────────────────────────────────────────────────────

func TestRekeningHandler_InvalidJSONBody(t *testing.T) {
	d := &rekening.Descriptor{}
	r := chi.NewRouter()
	r.Use(rekeningScopeMiddleware(rekeningTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			rekeningRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			rekeningRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			rekeningRespondValidationError(w, err.Error())
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

// ── Invalid UUID in path ─────────────────────────────────────────────────────

func TestRekeningHandler_InvalidUUIDInPath_GetByID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(rekeningScopeMiddleware(rekeningTestScope))
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

func TestRekeningHandler_InvalidUUIDInPath_Delete(t *testing.T) {
	r := chi.NewRouter()
	r.Use(rekeningScopeMiddleware(rekeningTestScope))
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

// ── Descriptor.Validate through HTTP pipeline ────────────────────────────────

func TestRekeningHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &rekening.Descriptor{}
	body := validRekeningBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestRekeningHandler_NegativeBalance_FailsValidation(t *testing.T) {
	d := &rekening.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"balance": -1000,
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "balance tidak boleh negatif")
}

// ── Table-driven: full validation pipeline ───────────────────────────────────

func TestRekeningHandler_CreateValidation_Table(t *testing.T) {
	d := &rekening.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid active account",
			body: map[string]any{
				"nasabah_id":     "00000000-0000-0000-0000-000000000001",
				"produk_akad_id": "00000000-0000-0000-0000-000000000002",
				"no_rekening":    "REK-001",
				"balance":        float64(500000),
				"hold_balance":   float64(0),
				"status":         rekening.StatusActive,
			},
			wantErr: false,
		},
		{
			name: "valid frozen account",
			body: map[string]any{
				"balance":  float64(100000),
				"status":   rekening.StatusFrozen,
			},
			wantErr: false,
		},
		{
			name: "valid closed account",
			body: map[string]any{
				"balance":  float64(0),
				"status":   rekening.StatusClosed,
			},
			wantErr: false,
		},
		{
			name: "valid no status",
			body: map[string]any{
				"balance":      float64(250000),
				"hold_balance": float64(50000),
			},
			wantErr: false,
		},
		{
			name: "negative balance",
			body: map[string]any{
				"balance": float64(-1),
			},
			wantErr: true,
			errMsg:  "balance tidak boleh negatif",
		},
		{
			name: "hold exceeds balance",
			body: map[string]any{
				"balance":      float64(100),
				"hold_balance": float64(200),
			},
			wantErr: true,
			errMsg:  "hold_balance tidak boleh lebih besar dari balance",
		},
		{
			name: "invalid status",
			body: map[string]any{
				"balance": float64(100),
				"status":  "pending",
			},
			wantErr: true,
			errMsg:  "status tidak valid",
		},
		{
			name: "zero balance zero hold",
			body: map[string]any{
				"balance":      float64(0),
				"hold_balance": float64(0),
			},
			wantErr: false,
		},
		{
			name: "hold equals balance",
			body: map[string]any{
				"balance":      float64(50000),
				"hold_balance": float64(50000),
			},
			wantErr: false,
		},
		{
			name: "large balance",
			body: map[string]any{
				"balance": float64(999999999999),
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

func TestRekeningHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &rekening.Descriptor{}
	r := chi.NewRouter()
	r.Use(rekeningScopeMiddleware(rekeningTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			rekeningRespondForbidden(w)
			return
		}
		assert.Equal(t, rekeningTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			rekeningRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			rekeningRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validRekeningBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestRekeningHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &rekening.Descriptor{}
	r := chi.NewRouter()
	r.Use(rekeningScopeMiddleware(rekeningTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			rekeningRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			rekeningRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			rekeningRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"balance": float64(-500),
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

func TestRekeningHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &rekening.Descriptor{}
	r := chi.NewRouter()
	// No scope middleware
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			rekeningRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			rekeningRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			rekeningRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validRekeningBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRekeningHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &rekening.Descriptor{}
	r := chi.NewRouter()
	r.Use(rekeningScopeMiddleware(rekeningTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			rekeningRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			rekeningRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			rekeningRespondValidationError(w, err.Error())
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

// ── helpers ──────────────────────────────────────────────────────────────────

func validRekeningBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"nasabah_id":     "00000000-0000-0000-0000-000000000001",
		"produk_akad_id": "00000000-0000-0000-0000-000000000002",
		"no_rekening":    "REK-001",
		"balance":        float64(500000),
		"hold_balance":   float64(0),
		"status":         rekening.StatusActive,
	})
	require.NoError(t, err)
	return body
}

func rekeningRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func rekeningRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func rekeningRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
