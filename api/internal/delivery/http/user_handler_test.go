package http_test

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

	"github.com/yourorg/boilerplate/internal/domain/user"
	httpdelivery "github.com/yourorg/boilerplate/internal/delivery/http"
)

func setupUserRouter(t *testing.T) *chi.Mux {
	t.Helper()

	readRepo := &testUserReadRepo{
		byID: map[uuid.UUID]*user.User{
			testUser.ID: testUser,
		},
	}
	writeRepo := &testUserWriteRepo{}

	handler := httpdelivery.NewUserHandler(readRepo, writeRepo, nil, nil)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func setupUserRouterWithScope(t *testing.T) *chi.Mux {
	t.Helper()

	s := newTestScope()

	readRepo := &testUserReadRepo{
		byID: map[uuid.UUID]*user.User{
			testUser.ID: testUser,
		},
	}
	writeRepo := &testUserWriteRepo{}

	handler := httpdelivery.NewUserHandler(readRepo, writeRepo, nil, nil)

	r := chi.NewRouter()
	newRouterWithScope(r, s)
	handler.RegisterRoutes(r)
	return r
}

// ── GET /api/v1/users ──────────────────────────────────────────────────────────

func TestUserHandler_List_Success(t *testing.T) {
	r := setupUserRouterWithScope(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["data"])
}

func TestUserHandler_List_NoScope(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ── GET /api/v1/users/{id} ─────────────────────────────────────────────────────

func TestUserHandler_GetByID_Success(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+testUser.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "user@test.com", resp["email"])
	assert.Equal(t, "Test User", resp["full_name"])
}

func TestUserHandler_GetByID_InvalidUUID(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ── DELETE /api/v1/users/{id} ──────────────────────────────────────────────────

func TestUserHandler_Delete_Success(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+testUser.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUserHandler_Delete_InvalidUUID(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/invalid-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ── PUT /api/v1/users/{id} ─────────────────────────────────────────────────────

func TestUserHandler_Update_Success(t *testing.T) {
	r := setupUserRouter(t)

	body, _ := json.Marshal(map[string]string{
		"full_name": "Updated Name",
		"phone":     "08999999999",
		"email":     "updated@test.com",
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+testUser.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	r := setupUserRouter(t)

	body, _ := json.Marshal(map[string]string{
		"full_name": "Name",
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+uuid.New().String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUserHandler_Update_InvalidBody(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+testUser.ID.String(), bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
