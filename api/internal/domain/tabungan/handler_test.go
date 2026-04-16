package tabungan_test

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

	"github.com/yourorg/boilerplate/internal/domain/tabungan"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var tabunganTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func tabunganScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestTabunganHandler_MissingScope_Returns403(t *testing.T) {
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

	body := validTabunganBody(t)
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

func TestTabunganHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(tabunganScopeMiddleware(tabunganTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, tabunganTestScope.TenantID, s.TenantID)
		assert.Equal(t, tabunganTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestTabunganHandler_InvalidJSONBody(t *testing.T) {
	d := &tabungan.Descriptor{}
	r := chi.NewRouter()
	r.Use(tabunganScopeMiddleware(tabunganTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			tabunganRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			tabunganRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			tabunganRespondValidationError(w, err.Error())
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

func TestTabunganHandler_InvalidUUIDInPath_GetByID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(tabunganScopeMiddleware(tabunganTestScope))
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

func TestTabunganHandler_InvalidUUIDInPath_Delete(t *testing.T) {
	r := chi.NewRouter()
	r.Use(tabunganScopeMiddleware(tabunganTestScope))
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

func TestTabunganHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &tabungan.Descriptor{}
	body := validTabunganBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestTabunganHandler_NegativeBalance_FailsValidation(t *testing.T) {
	d := &tabungan.Descriptor{}
	body, _ := json.Marshal(map[string]any{
		"nasabah_id":    "00000000-0000-0000-0000-000000000001",
		"rekening_id":   "00000000-0000-0000-0000-000000000002",
		"product_type":  tabungan.ProductRegular,
		"balance":       float64(-1000),
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "balance tidak boleh negatif")
}

// -- Table-driven: full validation pipeline --------------------------------------

func TestTabunganHandler_CreateValidation_Table(t *testing.T) {
	d := &tabungan.Descriptor{}

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid regular active",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"balance":       float64(1000000),
				"status":        tabungan.StatusActive,
			},
			wantErr: false,
		},
		{
			name: "valid education",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductEducation,
				"balance":       float64(500000),
			},
			wantErr: false,
		},
		{
			name: "valid holiday",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductHoliday,
			},
			wantErr: false,
		},
		{
			name: "valid qurban",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductQurban,
			},
			wantErr: false,
		},
		{
			name: "valid goal",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductGoal,
				"goal_amount":   float64(10000000),
			},
			wantErr: false,
		},
		{
			name: "valid dormant",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"status":        tabungan.StatusDormant,
			},
			wantErr: false,
		},
		{
			name: "valid frozen",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"status":        tabungan.StatusFrozen,
			},
			wantErr: false,
		},
		{
			name: "valid closed",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"status":        tabungan.StatusClosed,
			},
			wantErr: false,
		},
		{
			name: "missing nasabah_id",
			body: map[string]any{
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"balance":       float64(1000000),
			},
			wantErr: true,
			errMsg:  "nasabah_id wajib diisi",
		},
		{
			name: "missing rekening_id",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"product_type":  tabungan.ProductRegular,
				"balance":       float64(1000000),
			},
			wantErr: true,
			errMsg:  "rekening_id wajib diisi",
		},
		{
			name: "missing product_type",
			body: map[string]any{
				"nasabah_id":  "00000000-0000-0000-0000-000000000001",
				"rekening_id": "00000000-0000-0000-0000-000000000002",
				"balance":     float64(1000000),
			},
			wantErr: true,
			errMsg:  "product_type wajib diisi",
		},
		{
			name: "invalid product_type",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  "deposito",
			},
			wantErr: true,
			errMsg:  "product_type tidak valid",
		},
		{
			name: "negative balance",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"balance":       float64(-1),
			},
			wantErr: true,
			errMsg:  "balance tidak boleh negatif",
		},
		{
			name: "hold exceeds balance",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"balance":       float64(100),
				"hold_balance":  float64(200),
			},
			wantErr: true,
			errMsg:  "hold_balance tidak boleh lebih besar dari balance",
		},
		{
			name: "invalid status",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"status":        "pending",
			},
			wantErr: true,
			errMsg:  "status tidak valid",
		},
		{
			name: "goal_amount on non-goal product",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"goal_amount":   float64(10000000),
			},
			wantErr: true,
			errMsg:  "goal_amount hanya untuk product_type goal",
		},
		{
			name: "zero balance zero hold",
			body: map[string]any{
				"nasabah_id":    "00000000-0000-0000-0000-000000000001",
				"rekening_id":   "00000000-0000-0000-0000-000000000002",
				"product_type":  tabungan.ProductRegular,
				"balance":       float64(0),
				"hold_balance":  float64(0),
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

func TestTabunganHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &tabungan.Descriptor{}
	r := chi.NewRouter()
	r.Use(tabunganScopeMiddleware(tabunganTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			tabunganRespondForbidden(w)
			return
		}
		assert.Equal(t, tabunganTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			tabunganRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			tabunganRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validTabunganBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestTabunganHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &tabungan.Descriptor{}
	r := chi.NewRouter()
	r.Use(tabunganScopeMiddleware(tabunganTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			tabunganRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			tabunganRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			tabunganRespondValidationError(w, err.Error())
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

func TestTabunganHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &tabungan.Descriptor{}
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			tabunganRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			tabunganRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			tabunganRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validTabunganBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTabunganHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &tabungan.Descriptor{}
	r := chi.NewRouter()
	r.Use(tabunganScopeMiddleware(tabunganTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			tabunganRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			tabunganRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			tabunganRespondValidationError(w, err.Error())
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

func validTabunganBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"nasabah_id":    "00000000-0000-0000-0000-000000000001",
		"rekening_id":   "00000000-0000-0000-0000-000000000002",
		"product_type":  tabungan.ProductRegular,
		"balance":       float64(1000000),
		"status":        tabungan.StatusActive,
	})
	require.NoError(t, err)
	return body
}

func tabunganRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func tabunganRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func tabunganRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
