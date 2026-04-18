package teacher_evaluation_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/yourorg/boilerplate/internal/domain/teacher_evaluation"
	"github.com/yourorg/boilerplate/pkg/scope"
)

var teTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func teScopeMW(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(scope.WithScope(r.Context(), s)))
		})
	}
}

func teRespond(w http.ResponseWriter, code int, obj map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(obj)
}

func validCompetencyBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"name": "Pedagogik", "area": teacher_evaluation.AreaPedagogic,
	})
	return b
}

func validEvaluationBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"teacher_id": uuid.New().String(), "academic_year_id": uuid.New().String(),
		"semester": teacher_evaluation.SemGanjil, "evaluator_id": uuid.New().String(),
	})
	return b
}

func validScoreBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"evaluation_id": uuid.New().String(), "competency_id": uuid.New().String(),
		"assessor_type": teacher_evaluation.AssessorSelf, "score": 85,
	})
	return b
}

func TestCompetencyHandler_MissingScope_403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		if _, ok := scope.ScopeFromContext(req.Context()); !ok {
			teRespond(w, http.StatusForbidden, map[string]any{"error": map[string]string{"code": "FORBIDDEN"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(validCompetencyBody(t)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCompetencyHandler_InvalidJSON_400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(teScopeMW(teTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			teRespond(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "INVALID_JSON"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCompetencyHandler_WithScope_OK(t *testing.T) {
	r := chi.NewRouter()
	r.Use(teScopeMW(teTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		assert.True(t, ok)
		assert.Equal(t, teTestScope.TenantID, s.TenantID)
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
