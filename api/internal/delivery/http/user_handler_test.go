package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	registercmd "github.com/yourorg/boilerplate/internal/command/register_user"
	"github.com/yourorg/boilerplate/internal/domain/user"
	httpdelivery "github.com/yourorg/boilerplate/internal/delivery/http"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Router Setup Helpers ────────────────────────────────────────────────────────

func setupUserRouter(t *testing.T) *chi.Mux {
	t.Helper()
	return setupUserRouterWithDeps(t, nil, nil)
}

func setupUserRouterWithScope(t *testing.T) *chi.Mux {
	t.Helper()
	return setupUserRouterWithDepsAndScope(t, nil, nil, newTestScope(), nil)
}

func setupUserRouterWithDeps(
	t *testing.T,
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
) *chi.Mux {
	t.Helper()

	readRepo := &testUserReadRepo{
		byID: map[uuid.UUID]*user.User{
			testUser.ID: testUser,
		},
	}
	writeRepo := &testUserWriteRepo{}

	handler := httpdelivery.NewUserHandler(readRepo, writeRepo, cb, qb)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func setupUserRouterWithDepsAndScope(
	t *testing.T,
	cb *commandbus.CommandBus,
	qb *querybus.QueryBus,
	s scope.Scope,
	claims *jwtpkg.Claims,
) *chi.Mux {
	t.Helper()

	readRepo := &testUserReadRepo{
		byID: map[uuid.UUID]*user.User{
			testUser.ID: testUser,
		},
	}
	writeRepo := &testUserWriteRepo{}

	handler := httpdelivery.NewUserHandler(readRepo, writeRepo, cb, qb)

	r := chi.NewRouter()
	newRouterWithScope(r, s)
	if claims != nil {
		newRouterWithClaims(r, claims)
	}
	handler.RegisterRoutes(r)
	return r
}

func newCommandBusWithRegisterStub(
	t *testing.T,
	resultID uuid.UUID,
	stubErr error,
) *commandbus.CommandBus {
	t.Helper()
	cb := commandbus.New()
	commandbus.Register(cb, &stubRegisterHandler{
		resultID: resultID,
		err:      stubErr,
	})
	return cb
}

// ── GET /api/v1/users (List) ────────────────────────────────────────────────────

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
	assert.NotNil(t, resp["total"])
}

func TestUserHandler_List_NoScope(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUserHandler_List_Pagination(t *testing.T) {
	r := setupUserRouterWithScope(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=10&offset=5", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(10), resp["limit"])
	assert.Equal(t, float64(5), resp["offset"])
}

func TestUserHandler_List_ScopeIsolation(t *testing.T) {
	isolatedScope := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}

	readRepo := &testUserReadRepo{
		byID: map[uuid.UUID]*user.User{
			testUser.ID: testUser,
		},
	}
	writeRepo := &testUserWriteRepo{}
	handler := httpdelivery.NewUserHandler(readRepo, writeRepo, nil, nil)

	r := chi.NewRouter()
	newRouterWithScope(r, isolatedScope)
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── POST /api/v1/users (Create) ────────────────────────────────────────────────

func TestUserHandler_Create_MalformedJSON(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_Create_NoHandlerRegistered(t *testing.T) {
	// Empty command bus — no handler registered for register_user command
	cb := commandbus.New()
	r := setupUserRouterWithDeps(t, cb, nil)

	body, _ := json.Marshal(map[string]string{
		"email":     "new@test.com",
		"password":  "password123",
		"full_name": "New User",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// No handler registered for the command
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestUserHandler_Create_ValidUser(t *testing.T) {
	expectedID := uuid.New()
	cb := newCommandBusWithRegisterStub(t, expectedID, nil)

	r := setupUserRouterWithDeps(t, cb, nil)

	body, _ := json.Marshal(map[string]string{
		"email":     "new@test.com",
		"password":  "password123",
		"full_name": "New User",
		"phone":     "08123456789",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, expectedID.String(), resp["id"])
}

func TestUserHandler_Create_DuplicateEmail(t *testing.T) {
	cb := newCommandBusWithRegisterStub(t, uuid.Nil, user.ErrEmailExists)

	r := setupUserRouterWithDeps(t, cb, nil)

	body, _ := json.Marshal(map[string]string{
		"email":     "user@test.com",
		"password":  "password123",
		"full_name": "Duplicate",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestUserHandler_Create_EmptyEmail(t *testing.T) {
	cb := newCommandBusWithRegisterStub(t, uuid.Nil, user.ErrEmailEmpty)

	r := setupUserRouterWithDeps(t, cb, nil)

	body, _ := json.Marshal(map[string]string{
		"email":     "",
		"password":  "password123",
		"full_name": "No Email",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestUserHandler_Create_WeakPassword(t *testing.T) {
	cb := newCommandBusWithRegisterStub(t, uuid.Nil, user.ErrPasswordWeak)

	r := setupUserRouterWithDeps(t, cb, nil)

	body, _ := json.Marshal(map[string]string{
		"email":     "new@test.com",
		"password":  "short",
		"full_name": "Weak PW",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// ── GET /api/v1/users/{id} (GetByID) ───────────────────────────────────────────

func TestUserHandler_GetByID_Success(t *testing.T) {
	r := setupUserRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+testUser.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, testUserEmail, resp["email"])
	assert.Equal(t, testUserFullName, resp["full_name"])
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

// ── PUT /api/v1/users/{id} (Update) ────────────────────────────────────────────

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

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, testUser.ID.String(), resp["id"])
}

func TestUserHandler_Update_InvalidUUID(t *testing.T) {
	r := setupUserRouter(t)

	body, _ := json.Marshal(map[string]string{
		"full_name": "Updated Name",
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/invalid-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
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

// ── DELETE /api/v1/users/{id} (Delete) ─────────────────────────────────────────

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

// ── GET /api/v1/users/me (GetMe) ───────────────────────────────────────────────

func TestUserHandler_GetMe_Success(t *testing.T) {
	claims := makeTestClaims(t)
	s := newTestScope()

	// Create a fresh user for this test to avoid mutation from other tests
	freshUser := &user.User{
		ID:        testUserID,
		Email:     testUserEmail,
		FullName:  testUserFullName,
		Phone:     testUserPhone,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	readRepo := &testUserReadRepo{
		byID: map[uuid.UUID]*user.User{
			testUserID: freshUser,
		},
	}
	writeRepo := &testUserWriteRepo{}
	handler := httpdelivery.NewUserHandler(readRepo, writeRepo, nil, nil)

	r := chi.NewRouter()
	newRouterWithScope(r, s)
	newRouterWithClaims(r, claims)
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, testUserEmail, resp["email"])
	assert.Equal(t, testUserFullName, resp["full_name"])
}

func TestUserHandler_GetMe_NoAuth(t *testing.T) {
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
	// No claims middleware — simulates no auth
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUserHandler_GetMe_UserNotFound(t *testing.T) {
	otherUserID := uuid.New()
	svc := newJWTService()
	pair, err := svc.GenerateScopedTokenPair(
		otherUserID, "other@test.com", "admin",
		testTenantID, testCompanyID, nil, nil,
	)
	require.NoError(t, err)
	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	s := newTestScope()
	r := setupUserRouterWithDepsAndScope(t, nil, nil, s, claims)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ── Stub Command Handlers ───────────────────────────────────────────────────────

// stubRegisterHandler implements commandbus.CommandHandler[registercmd.Command].
type stubRegisterHandler struct {
	resultID uuid.UUID
	err      error
}

func (h *stubRegisterHandler) Handle(ctx context.Context, cmd registercmd.Command) error {
	if h.err != nil {
		return h.err
	}
	commandbus.SetCreatedID(ctx, h.resultID)
	return nil
}
