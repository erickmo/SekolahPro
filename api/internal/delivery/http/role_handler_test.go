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

	"github.com/yourorg/boilerplate/internal/domain/role"
	httpdelivery "github.com/yourorg/boilerplate/internal/delivery/http"
)

func setupRoleRouter(t *testing.T) *chi.Mux {
	t.Helper()

	s := newTestScope()

	roleReadRepo := &testRoleReadRepo{
		roles: map[uuid.UUID]*role.Role{
			testRole.ID: testRole,
		},
	}
	userRoleReadRepo := &testUserRoleReadRepo{
		roles: []*role.Role{testRole},
	}

	handler := httpdelivery.NewRoleHandler(roleReadRepo, userRoleReadRepo, nil, nil)

	r := chi.NewRouter()
	newRouterWithScope(r, s)
	handler.RegisterRoutes(r)
	return r
}

func setupRoleRouterNoScope(t *testing.T) *chi.Mux {
	t.Helper()

	roleReadRepo := &testRoleReadRepo{
		roles: map[uuid.UUID]*role.Role{
			testRole.ID: testRole,
		},
	}
	userRoleReadRepo := &testUserRoleReadRepo{}

	handler := httpdelivery.NewRoleHandler(roleReadRepo, userRoleReadRepo, nil, nil)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

// ── GET /api/v1/roles ──────────────────────────────────────────────────────────

func TestRoleHandler_List_Success(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["data"])
}

func TestRoleHandler_List_NoScope(t *testing.T) {
	r := setupRoleRouterNoScope(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ── GET /api/v1/roles/{id} ─────────────────────────────────────────────────────

func TestRoleHandler_GetByID_Success(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+testRole.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Admin", resp["name"])
	assert.Equal(t, "admin", resp["code"])
}

func TestRoleHandler_GetByID_InvalidUUID(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/invalid-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_GetByID_NotFound(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ── GET /api/v1/users/{id}/roles ───────────────────────────────────────────────

func TestRoleHandler_GetUserRoles_Success(t *testing.T) {
	r := setupRoleRouter(t)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID.String()+"/roles", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["data"])
}

func TestRoleHandler_GetUserRoles_InvalidUserID(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/not-a-uuid/roles", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_GetUserRoles_NoScope(t *testing.T) {
	r := setupRoleRouterNoScope(t)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID.String()+"/roles", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ── POST /api/v1/users/{id}/roles (assign role) ────────────────────────────────

func TestRoleHandler_AssignRole_InvalidUserID(t *testing.T) {
	r := setupRoleRouter(t)

	body, _ := json.Marshal(map[string]string{"role_id": uuid.New().String()})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/not-a-uuid/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_AssignRole_InvalidRoleID(t *testing.T) {
	r := setupRoleRouter(t)

	userID := uuid.New()
	body, _ := json.Marshal(map[string]string{"role_id": "not-a-uuid"})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+userID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_AssignRole_InvalidBody(t *testing.T) {
	r := setupRoleRouter(t)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+userID.String()+"/roles", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ── DELETE /api/v1/roles/{id} ──────────────────────────────────────────────────

func TestRoleHandler_Delete_InvalidUUID(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/invalid-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_Delete_Success(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/"+testRole.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// ── DELETE /api/v1/users/{id}/roles/{roleId} (revoke role) ─────────────────────

func TestRoleHandler_RevokeRole_InvalidUserID(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/not-a-uuid/roles/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_RevokeRole_InvalidRoleID(t *testing.T) {
	r := setupRoleRouter(t)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+userID.String()+"/roles/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_RevokeRole_ValidIDs(t *testing.T) {
	r := setupRoleRouter(t)

	userID := uuid.New()
	roleID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+userID.String()+"/roles/"+roleID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
