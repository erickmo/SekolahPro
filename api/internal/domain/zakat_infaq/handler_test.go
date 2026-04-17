package zakat_infaq_test

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

	"github.com/yourorg/boilerplate/internal/domain/zakat_infaq"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var ziTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func ziScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- Scope enforcement tests -----------------------------------------------------

func TestZakatInfaqHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			ziRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validZCBody(t)
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

func TestZakatInfaqHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(ziScopeMiddleware(ziTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, ziTestScope.TenantID, s.TenantID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// -- Invalid JSON body -----------------------------------------------------------

func TestZakatInfaqHandler_InvalidJSONBody(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	r := chi.NewRouter()
	r.Use(ziScopeMiddleware(ziTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			ziRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			ziRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			ziRespondValidationError(w, err.Error())
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

func TestZakatInfaqHandler_InvalidUUIDInPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(ziScopeMiddleware(ziTestScope))
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

// -- Full handler pipeline simulation -------------------------------------------

func TestZakatInfaqHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	r := chi.NewRouter()
	r.Use(ziScopeMiddleware(ziTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			ziRespondForbidden(w)
			return
		}
		assert.Equal(t, ziTestScope.TenantID, s.TenantID)

		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			ziRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			ziRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validZCBody(t)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestZakatInfaqHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	r := chi.NewRouter()
	r.Use(ziScopeMiddleware(ziTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			ziRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			ziRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			ziRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"zakat_type": "zakat_mal",
		// Missing: muzakki_name, amount, payment_method
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

// -- Cross-descriptor validation tests -----------------------------------------

func TestZakatInfaqHandler_DistributionCreate_Valid(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	var parsed map[string]any
	body := validZDBody(t)
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)
	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestZakatInfaqHandler_InfaqCreate_Valid(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	var parsed map[string]any
	body := validInfaqBody(t)
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)
	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestZakatInfaqHandler_MustahikCreate_Valid(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	var parsed map[string]any
	body := validMustahikBody(t)
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)
	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestZakatInfaqHandler_TazirFundCreate_Valid(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	var parsed map[string]any
	body := validTFBody(t)
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)
	err = d.Validate(parsed)
	assert.NoError(t, err)
}

// -- helpers ---------------------------------------------------------------------

func validZCBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(validZakatCollectionData())
	require.NoError(t, err)
	return body
}

func validZDBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(validZakatDistributionData())
	require.NoError(t, err)
	return body
}

func validInfaqBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(validInfaqData())
	require.NoError(t, err)
	return body
}

func validMustahikBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(validMustahikData())
	require.NoError(t, err)
	return body
}

func validTFBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(validTazirFundData())
	require.NoError(t, err)
	return body
}

func ziRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func ziRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func ziRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
