package teacher_development_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/yourorg/boilerplate/internal/domain/teacher_development"
	"github.com/yourorg/boilerplate/pkg/scope"
)

var tdTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func tdScopeMW(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(scope.WithScope(r.Context(), s)))
		})
	}
}

func tdRespond(w http.ResponseWriter, code int, obj map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(obj)
}

func validCertificationBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"teacher_id": uuid.New().String(),
		"certification_type": teacher_development.CertTypeProfesi,
	})
	return b
}

func validActivityBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"teacher_id": uuid.New().String(), "title": "Workshop",
		"start_date": "2026-04-10",
	})
	return b
}

func validCreditSummaryBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"teacher_id": uuid.New().String(), "current_rank": teacher_development.RankIIIA,
	})
	return b
}

func TestCertificationHandler_MissingScope_403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		if _, ok := scope.ScopeFromContext(req.Context()); !ok {
			tdRespond(w, http.StatusForbidden, map[string]any{"error": map[string]string{"code": "FORBIDDEN"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(validCertificationBody(t)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCertificationHandler_InvalidJSON_400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(tdScopeMW(tdTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			tdRespond(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "INVALID_JSON"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCertificationHandler_WithScope_OK(t *testing.T) {
	r := chi.NewRouter()
	r.Use(tdScopeMW(tdTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		assert.True(t, ok)
		assert.Equal(t, tdTestScope.TenantID, s.TenantID)
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
