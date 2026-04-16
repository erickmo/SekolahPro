package http_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/boilerplate/internal/domain/role"
	"github.com/yourorg/boilerplate/internal/domain/user"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/middleware"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Named Constants ─────────────────────────────────────────────────────────────

const (
	testJWTSecret = "test-secret-key-must-be-at-least-32-chars"

	testUserEmail    = "user@test.com"
	testUserFullName = "Test User"
	testUserPhone    = "08123456789"

	testRoleName        = "Admin"
	testRoleCode        = "admin"
	testRoleDescription = "Administrator"
)

// ── Shared Test Fixtures ─────────────────────────────────────────────────────────

var (
	testUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

	testTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000099")
	testCompanyID = uuid.MustParse("00000000-0000-0000-0000-000000000100")

	testRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000101")

	testUser = &user.User{
		ID:        testUserID,
		Email:     testUserEmail,
		FullName:  testUserFullName,
		Phone:     testUserPhone,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	testRole = &role.Role{
		ID:          testRoleID,
		TenantID:    testTenantID,
		CompanyID:   testCompanyID,
		Name:        testRoleName,
		Code:        testRoleCode,
		Description: testRoleDescription,
		RoleType:    role.TypeAdmin,
		IsSystem:    true,
		Permissions: []string{"*"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
)

// ── Shared Scope / Router Helpers ───────────────────────────────────────────────

func newTestScope() scope.Scope {
	return scope.Scope{
		TenantID:  testTenantID,
		CompanyID: testCompanyID,
	}
}

func newRouterWithScope(r chi.Router, s scope.Scope) {
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := scope.WithScope(req.Context(), s)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
}

func newRouterWithClaims(r chi.Router, claims *jwtpkg.Claims) {
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), middleware.ContextKeyClaims, claims)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
}

func newJWTService() *jwtpkg.Service {
	return jwtpkg.NewService(testJWTSecret, 1, "test-issuer", 7)
}

func makeTestClaims(t *testing.T) *jwtpkg.Claims {
	t.Helper()
	svc := newJWTService()
	pair, err := svc.GenerateScopedTokenPair(
		testUserID, testUserEmail, "admin",
		testTenantID, testCompanyID, nil, nil,
	)
	require.NoError(t, err)
	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)
	return claims
}

func hashPW(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	require.NoError(t, err)
	return string(h)
}

// ── Shared Mock Repositories ────────────────────────────────────────────────────

type testUserReadRepo struct {
	byEmail map[string]*user.User
	byID    map[uuid.UUID]*user.User
	listErr error
}

func (m *testUserReadRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (m *testUserReadRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (m *testUserReadRepo) List(ctx context.Context, s scope.Scope, limit, offset int) ([]*user.User, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []*user.User
	for _, u := range m.byID {
		result = append(result, u)
	}
	return result, len(result), nil
}

type testUserWriteRepo struct {
	savedUser *user.User
	updated   *user.User
	deletedID *uuid.UUID
	saveErr   error
	updateErr error
	deleteErr error
}

func (m *testUserWriteRepo) Save(ctx context.Context, u *user.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUser = u
	return nil
}

func (m *testUserWriteRepo) Update(ctx context.Context, u *user.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated = u
	return nil
}
func (m *testUserWriteRepo) UpdatePassword(ctx context.Context, id uuid.UUID, h string) error { return nil }
func (m *testUserWriteRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error           { return nil }
func (m *testUserWriteRepo) Deactivate(ctx context.Context, id uuid.UUID) error                { return nil }
func (m *testUserWriteRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedID = &id
	return nil
}

// ── Role Mock Repositories ──────────────────────────────────────────────────────

type testRoleReadRepo struct {
	roles map[uuid.UUID]*role.Role
}

func (m *testRoleReadRepo) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*role.Role, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, role.ErrNotFound
	}
	return r, nil
}

func (m *testRoleReadRepo) GetByCode(ctx context.Context, s scope.Scope, code string) (*role.Role, error) {
	for _, r := range m.roles {
		if r.Code == code {
			return r, nil
		}
	}
	return nil, role.ErrNotFound
}

func (m *testRoleReadRepo) List(ctx context.Context, s scope.Scope) ([]*role.Role, error) {
	var result []*role.Role
	for _, r := range m.roles {
		result = append(result, r)
	}
	return result, nil
}

type testUserRoleReadRepo struct {
	roles []*role.Role
}

func (m *testUserRoleReadRepo) GetByUserAndCompany(ctx context.Context, userID, companyID uuid.UUID) ([]*role.UserRole, error) {
	return nil, nil
}

func (m *testUserRoleReadRepo) GetRolesForUser(ctx context.Context, userID, companyID uuid.UUID) ([]*role.Role, error) {
	return m.roles, nil
}
