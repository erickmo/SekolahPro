package deposito_test

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

	"github.com/yourorg/boilerplate/internal/domain/deposito"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var depositoTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func depositoScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestDepositoHandler_MissingScope_Returns403(t *testing.T) {
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

	body := validDepositoBody(t)
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

func TestDepositoHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(depositoScopeMiddleware(depositoTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, depositoTestScope.TenantID, s.TenantID)
		assert.Equal(t, depositoTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestDepositoHandler_InvalidJSONBody(t *testing.T) {
	d := &deposito.Descriptor{}
	r := chi.NewRouter()
	r.Use(depositoScopeMiddleware(depositoTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			depositoRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			depositoRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			depositoRespondValidationError(w, err.Error())
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

func TestDepositoHandler_InvalidUUIDInPath_GetByID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(depositoScopeMiddleware(depositoTestScope))
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

func TestDepositoHandler_InvalidUUIDInPath_Delete(t *testing.T) {
	r := chi.NewRouter()
	r.Use(depositoScopeMiddleware(depositoTestScope))
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

func TestDepositoHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &deposito.Descriptor{}
	body := validDepositoBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestDepositoHandler_NegativeRate_FailsValidation(t *testing.T) {
	d := &deposito.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"nasabah_id":   "00000000-0000-0000-0000-000000000001",
		"rekening_id":  "00000000-0000-0000-0000-000000000002",
		"principal":    float64(10000000),
		"tenor_months": float64(12),
		"rate":         float64(-1.5),
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rate tidak boleh negatif")
}

// -- Table-driven: full validation pipeline --------------------------------------

func TestDepositoHandler_CreateValidation_Table(t *testing.T) {
	d := &deposito.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid active 12 months monthly",
			body: map[string]any{
				"nasabah_id":      "00000000-0000-0000-0000-000000000001",
				"rekening_id":     "00000000-0000-0000-0000-000000000002",
				"principal":       float64(10000000),
				"tenor_months":    float64(12),
				"rate":            float64(5.5),
				"status":          deposito.StatusActive,
				"payment_method":  deposito.PaymentMonthly,
			},
			wantErr: false,
		},
		{
			name: "valid 1 month",
			body: map[string]any{
				"nasabah_id":   "00000000-0000-0000-0000-000000000001",
				"rekening_id":  "00000000-0000-0000-0000-000000000002",
				"principal":    float64(5000000),
				"tenor_months": float64(1),
				"rate":         float64(3.0),
			},
			wantErr: false,
		},
		{
			name: "valid 3 months",
			body: map[string]any{
				"nasabah_id":   "00000000-0000-0000-0000-000000000001",
				"rekening_id":  "00000000-0000-0000-0000-000000000002",
				"principal":    float64(5000000),
				"tenor_months": float64(3),
				"rate":         float64(4.0),
			},
			wantErr: false,
		},
		{
			name: "valid 6 months",
			body: map[string]any{
				"nasabah_id":   "00000000-0000-0000-0000-000000000001",
				"rekening_id":  "00000000-0000-0000-0000-000000000002",
				"principal":    float64(5000000),
				"tenor_months": float64(6),
				"rate":         float64(4.5),
			},
			wantErr: false,
		},
		{
			name: "valid 24 months",
			body: map[string]any{
				"nasabah_id":   "00000000-0000-0000-0000-000000000001",
				"rekening_id":  "00000000-0000-0000-0000-000000000002",
				"principal":    float64(5000000),
				"tenor_months": float64(24),
				"rate":         float64(6.0),
			},
			wantErr: false,
		},
		{
			name: "valid at_maturity",
			body: map[string]any{
				"nasabah_id":      "00000000-0000-0000-0000-000000000001",
				"rekening_id":     "00000000-0000-0000-0000-000000000002",
				"principal":       float64(10000000),
				"tenor_months":    float64(12),
				"rate":            float64(5.5),
				"payment_method":  deposito.PaymentMaturity,
			},
			wantErr: false,
		},
		{
			name: "valid compounded",
			body: map[string]any{
				"nasabah_id":      "00000000-0000-0000-0000-000000000001",
				"rekening_id":     "00000000-0000-0000-0000-000000000002",
				"principal":       float64(10000000),
				"tenor_months":    float64(12),
				"rate":            float64(5.5),
				"payment_method":  deposito.PaymentCompound,
			},
			wantErr: false,
		},
		{
			name: "missing nasabah_id",
			body: map[string]any{
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
			},
			wantErr: true,
			errMsg:  "nasabah_id wajib diisi",
		},
		{
			name: "missing rekening_id",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
			},
			wantErr: true,
			errMsg:  "rekening_id wajib diisi",
		},
		{
			name: "missing principal",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
			},
			wantErr: true,
			errMsg:  "principal wajib diisi",
		},
		{
			name: "zero principal",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(0),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
			},
			wantErr: true,
			errMsg:  "principal harus lebih dari 0",
		},
		{
			name: "negative principal",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(-100),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
			},
			wantErr: true,
			errMsg:  "principal harus lebih dari 0",
		},
		{
			name: "missing tenor_months",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"principal":   float64(10000000),
				"rate":        float64(5.5),
			},
			wantErr: true,
			errMsg:  "tenor_months wajib diisi",
		},
		{
			name: "invalid tenor 2 months",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(2),
				"rate":          float64(5.5),
			},
			wantErr: true,
			errMsg:  "tenor_months tidak valid",
		},
		{
			name: "invalid tenor 18 months",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(18),
				"rate":          float64(5.5),
			},
			wantErr: true,
			errMsg:  "tenor_months tidak valid",
		},
		{
			name: "missing rate",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
			},
			wantErr: true,
			errMsg:  "rate wajib diisi",
		},
		{
			name: "negative rate",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(-0.5),
			},
			wantErr: true,
			errMsg:  "rate tidak boleh negatif",
		},
		{
			name: "invalid status",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
				"status":        "pending",
			},
			wantErr: true,
			errMsg:  "status tidak valid",
		},
		{
			name: "invalid payment_method",
			body: map[string]any{
				"nasabah_id":      "00000000-0000-0000-0000-000000000001",
				"rekening_id":     "00000000-0000-0000-0000-000000000002",
				"principal":       float64(10000000),
				"tenor_months":    float64(12),
				"rate":            float64(5.5),
				"payment_method":  "quarterly",
			},
			wantErr: true,
			errMsg:  "payment_method tidak valid",
		},
		{
			name: "valid matured status",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
				"status":        deposito.StatusMatured,
			},
			wantErr: false,
		},
		{
			name: "valid rolled_over status",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
				"status":        deposito.StatusRolledOver,
			},
			wantErr: false,
		},
		{
			name: "valid early_withdrawn status",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
				"status":        deposito.StatusEarlyWithdraw,
			},
			wantErr: false,
		},
		{
			name: "valid closed status",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(5.5),
				"status":        deposito.StatusClosed,
			},
			wantErr: false,
		},
		{
			name: "valid zero rate",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"principal":     float64(10000000),
				"tenor_months":  float64(12),
				"rate":          float64(0),
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

func TestDepositoHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &deposito.Descriptor{}
	r := chi.NewRouter()
	r.Use(depositoScopeMiddleware(depositoTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			depositoRespondForbidden(w)
			return
		}
		assert.Equal(t, depositoTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			depositoRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			depositoRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validDepositoBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestDepositoHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &deposito.Descriptor{}
	r := chi.NewRouter()
	r.Use(depositoScopeMiddleware(depositoTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			depositoRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			depositoRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			depositoRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"principal":     float64(10000000),
		"tenor_months":  float64(12),
		"rate":          float64(5.5),
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

func TestDepositoHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &deposito.Descriptor{}
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			depositoRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			depositoRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			depositoRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validDepositoBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDepositoHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &deposito.Descriptor{}
	r := chi.NewRouter()
	r.Use(depositoScopeMiddleware(depositoTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			depositoRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			depositoRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			depositoRespondValidationError(w, err.Error())
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

func validDepositoBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"nasabah_id":      "00000000-0000-0000-0000-000000000001",
		"rekening_id":     "00000000-0000-0000-0000-000000000002",
		"principal":       float64(10000000),
		"tenor_months":    float64(12),
		"rate":            float64(5.5),
		"status":          deposito.StatusActive,
		"payment_method":  deposito.PaymentMonthly,
	})
	require.NoError(t, err)
	return body
}

func depositoRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func depositoRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func depositoRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
