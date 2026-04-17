package shu_test

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

	"github.com/yourorg/boilerplate/internal/domain/shu"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var shuTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func shuScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- SHUPeriode handler tests ----------------------------------------------------

func TestSHUPeriodeHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/shu-periode", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			shuRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validSPBody(t)
	req := httptest.NewRequest(http.MethodPost, "/shu-periode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	errMap, ok := resp["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", errMap["code"])
}

func TestSHUPeriodeHandler_WithScope_PassesThrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(shuScopeMiddleware(shuTestScope))
	r.Post("/shu-periode", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, shuTestScope.TenantID, s.TenantID)
		assert.Equal(t, shuTestScope.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/shu-periode", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSHUPeriodeHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	body := validSPBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestSHUPeriodeHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	r := chi.NewRouter()
	r.Use(shuScopeMiddleware(shuTestScope))
	r.Post("/shu-periode", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			shuRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			shuRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			shuRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validSPBody(t)
	req := httptest.NewRequest(http.MethodPost, "/shu-periode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestSHUPeriodeHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	r := chi.NewRouter()
	r.Use(shuScopeMiddleware(shuTestScope))
	r.Post("/shu-periode", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			shuRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			shuRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			shuRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"period_start": "2026-01-01",
		// Missing tahun_buku, period_end, total_pendapatan, etc.
	})
	req := httptest.NewRequest(http.MethodPost, "/shu-periode", bytes.NewReader(body))
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

func TestSHUPeriodeHandler_FullCreatePipeline_NoScope(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	r := chi.NewRouter()
	r.Post("/shu-periode", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			shuRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			shuRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			shuRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validSPBody(t)
	req := httptest.NewRequest(http.MethodPost, "/shu-periode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSHUPeriodeHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &shu.SHUPeriodeDescriptor{}
	r := chi.NewRouter()
	r.Use(shuScopeMiddleware(shuTestScope))
	r.Post("/shu-periode", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			shuRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			shuRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			shuRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/shu-periode", bytes.NewReader([]byte("bukan json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// -- SHUAnggota handler tests ----------------------------------------------------

func TestSHUAnggotaHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	body := validSABody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestSHUAnggotaHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	r := chi.NewRouter()
	r.Use(shuScopeMiddleware(shuTestScope))
	r.Post("/shu-anggota", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			shuRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			shuRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			shuRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validSABody(t)
	req := httptest.NewRequest(http.MethodPost, "/shu-anggota", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestSHUAnggotaHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &shu.SHUAnggotaDescriptor{}
	r := chi.NewRouter()
	r.Use(shuScopeMiddleware(shuTestScope))
	r.Post("/shu-anggota", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			shuRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			shuRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			shuRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"nasabah_id": "00000000-0000-0000-0000-000000000002",
		// Missing shu_periode_id, avg_simpanan, total_transaksi, etc.
	})
	req := httptest.NewRequest(http.MethodPost, "/shu-anggota", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// -- Invalid UUID in path --------------------------------------------------------

func TestSHUPeriodeHandler_InvalidUUIDInPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(shuScopeMiddleware(shuTestScope))
	r.Get("/shu-periode/{id}", func(w http.ResponseWriter, req *http.Request) {
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
		{"not-a-uuid", "bad-id", http.StatusBadRequest},
		{"valid-uuid", uuid.New().String(), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/shu-periode/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// -- helpers ---------------------------------------------------------------------

func validSPBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"tahun_buku":        2026,
		"period_start":      "2026-01-01",
		"period_end":        "2026-12-31",
		"total_pendapatan":  50000000,
		"total_beban":       30000000,
		"shu_bruto":         20000000,
		"shu_neto":          20000000,
		"status":            shu.SPStatusCalculated,
	})
	require.NoError(t, err)
	return body
}

func validSABody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"shu_periode_id":    "00000000-0000-0000-0000-000000000001",
		"nasabah_id":        "00000000-0000-0000-0000-000000000002",
		"avg_simpanan":      5000000,
		"total_transaksi":   10000000,
		"active_days":       365,
		"jasa_modal":        500000,
		"jasa_usaha":        300000,
		"total_shu":         800000,
		"distribution_method": shu.DistMethodCreditTabungan,
	})
	require.NoError(t, err)
	return body
}

func shuRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func shuRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func shuRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
