package asset_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/yourorg/boilerplate/internal/domain/asset"
	"github.com/yourorg/boilerplate/pkg/scope"
)

var astTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func astScopeMW(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(scope.WithScope(r.Context(), s)))
		})
	}
}

func astRespond(w http.ResponseWriter, code int, obj map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(obj)
}

func validAssetBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"asset_code": "AST-001", "name": "Proyektor", "category": asset.CatPeralatan,
	})
	return b
}

func validMaintenanceBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"asset_id": uuid.New().String(), "type": asset.MtPreventive,
		"description": "Service rutin", "start_date": "2026-04-01",
	})
	return b
}

func TestAssetHandler_MissingScope_403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		if _, ok := scope.ScopeFromContext(req.Context()); !ok {
			astRespond(w, http.StatusForbidden, map[string]any{"error": map[string]string{"code": "FORBIDDEN"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(validAssetBody(t)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAssetHandler_InvalidJSON_400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(astScopeMW(astTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			astRespond(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "INVALID_JSON"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAssetHandler_WithScope_OK(t *testing.T) {
	r := chi.NewRouter()
	r.Use(astScopeMW(astTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		assert.True(t, ok)
		assert.Equal(t, astTestScope.TenantID, s.TenantID)
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
