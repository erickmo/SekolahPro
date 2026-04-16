package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/middleware"
)

const testSecret = "test-secret-key-must-be-at-least-32-chars"

func newTestJWTService() *jwtpkg.Service {
	return jwtpkg.NewService(testSecret, 1, "test-issuer", 7)
}

// ── RequireAuth ────────────────────────────────────────────────────────────────

func TestRequireAuth_ValidToken(t *testing.T) {
	jwtSvc := newTestJWTService()
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := jwtSvc.GenerateTokenPair(userID, "user@test.com", "admin", tenantID)
	require.NoError(t, err)

	called := false
	handler := middleware.RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		claims, ok := middleware.ClaimsFromContext(r.Context())
		require.True(t, ok)
		assert.Equal(t, userID.String(), claims.UserID)
		assert.Equal(t, "user@test.com", claims.Email)
		assert.Equal(t, "admin", claims.Role)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuth_ScopedToken_SetsCompanyID(t *testing.T) {
	jwtSvc := newTestJWTService()
	userID := uuid.New()
	tenantID := uuid.New()
	companyID := uuid.New()

	pair, err := jwtSvc.GenerateScopedTokenPair(userID, "user@test.com", "admin", tenantID, companyID, nil, nil)
	require.NoError(t, err)

	called := false
	handler := middleware.RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		companyIDStr, ok := middleware.CompanyIDFromContext(r.Context())
		require.True(t, ok)
		assert.Equal(t, companyID.String(), companyIDStr)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	jwtSvc := newTestJWTService()

	called := false
	handler := middleware.RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "authorization header required")
}

func TestRequireAuth_InvalidFormat_NoBearer(t *testing.T) {
	jwtSvc := newTestJWTService()

	called := false
	handler := middleware.RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid authorization format")
}

func TestRequireAuth_InvalidFormat_NoSpace(t *testing.T) {
	jwtSvc := newTestJWTService()

	called := false
	handler := middleware.RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	jwtSvc := newTestJWTService()

	called := false
	handler := middleware.RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "token invalid or expired")
}

func TestRequireAuth_CaseInsensitiveBearer(t *testing.T) {
	jwtSvc := newTestJWTService()
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := jwtSvc.GenerateTokenPair(userID, "u@t.com", "user", tenantID)
	require.NoError(t, err)

	called := false
	handler := middleware.RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuth_Phase1Token_NoCompanyID(t *testing.T) {
	jwtSvc := newTestJWTService()
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := jwtSvc.GenerateTokenPair(userID, "u@t.com", "user", tenantID)
	require.NoError(t, err)

	called := false
	handler := middleware.RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, ok := middleware.CompanyIDFromContext(r.Context())
		assert.False(t, ok) // Phase 1 tidak punya company_id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
}

// ── RequireRole ────────────────────────────────────────────────────────────────

func TestRequireRole_Allowed(t *testing.T) {
	claims := &jwtpkg.Claims{
		UserID:   uuid.New().String(),
		Email:    "admin@test.com",
		Role:     "admin",
		TenantID: uuid.New().String(),
	}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyClaims, claims)

	called := false
	handler := middleware.RequireRole("admin", "superadmin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_Denied(t *testing.T) {
	claims := &jwtpkg.Claims{
		UserID:   uuid.New().String(),
		Email:    "teacher@test.com",
		Role:     "teacher",
		TenantID: uuid.New().String(),
	}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyClaims, claims)

	called := false
	handler := middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "forbidden")
}

func TestRequireRole_NoClaims(t *testing.T) {
	called := false
	handler := middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireRole_NilClaims(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.ContextKeyClaims, (*jwtpkg.Claims)(nil))

	called := false
	handler := middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireRole_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.ContextKeyClaims, "not-claims")

	called := false
	handler := middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireRole_SingleRole(t *testing.T) {
	claims := &jwtpkg.Claims{
		UserID: uuid.New().String(),
		Role:   "finance",
	}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyClaims, claims)

	called := false
	handler := middleware.RequireRole("finance")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
}

// ── ClaimsFromContext ──────────────────────────────────────────────────────────

func TestClaimsFromContext_Success(t *testing.T) {
	claims := &jwtpkg.Claims{
		UserID: uuid.New().String(),
		Email:  "test@example.com",
	}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyClaims, claims)

	got, ok := middleware.ClaimsFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, claims, got)
}

func TestClaimsFromContext_Missing(t *testing.T) {
	got, ok := middleware.ClaimsFromContext(context.Background())
	assert.False(t, ok)
	assert.Nil(t, got)
}

// ── CompanyIDFromContext ───────────────────────────────────────────────────────

func TestCompanyIDFromContext_Success(t *testing.T) {
	companyID := uuid.New().String()
	ctx := context.WithValue(context.Background(), middleware.ContextKeyCompanyID, companyID)

	got, ok := middleware.CompanyIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, companyID, got)
}

func TestCompanyIDFromContext_Missing(t *testing.T) {
	got, ok := middleware.CompanyIDFromContext(context.Background())
	assert.False(t, ok)
	assert.Empty(t, got)
}

func TestCompanyIDFromContext_EmptyString(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.ContextKeyCompanyID, "")

	got, ok := middleware.CompanyIDFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, got)
}
