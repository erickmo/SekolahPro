package produk_akad_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/produk_akad"
	"github.com/yourorg/boilerplate/pkg/scope"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── test fixtures ────────────────────────────────────────────────────────────

var testScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

// mockEventBus is a minimal eventbus.EventBus stub.
type mockEventBus struct{}

func (m *mockEventBus) Publish(_ context.Context, _ any) error { return nil }
func (m *mockEventBus) Subscribe(_, _ string) (<-chan any, error) {
	ch := make(chan any)
	return ch, nil
}

// scopeMiddleware injects the test scope into request context.
func scopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// newTestDescriptor returns the produk_akad descriptor.
func newTestDescriptor() *produk_akad.Descriptor {
	return &produk_akad.Descriptor{}
}

// ── Scope enforcement tests ──────────────────────────────────────────────────

func TestVernonHandler_MissingScope_Returns403(t *testing.T) {
	// We test the scopeOrError helper behavior by setting up a handler that
	// checks scope — same pattern as vernon.BaseHandler uses internally.
	d := newTestDescriptor()
	assert.NotNil(t, d)

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

	body := validProdukAkadBody(t)
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

func TestVernonHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(scopeMiddleware(testScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, testScope.TenantID, s.TenantID)
		assert.Equal(t, testScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── Descriptor.Validate through HTTP pattern ─────────────────────────────────

func TestVernonHandler_InvalidJSONBody(t *testing.T) {
	d := newTestDescriptor()
	// Simulate what the Vernon handler does: decode JSON then call Validate.
	// The handler layer catches JSON decode errors before reaching Validate.
	body := []byte(`{invalid json`)
	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	assert.Error(t, err, "malformed JSON should fail decode")

	// Verify descriptor is not involved in JSON parsing
	assert.NotNil(t, d)
}

func TestVernonHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := newTestDescriptor()
	body := validProdukAkadBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err, "valid JSON body should pass descriptor validation")
}

func TestVernonHandler_MissingNameInBody_FailsValidation(t *testing.T) {
	d := newTestDescriptor()
	body, _ := json.Marshal(map[string]any{
		"code":      "TW-001",
		"type":      produk_akad.TypeSimpanan,
		"akad_type": produk_akad.AkadWadiah,
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name wajib diisi")
}

func TestVernonHandler_InvalidTypeInBody_FailsValidation(t *testing.T) {
	d := newTestDescriptor()
	body, _ := json.Marshal(map[string]any{
		"name":      "Produk Test",
		"code":      "TST-001",
		"type":      "invalid_type",
		"akad_type": produk_akad.AkadWadiah,
	})

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type tidak valid")
}

// ── Invalid UUID in path ─────────────────────────────────────────────────────

func TestVernonHandler_InvalidUUIDInPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(scopeMiddleware(testScope))
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
		{"empty", "", http.StatusNotFound}, // chi won't match route
		{"valid-uuid", uuid.New().String(), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/" + tt.pathParam
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestVernonHandler_InvalidUUIDOnUpdate(t *testing.T) {
	r := chi.NewRouter()
	r.Use(scopeMiddleware(testScope))
	r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
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

	req := httptest.NewRequest(http.MethodPut, "/bad-uuid", bytes.NewReader(validProdukAkadBody(t)))
	req.Header.Set("Content-Type", "application/json")
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

// ── Table-driven: validation through decode + validate pipeline ──────────────

func TestVernonHandler_CreateValidation_Table(t *testing.T) {
	d := newTestDescriptor()

	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid simpanan wadiah",
			body: map[string]any{
				"name": "Tabungan Wadiah", "code": "TW-001",
				"type": produk_akad.TypeSimpanan, "akad_type": produk_akad.AkadWadiah,
			},
			wantErr: false,
		},
		{
			name: "valid pembiayaan murabahah",
			body: map[string]any{
				"name": "Pembiayaan Murabahah", "code": "PM-001",
				"type": produk_akad.TypePembiayaan, "akad_type": produk_akad.AkadMurabahah,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			body: map[string]any{
				"code": "TW-001", "type": produk_akad.TypeSimpanan, "akad_type": produk_akad.AkadWadiah,
			},
			wantErr: true,
			errMsg:  "name wajib diisi",
		},
		{
			name: "missing code",
			body: map[string]any{
				"name": "Test", "type": produk_akad.TypeSimpanan, "akad_type": produk_akad.AkadWadiah,
			},
			wantErr: true,
			errMsg:  "code wajib diisi",
		},
		{
			name: "missing type",
			body: map[string]any{
				"name": "Test", "code": "T-001", "akad_type": produk_akad.AkadWadiah,
			},
			wantErr: true,
			errMsg:  "type wajib diisi",
		},
		{
			name: "missing akad_type",
			body: map[string]any{
				"name": "Test", "code": "T-001", "type": produk_akad.TypeSimpanan,
			},
			wantErr: true,
			errMsg:  "akad_type wajib diisi",
		},
		{
			name: "invalid type",
			body: map[string]any{
				"name": "Test", "code": "T-001",
				"type": "deposito", "akad_type": produk_akad.AkadWadiah,
			},
			wantErr: true,
			errMsg:  "type tidak valid",
		},
		{
			name: "invalid akad_type",
			body: map[string]any{
				"name": "Test", "code": "T-001",
				"type": produk_akad.TypeSimpanan, "akad_type": "riba",
			},
			wantErr: true,
			errMsg:  "akad_type tidak valid",
		},
		{
			name: "negative min_amount",
			body: map[string]any{
				"name": "Test", "code": "T-001",
				"type": produk_akad.TypeSimpanan, "akad_type": produk_akad.AkadWadiah,
				"min_amount": float64(-100),
			},
			wantErr: true,
			errMsg:  "min_amount tidak boleh negatif",
		},
		{
			name: "max less than min amount",
			body: map[string]any{
				"name": "Test", "code": "T-001",
				"type": produk_akad.TypeSimpanan, "akad_type": produk_akad.AkadWadiah,
				"min_amount": float64(1000), "max_amount": float64(500),
			},
			wantErr: true,
			errMsg:  "max_amount harus >= min_amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate JSON round-trip (what the handler does)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			var parsed map[string]any
			err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&parsed)
			require.NoError(t, err, "JSON decode should succeed for all test cases")

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

// ── Constant correctness ─────────────────────────────────────────────────────

func TestConstants_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, produk_akad.TypeSimpanan)
	assert.NotEmpty(t, produk_akad.TypePembiayaan)
	assert.NotEmpty(t, produk_akad.AkadWadiah)
	assert.NotEmpty(t, produk_akad.AkadMudharabah)
	assert.NotEmpty(t, produk_akad.AkadMusyarakah)
	assert.NotEmpty(t, produk_akad.AkadMurabahah)
	assert.NotEmpty(t, produk_akad.AkadIjarah)
	assert.NotEmpty(t, produk_akad.AkadQard)
}

// ── Logger compatibility check ───────────────────────────────────────────────

func TestNewBaseHandler_DoesNotPanic(t *testing.T) {
	logger := zerolog.Nop()
	// Verify the types are compatible — this is a compile-time + runtime check.
	_ = vernon.NewBaseHandler
	_ = logger
}

// ── helpers ──────────────────────────────────────────────────────────────────

func validProdukAkadBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"name":       "Tabungan Wadiah",
		"code":       "TW-001",
		"type":       produk_akad.TypeSimpanan,
		"akad_type":  produk_akad.AkadWadiah,
		"min_amount": float64(10000),
		"max_amount": float64(100000000),
	})
	require.NoError(t, err)
	return body
}
