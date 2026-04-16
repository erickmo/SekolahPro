package jwt_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/pkg/jwt"
)

const testSecret = "test-secret-key-must-be-at-least-32-chars"

func newTestService() *jwt.Service {
	return jwt.NewService(testSecret, 1, "test-issuer", 7)
}

func TestNewService(t *testing.T) {
	svc := jwt.NewService("secret", 24, "myapp", 7)
	assert.NotNil(t, svc)
}

// ── GenerateTokenPair (Phase 1) ────────────────────────────────────────────────

func TestGenerateTokenPair_Phase1(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, "user@test.com", "admin", tenantID)
	require.NoError(t, err)
	require.NotNil(t, pair)

	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotZero(t, pair.ExpiresAt)

	// Validate access token contains Phase 1 claims
	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, "user@test.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, tenantID.String(), claims.TenantID)

	// Phase 1: CompanyID, BranchID, WarehouseID harus kosong
	assert.Nil(t, claims.CompanyID)
	assert.Nil(t, claims.BranchID)
	assert.Nil(t, claims.WarehouseID)
}

func TestGenerateTokenPair_Issuer(t *testing.T) {
	svc := jwt.NewService(testSecret, 1, "my-test-issuer", 7)
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, "a@b.com", "user", tenantID)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "my-test-issuer", claims.Issuer)
}

// ── GenerateScopedTokenPair (Phase 2) ──────────────────────────────────────────

func TestGenerateScopedTokenPair_FullScope(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()
	tenantID := uuid.New()
	companyID := uuid.New()
	branchID := uuid.New()
	warehouseID := uuid.New()

	pair, err := svc.GenerateScopedTokenPair(userID, "user@test.com", "teacher", tenantID, companyID, &branchID, &warehouseID)
	require.NoError(t, err)
	require.NotNil(t, pair)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, tenantID.String(), claims.TenantID)
	require.NotNil(t, claims.CompanyID)
	assert.Equal(t, companyID.String(), *claims.CompanyID)
	require.NotNil(t, claims.BranchID)
	assert.Equal(t, branchID.String(), *claims.BranchID)
	require.NotNil(t, claims.WarehouseID)
	assert.Equal(t, warehouseID.String(), *claims.WarehouseID)
}

func TestGenerateScopedTokenPair_CompanyOnly(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()
	tenantID := uuid.New()
	companyID := uuid.New()

	pair, err := svc.GenerateScopedTokenPair(userID, "user@test.com", "staff", tenantID, companyID, nil, nil)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	require.NotNil(t, claims.CompanyID)
	assert.Equal(t, companyID.String(), *claims.CompanyID)
	assert.Nil(t, claims.BranchID)
	assert.Nil(t, claims.WarehouseID)
}

func TestGenerateScopedTokenPair_NilUUIDBranchAndWarehouse(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()
	tenantID := uuid.New()
	companyID := uuid.New()
	nilBranch := uuid.Nil
	nilWarehouse := uuid.Nil

	pair, err := svc.GenerateScopedTokenPair(userID, "u@t.com", "admin", tenantID, companyID, &nilBranch, &nilWarehouse)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	// Nil UUID tidak boleh di-set ke token
	assert.Nil(t, claims.BranchID)
	assert.Nil(t, claims.WarehouseID)
}

// ── ValidateClaims ─────────────────────────────────────────────────────────────

func TestValidateClaims_ValidToken(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, "valid@test.com", "user", tenantID)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "valid@test.com", claims.Email)
}

func TestValidateClaims_ExpiredToken(t *testing.T) {
	// Buat service dengan expiry 0 jam (langsung expired)
	svc := jwt.NewService(testSecret, -1, "test-issuer", 7)
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, "expired@test.com", "user", tenantID)
	require.NoError(t, err)

	_, err = svc.ValidateClaims(pair.AccessToken)
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
}

func TestValidateClaims_InvalidTokenString(t *testing.T) {
	svc := newTestService()

	_, err := svc.ValidateClaims("this.is.not.a.valid.token")
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
}

func TestValidateClaims_EmptyToken(t *testing.T) {
	svc := newTestService()

	_, err := svc.ValidateClaims("")
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
}

func TestValidateClaims_WrongSecret(t *testing.T) {
	svc1 := jwt.NewService(testSecret, 1, "issuer", 7)
	svc2 := jwt.NewService("different-secret-at-least-32-chars-long", 1, "issuer", 7)

	userID := uuid.New()
	tenantID := uuid.New()
	pair, err := svc1.GenerateTokenPair(userID, "u@t.com", "user", tenantID)
	require.NoError(t, err)

	_, err = svc2.ValidateClaims(pair.AccessToken)
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
}

func TestValidateClaims_TamperedToken(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, "u@t.com", "user", tenantID)
	require.NoError(t, err)

	// Tamper dengan mengubah satu karakter
	tampered := pair.AccessToken[:len(pair.AccessToken)-5] + "XXXXX"
	_, err = svc.ValidateClaims(tampered)
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
}

// ── Refresh Token ──────────────────────────────────────────────────────────────

func TestRefreshToken_HasLongerExpiry(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()
	tenantID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, "u@t.com", "user", tenantID)
	require.NoError(t, err)

	accessClaims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	refreshClaims, err := svc.ValidateClaims(pair.RefreshToken)
	require.NoError(t, err)

	// Refresh token harus expire lebih lambat dari access token
	assert.True(t, refreshClaims.ExpiresAt.After(accessClaims.ExpiresAt.Time))
}

// ── Claims Interface Methods ───────────────────────────────────────────────────

func TestClaims_GetTenantID(t *testing.T) {
	c := &jwt.Claims{TenantID: "test-tenant-id"}
	assert.Equal(t, "test-tenant-id", c.GetTenantID())
}

func TestClaims_GetCompanyID(t *testing.T) {
	companyID := "test-company-id"
	c := &jwt.Claims{CompanyID: &companyID}
	require.NotNil(t, c.GetCompanyID())
	assert.Equal(t, "test-company-id", *c.GetCompanyID())
}

func TestClaims_GetCompanyID_Nil(t *testing.T) {
	c := &jwt.Claims{}
	assert.Nil(t, c.GetCompanyID())
}

func TestClaims_GetBranchID(t *testing.T) {
	branchID := "test-branch-id"
	c := &jwt.Claims{BranchID: &branchID}
	require.NotNil(t, c.GetBranchID())
	assert.Equal(t, "test-branch-id", *c.GetBranchID())
}

func TestClaims_GetWarehouseID(t *testing.T) {
	warehouseID := "test-warehouse-id"
	c := &jwt.Claims{WarehouseID: &warehouseID}
	require.NotNil(t, c.GetWarehouseID())
	assert.Equal(t, "test-warehouse-id", *c.GetWarehouseID())
}

// ── TokenPair ──────────────────────────────────────────────────────────────────

func TestTokenPair_Fields(t *testing.T) {
	pair := &jwt.TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Unix(),
	}
	assert.Equal(t, "access", pair.AccessToken)
	assert.Equal(t, "refresh", pair.RefreshToken)
	assert.NotZero(t, pair.ExpiresAt)
}

func TestErrTokenInvalid(t *testing.T) {
	assert.Equal(t, "token invalid or expired", jwt.ErrTokenInvalid.Error())
}
