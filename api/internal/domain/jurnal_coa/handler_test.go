package jurnal_coa_test

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

	"github.com/yourorg/boilerplate/internal/domain/jurnal_coa"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// -- test fixtures ---------------------------------------------------------------

var jcTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func jcScopeMiddleware(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := scope.WithScope(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -- COA handler tests -----------------------------------------------------------

func TestCOAHandler_MissingScope_Returns403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/coa", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			jcRespondForbidden(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	body := validCOABody(t)
	req := httptest.NewRequest(http.MethodPost, "/coa", bytes.NewReader(body))
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

func TestCOAHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	body := validCOABody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestCOAHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	r := chi.NewRouter()
	r.Use(jcScopeMiddleware(jcTestScope))
	r.Post("/coa", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			jcRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			jcRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			jcRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validCOABody(t)
	req := httptest.NewRequest(http.MethodPost, "/coa", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCOAHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	r := chi.NewRouter()
	r.Use(jcScopeMiddleware(jcTestScope))
	r.Post("/coa", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			jcRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			jcRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			jcRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"account_name": "Kas",
		// Missing account_code, account_type, normal_balance, level
	})
	req := httptest.NewRequest(http.MethodPost, "/coa", bytes.NewReader(body))
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

func TestCOAHandler_FullCreatePipeline_InvalidJSON(t *testing.T) {
	d := &jurnal_coa.COADescriptor{}
	r := chi.NewRouter()
	r.Use(jcScopeMiddleware(jcTestScope))
	r.Post("/coa", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			jcRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			jcRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			jcRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/coa", bytes.NewReader([]byte("bukan json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// -- Jurnal handler tests --------------------------------------------------------

func TestJurnalHandler_ValidCreateBody_PassesValidation(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	body := validJurnalBody(t)

	var parsed map[string]any
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)
	require.NoError(t, err)

	err = d.Validate(parsed)
	assert.NoError(t, err)
}

func TestJurnalHandler_FullCreatePipeline_Valid(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	r := chi.NewRouter()
	r.Use(jcScopeMiddleware(jcTestScope))
	r.Post("/jurnal", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			jcRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			jcRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			jcRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body := validJurnalBody(t)
	req := httptest.NewRequest(http.MethodPost, "/jurnal", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestJurnalHandler_FullCreatePipeline_ValidationFails(t *testing.T) {
	d := &jurnal_coa.JurnalDescriptor{}
	r := chi.NewRouter()
	r.Use(jcScopeMiddleware(jcTestScope))
	r.Post("/jurnal", func(w http.ResponseWriter, req *http.Request) {
		_, ok := scope.ScopeFromContext(req.Context())
		if !ok {
			jcRespondForbidden(w)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			jcRespondInvalidJSON(w)
			return
		}
		if err := d.Validate(body); err != nil {
			jcRespondValidationError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	body, _ := json.Marshal(map[string]any{
		"description": "Test",
		// Missing journal_number, journal_date, source_type, period_id, total_debit, total_credit
	})
	req := httptest.NewRequest(http.MethodPost, "/jurnal", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// -- Invalid UUID in path --------------------------------------------------------

func TestJurnalHandler_InvalidUUIDInPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(jcScopeMiddleware(jcTestScope))
	r.Get("/jurnal/{id}", func(w http.ResponseWriter, req *http.Request) {
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
			req := httptest.NewRequest(http.MethodGet, "/jurnal/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// -- helpers ---------------------------------------------------------------------

func validCOABody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"account_code":   "1101",
		"account_name":   "Kas Teller",
		"level":          2,
		"account_type":   jurnal_coa.AccountTypeAsset,
		"normal_balance": jurnal_coa.NormalBalanceDebit,
		"coop_type_required": jurnal_coa.CoopTypeBoth,
	})
	require.NoError(t, err)
	return body
}

func validJurnalBody(t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"journal_number": "JRN-2026-01-00000001",
		"journal_date":   "2026-01-15",
		"description":    "Setoran tunai tabungan",
		"source_type":    jurnal_coa.SourceTypeTransaction,
		"period_id":      "00000000-0000-0000-0000-000000000001",
		"total_debit":    100000,
		"total_credit":   100000,
		"status":         jurnal_coa.JurnalStatusUnposted,
	})
	require.NoError(t, err)
	return body
}

func jcRespondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "FORBIDDEN", "message": "scope organisasi tidak teridentifikasi"},
	})
}

func jcRespondInvalidJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "INVALID_JSON", "message": "request body bukan JSON valid"},
	})
}

func jcRespondValidationError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "VALIDATION_ERROR", "message": msg},
	})
}
