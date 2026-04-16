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
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/boilerplate/internal/command/change_password"
	"github.com/yourorg/boilerplate/internal/command/login"
	"github.com/yourorg/boilerplate/internal/domain/user"
	httpdelivery "github.com/yourorg/boilerplate/internal/delivery/http"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/middleware"
)

const testSecret = "test-secret-key-must-be-at-least-32-chars"

func newJWTService() *jwtpkg.Service {
	return jwtpkg.NewService(testSecret, 1, "test-issuer", 7)
}

func hashPwForAuth(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	require.NoError(t, err)
	return string(h)
}

func setupAuthRouter(t *testing.T) (*chi.Mux, *jwtpkg.Service) {
	t.Helper()
	jwtSvc := newJWTService()
	readRepo := &testUserReadRepo{
		byEmail: map[string]*user.User{
			"user@test.com": {
				ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Email:        "user@test.com",
				PasswordHash: hashPwForAuth(t, "password123"),
				IsActive:     true,
				FullName:     "Test User",
			},
		},
		byID: map[uuid.UUID]*user.User{
			uuid.MustParse("00000000-0000-0000-0000-000000000001"): {
				ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Email:        "user@test.com",
				PasswordHash: hashPwForAuth(t, "password123"),
				IsActive:     true,
				FullName:     "Test User",
			},
		},
	}
	writeRepo := &testUserWriteRepo{}

	loginHandler := login.NewHandler(readRepo, writeRepo, jwtSvc)
	changePasswordHandler := change_password.NewHandler(readRepo, writeRepo)

	authHandler := httpdelivery.NewAuthHandler(
		loginHandler,
		changePasswordHandler,
		readRepo,
		jwtSvc,
		nil,
		nil,
	)

	r := chi.NewRouter()
	authHandler.RegisterRoutes(r)

	return r, jwtSvc
}

// ── POST /api/v1/auth/login ────────────────────────────────────────────────────

func TestAuthHandler_Login_Success(t *testing.T) {
	r, _ := setupAuthRouter(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "user@test.com",
		"password": "password123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp["access_token"])
	assert.NotEmpty(t, resp["refresh_token"])
	assert.NotZero(t, resp["expires_at"])
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	r, _ := setupAuthRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Login_EmptyEmail(t *testing.T) {
	r, _ := setupAuthRouter(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "",
		"password": "password123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	r, _ := setupAuthRouter(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "user@test.com",
		"password": "wrongpassword",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	r, _ := setupAuthRouter(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "nonexistent@test.com",
		"password": "password123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ── POST /api/v1/auth/refresh ──────────────────────────────────────────────────

func TestAuthHandler_Refresh_Success(t *testing.T) {
	r, jwtSvc := setupAuthRouter(t)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000099")
	pair, err := jwtSvc.GenerateTokenPair(userID, "user@test.com", "user", tenantID)
	require.NoError(t, err)

	claims, err := jwtSvc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err = json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp["access_token"])
}

func TestAuthHandler_Refresh_NoClaims(t *testing.T) {
	r, _ := setupAuthRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ── POST /api/v1/auth/change-password ──────────────────────────────────────────

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	r, jwtSvc := setupAuthRouter(t)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000099")
	pair, err := jwtSvc.GenerateTokenPair(userID, "user@test.com", "user", tenantID)
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{
		"old_password": "password123",
		"new_password": "newpassword456",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims, err := jwtSvc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err = json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "password berhasil diubah", resp["message"])
}

func TestAuthHandler_ChangePassword_NoAuth(t *testing.T) {
	r, _ := setupAuthRouter(t)

	body, _ := json.Marshal(map[string]string{
		"old_password": "old",
		"new_password": "newpassword123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_ChangePassword_WrongOldPassword(t *testing.T) {
	r, jwtSvc := setupAuthRouter(t)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000099")
	pair, err := jwtSvc.GenerateTokenPair(userID, "user@test.com", "user", tenantID)
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{
		"old_password": "wrongpassword",
		"new_password": "newpassword456",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims, _ := jwtSvc.ValidateClaims(pair.AccessToken)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAuthHandler_ChangePassword_WeakNewPassword(t *testing.T) {
	r, jwtSvc := setupAuthRouter(t)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000099")
	pair, err := jwtSvc.GenerateTokenPair(userID, "user@test.com", "user", tenantID)
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{
		"old_password": "password123",
		"new_password": "short",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims, _ := jwtSvc.ValidateClaims(pair.AccessToken)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
