package simpanan_pokok_wajib_test

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

	"github.com/yourorg/boilerplate/internal/domain/simpanan_pokok_wajib"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var spwTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func spwScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestSimpananPokokWajibHandler_MissingScope_Returns403(t *testing.T) {
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

	body := validSPWBody(t)
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

func TestSimpananPokokWajibHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(spwScopeMiddleware(spwTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, spwTestScope.TenantID, s.TenantID)
		assert.Equal(t, spwTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestSimpananPokokWajibHandler_InvalidJSONBody(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	r := chi.NewRouter()
	r.Use(spwScopeMiddleware(spwTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			spwRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			spwRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			spwRespondValidationError(w, err.Error())
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

func TestSimpananPokokWajibHandler_InvalidUUIDInPath_GetByID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(spwScopeMiddleware(spwTestScope))
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

func TestSimpananPokokWajibHandler_InvalidUUIDInPath_Delete(t *testing.T) {
	r := chi.NewRouter()
	r.Use(spwScopeMiddleware(spwTestScope))
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

func TestSimpananPokokWajibHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	body := validSPWBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestSimpananPokokWajibHandler_MissingNasabahID_FailsValidation(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"rekening_id": "00000000-0000-0000-0000-000000000002",
		"type":        simpanan_pokok_wajib.TypePokok,
		"amount":      float64(500000),
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nasabah_id wajib diisi")
}

// -- Table-driven: full validation pipeline --------------------------------------

func TestSimpananPokokWajibHandler_CreateValidation_Table(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pokok pending",
			body: map[string]any{
				"nasabah_id":   "00000000-0000-0000-0000-000000000001",
				"rekening_id":  "00000000-0000-0000-0000-000000000002",
				"type":         simpanan_pokok_wajib.TypePokok,
				"amount":       float64(500000),
				"status":       simpanan_pokok_wajib.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "valid wajib with period",
			body: map[string]any{
				"nasabah_id":   "00000000-0000-0000-0000-000000000001",
				"rekening_id":  "00000000-0000-0000-0000-000000000002",
				"type":         simpanan_pokok_wajib.TypeWajib,
				"amount":       float64(100000),
				"period":       "2026-01",
				"status":       simpanan_pokok_wajib.StatusPaid,
			},
			wantErr: false,
		},
		{
			name: "valid paid",
			body: map[string]any{
				"nasabah_id":   "00000000-0000-0000-0000-000000000001",
				"rekening_id":  "00000000-0000-0000-0000-000000000002",
				"type":         simpanan_pokok_wajib.TypePokok,
				"amount":       float64(500000),
				"status":       simpanan_pokok_wajib.StatusPaid,
			},
			wantErr: false,
		},
		{
			name: "missing nasabah_id",
			body: map[string]any{
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        simpanan_pokok_wajib.TypePokok,
				"amount":      float64(500000),
			},
			wantErr: true,
			errMsg:  "nasabah_id wajib diisi",
		},
		{
			name: "missing rekening_id",
			body: map[string]any{
				"nasabah_id": "00000000-0000-0000-0000-000000000001",
				"type":       simpanan_pokok_wajib.TypePokok,
				"amount":     float64(500000),
			},
			wantErr: true,
			errMsg:  "rekening_id wajib diisi",
		},
		{
			name: "missing type",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"amount":      float64(500000),
			},
			wantErr: true,
			errMsg:  "type wajib diisi",
		},
		{
			name: "missing amount",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        simpanan_pokok_wajib.TypePokok,
			},
			wantErr: true,
			errMsg:  "amount wajib diisi",
		},
		{
			name: "zero amount",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        simpanan_pokok_wajib.TypePokok,
				"amount":      float64(0),
			},
			wantErr: true,
			errMsg:  "amount harus lebih dari 0",
		},
		{
			name: "negative amount",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        simpanan_pokok_wajib.TypePokok,
				"amount":      float64(-100),
			},
			wantErr: true,
			errMsg:  "amount harus lebih dari 0",
		},
		{
			name: "invalid type",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        "sukarela",
				"amount":      float64(500000),
			},
			wantErr: true,
			errMsg:  "type tidak valid",
		},
		{
			name: "invalid status",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        simpanan_pokok_wajib.TypePokok,
				"amount":      float64(500000),
				"status":      "cancelled",
			},
			wantErr: true,
			errMsg:  "status tidak valid",
		},
		{
			name: "wajib missing period",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        simpanan_pokok_wajib.TypeWajib,
				"amount":      float64(100000),
			},
			wantErr: true,
			errMsg:  "period wajib diisi untuk simpanan wajib",
		},
		{
			name: "valid overdue",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        simpanan_pokok_wajib.TypePokok,
				"amount":      float64(500000),
				"status":      simpanan_pokok_wajib.StatusOverdue,
			},
			wantErr: false,
		},
		{
			name: "valid refunded",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"type":        simpanan_pokok_wajib.TypePokok,
				"amount":      float64(500000),
				"status":      simpanan_pokok_wajib.StatusRefunded,
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

func TestSimpananPokokWajibHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	r := chi.NewRouter()
	r.Use(spwScopeMiddleware(spwTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			spwRespondForbidden(w)
			return
		}
		assert.Equal(t, spwTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			spwRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			spwRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validSPWBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestSimpananPokokWajibHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	r := chi.NewRouter()
	r.Use(spwScopeMiddleware(spwTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			spwRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			spwRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			spwRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"type":    simpanan_pokok_wajib.TypePokok,
		"amount":  float64(500000),
		// Missing nasabah_id, rekening_id
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

func TestSimpananPokokWajibHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			spwRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			spwRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			spwRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validSPWBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSimpananPokokWajibHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	r := chi.NewRouter()
	r.Use(spwScopeMiddleware(spwTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			spwRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			spwRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			spwRespondValidationError(w, err.Error())
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

func validSPWBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"nasabah_id":  "00000000-0000-0000-0000-000000000001",
		"rekening_id": "00000000-0000-0000-0000-000000000002",
		"type":        simpanan_pokok_wajib.TypePokok,
		"amount":      float64(500000),
		"status":      simpanan_pokok_wajib.StatusPending,
	})
	require.NoError(t, err)
	return body
}

func spwRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func spwRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func spwRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
