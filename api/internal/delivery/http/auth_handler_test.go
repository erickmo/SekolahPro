package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/command/change_password"
	"github.com/yourorg/boilerplate/internal/command/login"
	"github.com/yourorg/boilerplate/internal/domain/user"
	httpdelivery "github.com/yourorg/boilerplate/internal/delivery/http"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/middleware"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Auth-specific Constants ──────────────────────────────────────────────────

const (
	validPassword = "password123"
	wrongPassword = "wrongpassword"
	weakPassword  = "short"

	pathLogin          = "/api/v1/auth/login"
	pathRefresh        = "/api/v1/auth/refresh"
	pathChangePassword = "/api/v1/auth/change-password"

	contentTypeJSON = "application/json"
)

// ── Auth Helpers ─────────────────────────────────────────────────────────────

func setupAuthRouter(t *testing.T) (*chi.Mux, *jwtpkg.Service) {
	t.Helper()
	jwtSvc := newJWTService()
	readRepo := &testUserReadRepo{
		byEmail: map[string]*user.User{
			testUserEmail: {
				ID:           testUserID,
				Email:        testUserEmail,
				PasswordHash: hashPW(t, validPassword),
				IsActive:     true,
				FullName:     testUserFullName,
			},
			"inactive@test.com": {
				ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Email:        "inactive@test.com",
				PasswordHash: hashPW(t, validPassword),
				IsActive:     false,
				FullName:     "Inactive User",
			},
		},
		byID: map[uuid.UUID]*user.User{
			testUserID: {
				ID:           testUserID,
				Email:        testUserEmail,
				PasswordHash: hashPW(t, validPassword),
				IsActive:     true,
				FullName:     testUserFullName,
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

func makeAuthRequest(method, path string, body any) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", contentTypeJSON)
	return req
}

func requestWithClaims(req *http.Request, claims *jwtpkg.Claims) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
	return req.WithContext(ctx)
}

func generateLoginClaims(jwtSvc *jwtpkg.Service) *jwtpkg.Claims {
	pair, err := jwtSvc.GenerateTokenPair(testUserID, testUserEmail, "user", testTenantID)
	if err != nil {
		panic(fmt.Sprintf("generate test token: %v", err))
	}
	claims, err := jwtSvc.ValidateClaims(pair.AccessToken)
	if err != nil {
		panic(fmt.Sprintf("validate test token: %v", err))
	}
	return claims
}

func decodeJSONAny(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	return resp
}

func decodeJSONStr(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	return resp
}

// ── POST /api/v1/auth/login ──────────────────────────────────────────────────

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "valid credentials return 200 with tokens",
			body:       map[string]string{"email": testUserEmail, "password": validPassword},
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty email returns 400",
			body:       map[string]string{"email": "", "password": validPassword},
			wantStatus: http.StatusBadRequest,
			wantErr:    "wajib diisi",
		},
		{
			name:       "empty password returns 400",
			body:       map[string]string{"email": testUserEmail, "password": ""},
			wantStatus: http.StatusBadRequest,
			wantErr:    "wajib diisi",
		},
		{
			name:       "wrong password returns 401",
			body:       map[string]string{"email": testUserEmail, "password": wrongPassword},
			wantStatus: http.StatusUnauthorized,
			wantErr:    "email atau password salah",
		},
		{
			name:       "nonexistent email returns 401",
			body:       map[string]string{"email": "nobody@test.com", "password": validPassword},
			wantStatus: http.StatusUnauthorized,
			wantErr:    "email atau password salah",
		},
		{
			name:       "inactive user returns 401",
			body:       map[string]string{"email": "inactive@test.com", "password": validPassword},
			wantStatus: http.StatusUnauthorized,
			wantErr:    "email atau password salah",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := setupAuthRouter(t)
			req := makeAuthRequest(http.MethodPost, pathLogin, tc.body)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				resp := decodeJSONAny(t, rec)
				assert.NotEmpty(t, resp["access_token"])
				assert.NotEmpty(t, resp["refresh_token"])
				assert.NotZero(t, resp["expires_at"])
			}
			if tc.wantErr != "" {
				resp := decodeJSONStr(t, rec)
				assert.Contains(t, resp["error"], tc.wantErr)
			}
		})
	}
}

func TestAuthHandler_Login_MalformedJSON(t *testing.T) {
	r, _ := setupAuthRouter(t)
	req := httptest.NewRequest(http.MethodPost, pathLogin, bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", contentTypeJSON)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	resp := decodeJSONStr(t, rec)
	assert.Contains(t, resp["error"], "tidak valid")
}

// ── POST /api/v1/auth/refresh ────────────────────────────────────────────────

func TestAuthHandler_Refresh(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func(*jwtpkg.Service) *jwtpkg.Claims
		wantStatus int
		wantErr    string
	}{
		{
			name:       "valid refresh token returns new tokens",
			setupCtx:   func(svc *jwtpkg.Service) *jwtpkg.Claims { return generateLoginClaims(svc) },
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing claims returns 401",
			setupCtx:   func(_ *jwtpkg.Service) *jwtpkg.Claims { return nil },
			wantStatus: http.StatusUnauthorized,
			wantErr:    "tidak valid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, jwtSvc := setupAuthRouter(t)
			claims := tc.setupCtx(jwtSvc)

			req := makeAuthRequest(http.MethodPost, pathRefresh, nil)
			if claims != nil {
				req = requestWithClaims(req, claims)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				resp := decodeJSONAny(t, rec)
				assert.NotEmpty(t, resp["access_token"])
				assert.NotEmpty(t, resp["refresh_token"])
				assert.NotZero(t, resp["expires_at"])
			}
			if tc.wantErr != "" {
				resp := decodeJSONStr(t, rec)
				assert.Contains(t, resp["error"], tc.wantErr)
			}
		})
	}
}

func TestAuthHandler_Refresh_InvalidUserIDInClaims(t *testing.T) {
	r, _ := setupAuthRouter(t)

	badClaims := &jwtpkg.Claims{
		UserID:   "not-a-uuid",
		Email:    testUserEmail,
		Role:     "user",
		TenantID: testTenantID.String(),
	}

	req := makeAuthRequest(http.MethodPost, pathRefresh, nil)
	req = requestWithClaims(req, badClaims)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ── POST /api/v1/auth/change-password ────────────────────────────────────────

func TestAuthHandler_ChangePassword(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]string
		noAuth     bool
		wantStatus int
		wantErr    string
	}{
		{
			name:       "valid change returns 200",
			body:       map[string]string{"old_password": validPassword, "new_password": "newpassword456"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "wrong old password returns 422",
			body:       map[string]string{"old_password": wrongPassword, "new_password": "newpassword456"},
			wantStatus: http.StatusUnprocessableEntity,
			wantErr:    "password salah",
		},
		{
			name:       "weak new password returns 400",
			body:       map[string]string{"old_password": validPassword, "new_password": weakPassword},
			wantStatus: http.StatusBadRequest,
			wantErr:    "minimal 8 karakter",
		},
		{
			name:       "no auth returns 401",
			body:       map[string]string{"old_password": validPassword, "new_password": "newpassword456"},
			noAuth:     true,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, jwtSvc := setupAuthRouter(t)
			req := makeAuthRequest(http.MethodPost, pathChangePassword, tc.body)

			if !tc.noAuth {
				claims := generateLoginClaims(jwtSvc)
				req = requestWithClaims(req, claims)
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				resp := decodeJSONStr(t, rec)
				assert.Equal(t, "password berhasil diubah", resp["message"])
			}
			if tc.wantErr != "" {
				resp := decodeJSONStr(t, rec)
				assert.Contains(t, resp["error"], tc.wantErr)
			}
		})
	}
}

func TestAuthHandler_ChangePassword_MalformedJSON(t *testing.T) {
	r, jwtSvc := setupAuthRouter(t)
	req := httptest.NewRequest(http.MethodPost, pathChangePassword, bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", contentTypeJSON)
	req = requestWithClaims(req, generateLoginClaims(jwtSvc))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_ChangePassword_EmptyOldPassword(t *testing.T) {
	r, jwtSvc := setupAuthRouter(t)
	body := map[string]string{"old_password": "", "new_password": "newpassword456"}
	req := makeAuthRequest(http.MethodPost, pathChangePassword, body)
	req = requestWithClaims(req, generateLoginClaims(jwtSvc))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ── JWT Claims Validation ────────────────────────────────────────────────────

func TestJWTClaims_TenantIDPresent(t *testing.T) {
	svc := newJWTService()
	pair, err := svc.GenerateTokenPair(testUserID, testUserEmail, "user", testTenantID)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, testTenantID.String(), claims.TenantID)
	assert.Equal(t, testUserID.String(), claims.UserID)
	assert.Equal(t, testUserEmail, claims.Email)
}

func TestJWTClaims_ScopedTokenContainsOrgIDs(t *testing.T) {
	svc := newJWTService()
	branchID := uuid.MustParse("00000000-0000-0000-0000-000000000200")

	pair, err := svc.GenerateScopedTokenPair(
		testUserID, testUserEmail, "admin",
		testTenantID, testCompanyID, &branchID, nil,
	)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, testTenantID.String(), claims.TenantID)
	require.NotNil(t, claims.CompanyID)
	assert.Equal(t, testCompanyID.String(), *claims.CompanyID)
	require.NotNil(t, claims.BranchID)
	assert.Equal(t, branchID.String(), *claims.BranchID)
	assert.Nil(t, claims.WarehouseID)
}

func TestJWTClaims_ExpiredTokenRejected(t *testing.T) {
	svc := jwtpkg.NewService(testJWTSecret, -1, "test-issuer", 7)
	pair, err := svc.GenerateTokenPair(testUserID, testUserEmail, "user", testTenantID)
	require.NoError(t, err)

	_, err = svc.ValidateClaims(pair.AccessToken)
	assert.ErrorIs(t, err, jwtpkg.ErrTokenInvalid)
}

func TestJWTClaims_InvalidSignatureRejected(t *testing.T) {
	goodSvc := newJWTService()
	pair, err := goodSvc.GenerateTokenPair(testUserID, testUserEmail, "user", testTenantID)
	require.NoError(t, err)

	badSvc := jwtpkg.NewService("wrong-secret-at-least-32-characters-", 1, "test-issuer", 7)
	_, err = badSvc.ValidateClaims(pair.AccessToken)
	assert.ErrorIs(t, err, jwtpkg.ErrTokenInvalid)
}

// ── RequireAuth Middleware ────────────────────────────────────────────────────

func TestRequireAuth_MissingHeader(t *testing.T) {
	svc := newJWTService()
	handler := middleware.RequireAuth(svc)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	svc := newJWTService()
	handler := middleware.RequireAuth(svc)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAuth_ValidToken(t *testing.T) {
	svc := newJWTService()
	pair, err := svc.GenerateTokenPair(testUserID, testUserEmail, "user", testTenantID)
	require.NoError(t, err)

	var gotClaims *jwtpkg.Claims
	handler := middleware.RequireAuth(svc)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotClaims, _ = middleware.ClaimsFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, gotClaims)
	assert.Equal(t, testUserID.String(), gotClaims.UserID)
	assert.Equal(t, testUserEmail, gotClaims.Email)
}

// ── RequireRole Middleware ────────────────────────────────────────────────────

func TestRequireRole_Allowed(t *testing.T) {
	handler := middleware.RequireRole("admin", "user")(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	claims := &jwtpkg.Claims{UserID: testUserID.String(), Role: "admin"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_Forbidden(t *testing.T) {
	handler := middleware.RequireRole("admin")(
		http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("should not reach handler")
		}),
	)

	claims := &jwtpkg.Claims{UserID: testUserID.String(), Role: "viewer"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRole_NoClaims(t *testing.T) {
	handler := middleware.RequireRole("admin")(
		http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("should not reach handler")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ── Scope Middleware ──────────────────────────────────────────────────────────

func TestScope_ConfigResolver_SingleMode(t *testing.T) {
	fixed := scope.Scope{
		TenantID:  testTenantID,
		CompanyID: testCompanyID,
	}
	resolver := scope.NewConfigScopeResolver(fixed)

	got, err := resolver.Resolve(nil)
	require.NoError(t, err)
	assert.Equal(t, fixed.TenantID, got.TenantID)
	assert.Equal(t, fixed.CompanyID, got.CompanyID)
}

func TestScope_JWTResolver_MultiMode(t *testing.T) {
	svc := newJWTService()
	pair, err := svc.GenerateScopedTokenPair(
		testUserID, testUserEmail, "user",
		testTenantID, testCompanyID, nil, nil,
	)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	resolver := scope.NewJWTScopeResolver(middleware.ContextKeyClaims)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
	req = req.WithContext(ctx)

	got, err := resolver.Resolve(req)
	require.NoError(t, err)
	assert.Equal(t, testTenantID, got.TenantID)
	assert.Equal(t, testCompanyID, got.CompanyID)
}

func TestScope_JWTResolver_MissingClaims(t *testing.T) {
	resolver := scope.NewJWTScopeResolver(middleware.ContextKeyClaims)
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := resolver.Resolve(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak ditemukan")
}

func TestScope_ResolveScope_Middleware(t *testing.T) {
	fixed := scope.Scope{
		TenantID:  testTenantID,
		CompanyID: testCompanyID,
	}
	resolver := scope.NewConfigScopeResolver(fixed)

	var gotScope scope.Scope
	handler := scope.ResolveScope(resolver)(
		http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			gotScope, _ = scope.ScopeFromContext(r.Context())
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, fixed.TenantID, gotScope.TenantID)
}

func TestScope_RequireScope_MissingScope_Returns403(t *testing.T) {
	handler := scope.RequireScope(scope.LevelTenant, scope.LevelCompany)(
		http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("should not reach handler")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestScope_RequireScope_ValidScope_Passes(t *testing.T) {
	s := scope.Scope{
		TenantID:  testTenantID,
		CompanyID: testCompanyID,
	}

	handler := scope.RequireScope(scope.LevelTenant, scope.LevelCompany)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := scope.WithScope(req.Context(), s)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── JWT Refresh Token Is Longer-Lived ────────────────────────────────────────

func TestJWT_RefreshTokenLongerLived(t *testing.T) {
	svc := newJWTService()
	pair, err := svc.GenerateTokenPair(testUserID, testUserEmail, "user", testTenantID)
	require.NoError(t, err)

	accessClaims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	refreshClaims, err := svc.ValidateClaims(pair.RefreshToken)
	require.NoError(t, err)

	assert.True(t, refreshClaims.ExpiresAt.After(accessClaims.ExpiresAt.Time),
		"refresh token should expire after access token")
}

// ── JWT Manual Expired Token ─────────────────────────────────────────────────

func TestJWT_ManualExpiredToken(t *testing.T) {
	claims := &jwtpkg.Claims{
		UserID:   testUserID.String(),
		Email:    testUserEmail,
		Role:     "user",
		TenantID: testTenantID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Hour)),
			Issuer:    "test-issuer",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)

	svc := newJWTService()
	_, err = svc.ValidateClaims(tokenStr)
	assert.ErrorIs(t, err, jwtpkg.ErrTokenInvalid)
}

// ── Scope IsValid ────────────────────────────────────────────────────────────

func TestScope_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		scope   scope.Scope
		wantErr error
	}{
		{
			name:    "valid scope returns nil",
			scope:   scope.Scope{TenantID: testTenantID, CompanyID: testCompanyID},
			wantErr: nil,
		},
		{
			name:    "missing tenant returns error",
			scope:   scope.Scope{CompanyID: testCompanyID},
			wantErr: scope.ErrMissingTenant,
		},
		{
			name:    "missing company returns error",
			scope:   scope.Scope{TenantID: testTenantID},
			wantErr: scope.ErrMissingCompany,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.scope.IsValid()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── Scope HasBranch / HasWarehouse ───────────────────────────────────────────

func TestScope_HasBranch(t *testing.T) {
	branchID := uuid.MustParse("00000000-0000-0000-0000-000000000200")

	assert.False(t, scope.Scope{}.HasBranch(), "nil branch should be false")
	assert.False(t, scope.Scope{BranchID: ptrUUID(uuid.Nil)}.HasBranch(), "nil UUID should be false")
	assert.True(t, scope.Scope{BranchID: &branchID}.HasBranch(), "valid branch should be true")
}

func TestScope_HasWarehouse(t *testing.T) {
	whID := uuid.MustParse("00000000-0000-0000-0000-000000000300")

	assert.False(t, scope.Scope{}.HasWarehouse())
	assert.True(t, scope.Scope{WarehouseID: &whID}.HasWarehouse())
}

// ptrUUID returns a pointer to the given uuid.UUID.
func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
