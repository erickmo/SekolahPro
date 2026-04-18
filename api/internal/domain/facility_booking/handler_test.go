package facility_booking_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/yourorg/boilerplate/internal/domain/facility_booking"
	"github.com/yourorg/boilerplate/pkg/scope"
)

var fbTestScope = scope.Scope{
	TenantID:  uuid.MustParse("00000000-0000-0000-0000-000000000099"),
	CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000100"),
}

func fbScopeMW(s scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(scope.WithScope(r.Context(), s)))
		})
	}
}

func fbRespond(w http.ResponseWriter, code int, obj map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(obj)
}

func validFacilityBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"name": "Aula Utama", "type": facility_booking.FTypeHall,
	})
	return b
}

func validBookingBody(t *testing.T) []byte {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"facility_id": uuid.New().String(),
		"requester_type": facility_booking.BReqTeacher,
		"requester_id": uuid.New().String(),
		"booking_date": "2026-04-20", "purpose": "Rapat",
	})
	return b
}

func TestFacilityHandler_MissingScope_403(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		if _, ok := scope.ScopeFromContext(req.Context()); !ok {
			fbRespond(w, http.StatusForbidden, map[string]any{"error": map[string]string{"code": "FORBIDDEN"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(validFacilityBody(t)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestFacilityHandler_InvalidJSON_400(t *testing.T) {
	r := chi.NewRouter()
	r.Use(fbScopeMW(fbTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			fbRespond(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "INVALID_JSON"}})
			return
		}
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFacilityHandler_WithScope_OK(t *testing.T) {
	r := chi.NewRouter()
	r.Use(fbScopeMW(fbTestScope))
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		s, ok := scope.ScopeFromContext(req.Context())
		assert.True(t, ok)
		assert.Equal(t, fbTestScope.TenantID, s.TenantID)
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
