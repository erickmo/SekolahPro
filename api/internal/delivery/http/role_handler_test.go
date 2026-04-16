package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assignrolecmd "github.com/yourorg/boilerplate/internal/command/assign_role"
	createrolecmd "github.com/yourorg/boilerplate/internal/command/create_role"
	"github.com/yourorg/boilerplate/internal/domain/role"
	httpdelivery "github.com/yourorg/boilerplate/internal/delivery/http"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Router Setup Helpers ────────────────────────────────────────────────────────

func setupRoleRouter(t *testing.T) *chi.Mux {
	t.Helper()
	return setupRoleRouterWithDepsAndScope(t, nil, nil, newTestScope())
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

func setupRoleRouterWithDepsAndScope(
	t *testing.T,
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
	s scope.Scope,
) *chi.Mux {
	t.Helper()

	roleReadRepo := &testRoleReadRepo{
		roles: map[uuid.UUID]*role.Role{
			testRole.ID: testRole,
		},
	}
	userRoleReadRepo := &testUserRoleReadRepo{
		roles: []*role.Role{testRole},
	}

	handler := httpdelivery.NewRoleHandler(roleReadRepo, userRoleReadRepo, cb, qb)

	r := chi.NewRouter()
	newRouterWithScope(r, s)
	handler.RegisterRoutes(r)
	return r
}

func newCommandBusWithCreateRoleStub(
	t *testing.T,
	resultID uuid.UUID,
	stubErr error,
) *commandbus.CommandBus {
	t.Helper()
	cb := commandbus.New()
	commandbus.Register(cb, &stubCreateRoleHandler{
		resultID: resultID,
		err:      stubErr,
	})
	return cb
}

func newCommandBusWithAssignRoleStub(
	t *testing.T,
	stubErr error,
) *commandbus.CommandBus {
	t.Helper()
	cb := commandbus.New()
	commandbus.Register(cb, &stubAssignRoleHandler{
		err: stubErr,
	})
	return cb
}

// ── GET /api/v1/roles (List) ───────────────────────────────────────────────────

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

// ── GET /api/v1/roles/{id} (GetByID) ───────────────────────────────────────────

func TestRoleHandler_GetByID_Success(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+testRole.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, testRoleName, resp["name"])
	assert.Equal(t, testRoleCode, resp["code"])
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

func TestRoleHandler_GetByID_NoScope(t *testing.T) {
	r := setupRoleRouterNoScope(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+testRole.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ── POST /api/v1/roles (Create) ────────────────────────────────────────────────

func TestRoleHandler_Create_MalformedJSON(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_Create_NoHandlerRegistered(t *testing.T) {
	// Empty command bus — no handler registered for create_role command
	cb := commandbus.New()
	r := setupRoleRouterWithDepsAndScope(t, cb, nil, newTestScope())

	body, _ := json.Marshal(map[string]interface{}{
		"name":        "Teacher",
		"code":        "teacher",
		"description": "Teacher role",
		"role_type":   "teacher",
		"permissions": []string{"read"},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// No handler registered for the command
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestRoleHandler_Create_ValidRole(t *testing.T) {
	expectedID := uuid.New()
	cb := newCommandBusWithCreateRoleStub(t, expectedID, nil)

	r := setupRoleRouterWithDepsAndScope(t, cb, nil, newTestScope())

	body, _ := json.Marshal(map[string]interface{}{
		"name":        "Teacher",
		"code":        "teacher",
		"description": "Teacher role",
		"role_type":   "teacher",
		"permissions": []string{"read"},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, expectedID.String(), resp["id"])
}

func TestRoleHandler_Create_DuplicateCode(t *testing.T) {
	cb := newCommandBusWithCreateRoleStub(t, uuid.Nil, role.ErrCodeExists)

	r := setupRoleRouterWithDepsAndScope(t, cb, nil, newTestScope())

	body, _ := json.Marshal(map[string]interface{}{
		"name":        "Admin2",
		"code":        "admin",
		"description": "Duplicate code",
		"role_type":   "admin",
		"permissions": []string{"*"},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestRoleHandler_Create_EmptyName(t *testing.T) {
	cb := newCommandBusWithCreateRoleStub(t, uuid.Nil, role.ErrNameEmpty)

	r := setupRoleRouterWithDepsAndScope(t, cb, nil, newTestScope())

	body, _ := json.Marshal(map[string]interface{}{
		"name":        "",
		"code":        "somecode",
		"role_type":   "staff",
		"permissions": []string{},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// ── PUT /api/v1/roles/{id} (Update) ────────────────────────────────────────────

func TestRoleHandler_Update_InvalidUUID(t *testing.T) {
	r := setupRoleRouter(t)

	body, _ := json.Marshal(map[string]interface{}{
		"name":        "Updated Role",
		"description": "Updated description",
		"permissions": []string{"read", "write"},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/roles/invalid-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_Update_MalformedJSON(t *testing.T) {
	r := setupRoleRouter(t)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/roles/"+testRole.ID.String(), bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoleHandler_Update_Success(t *testing.T) {
	r := setupRoleRouter(t)

	body, _ := json.Marshal(map[string]interface{}{
		"name":        "Updated Admin",
		"description": "Updated admin role",
		"role_type":   "admin",
		"permissions": []string{"*"},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/roles/"+testRole.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, testRole.ID.String(), resp["id"])
}

// ── DELETE /api/v1/roles/{id} (Delete) ─────────────────────────────────────────

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

// ── GET /api/v1/users/{id}/roles (GetUserRoles) ────────────────────────────────

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

// ── POST /api/v1/users/{id}/roles (AssignRole) ─────────────────────────────────

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

func TestRoleHandler_AssignRole_Success(t *testing.T) {
	cb := newCommandBusWithAssignRoleStub(t, nil)

	r := setupRoleRouterWithDepsAndScope(t, cb, nil, newTestScope())

	userID := uuid.New()
	roleID := uuid.New()
	body, _ := json.Marshal(map[string]string{"role_id": roleID.String()})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+userID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "role berhasil ditetapkan", resp["message"])
}

func TestRoleHandler_AssignRole_AlreadyAssigned(t *testing.T) {
	cb := newCommandBusWithAssignRoleStub(t, role.ErrAlreadyAssigned)

	r := setupRoleRouterWithDepsAndScope(t, cb, nil, newTestScope())

	userID := uuid.New()
	roleID := uuid.New()
	body, _ := json.Marshal(map[string]string{"role_id": roleID.String()})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+userID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// ── DELETE /api/v1/users/{id}/roles/{roleId} (RevokeRole) ──────────────────────

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

// ── Stub Command Handlers ───────────────────────────────────────────────────────

// stubCreateRoleHandler implements commandbus.CommandHandler[createrolecmd.Command].
type stubCreateRoleHandler struct {
	resultID uuid.UUID
	err      error
}

func (h *stubCreateRoleHandler) Handle(ctx context.Context, cmd createrolecmd.Command) error {
	if h.err != nil {
		return h.err
	}
	commandbus.SetCreatedID(ctx, h.resultID)
	return nil
}

// stubAssignRoleHandler implements commandbus.CommandHandler[assignrolecmd.Command].
type stubAssignRoleHandler struct {
	err error
}

func (h *stubAssignRoleHandler) Handle(ctx context.Context, cmd assignrolecmd.Command) error {
	return h.err
}
