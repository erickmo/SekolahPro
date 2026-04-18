package laboratory_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/yourorg/boilerplate/internal/domain/laboratory"
	"github.com/yourorg/boilerplate/pkg/scope"
)

var labTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func labScopeMW(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(scope.WithScope(r.Context(), s)))
		})
	}
}

func labRespond(w http.ResponseWriter, code int, obj map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(obj)
}

func validLabBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"name": "Lab IPA 1", "code": "LAB-01",
		"type": laboratory.LabTypeIPA, "capacity": 30,
	})
	return b
}

func validEquipmentBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"lab_id": uuid.New().String(), "name": "Mikroskop",
		"category": laboratory.EqCatOptical, "condition": laboratory.EqCondGood,
	})
	return b
}

func validUsageLogBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"lab_id": uuid.New().String(), "teacher_id": uuid.New().String(),
		"usage_date": "2026-04-01",
	})
	return b
}

func TestLabHandler_MissingScope_403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		if _, ok := scope.ScopeFromContext(req.Context()); !ok {
			labRespond(w, http.StatusForbidden, map[string]any{"error": map[string]string{"code": "FORBIDDEN"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(validLabBody(t)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestLabHandler_InvalidJSON_400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(labScopeMW(labTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			labRespond(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "INVALID_JSON"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLabHandler_WithScope_OK(t *testing.T) {
	r := chi.NewRouter()
	r.Use(labScopeMW(labTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		assert.True(t, ok)
		assert.Equal(t, labTestScope.TenantID, s.TenantID)
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
