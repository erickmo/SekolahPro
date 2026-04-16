package scope_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Scope Value Object ─────────────────────────────────────────────────────────

func TestScope_IsValid_Success(t *testing.T) {
	s := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	assert.NoError(t, s.IsValid())
}

func TestScope_IsValid_MissingTenant(t *testing.T) {
	s := scope.Scope{
		CompanyID: uuid.New(),
	}
	assert.ErrorIs(t, s.IsValid(), scope.ErrMissingTenant)
}

func TestScope_IsValid_MissingCompany(t *testing.T) {
	s := scope.Scope{
		TenantID: uuid.New(),
	}
	assert.ErrorIs(t, s.IsValid(), scope.ErrMissingCompany)
}

func TestScope_IsValid_BothMissing(t *testing.T) {
	s := scope.Scope{}
	err := s.IsValid()
	assert.Error(t, err)
	// TenantID check pertama
	assert.ErrorIs(t, err, scope.ErrMissingTenant)
}

func TestScope_IsValid_AllNil(t *testing.T) {
	s := scope.Scope{
		TenantID:  uuid.Nil,
		CompanyID: uuid.Nil,
	}
	err := s.IsValid()
	assert.Error(t, err)
}

func TestScope_HasBranch_True(t *testing.T) {
	branchID := uuid.New()
	s := scope.Scope{BranchID: &branchID}
	assert.True(t, s.HasBranch())
}

func TestScope_HasBranch_False_Nil(t *testing.T) {
	s := scope.Scope{}
	assert.False(t, s.HasBranch())
}

func TestScope_HasBranch_False_NilUUID(t *testing.T) {
	nilBranch := uuid.Nil
	s := scope.Scope{BranchID: &nilBranch}
	assert.False(t, s.HasBranch())
}

func TestScope_HasWarehouse_True(t *testing.T) {
	warehouseID := uuid.New()
	s := scope.Scope{WarehouseID: &warehouseID}
	assert.True(t, s.HasWarehouse())
}

func TestScope_HasWarehouse_False_Nil(t *testing.T) {
	s := scope.Scope{}
	assert.False(t, s.HasWarehouse())
}

func TestScope_HasWarehouse_False_NilUUID(t *testing.T) {
	nilWarehouse := uuid.Nil
	s := scope.Scope{WarehouseID: &nilWarehouse}
	assert.False(t, s.HasWarehouse())
}

// ── Context Helpers ────────────────────────────────────────────────────────────

func TestWithScope_ScopeFromContext(t *testing.T) {
	s := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	ctx := scope.WithScope(context.Background(), s)

	got, ok := scope.ScopeFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, s.TenantID, got.TenantID)
	assert.Equal(t, s.CompanyID, got.CompanyID)
}

func TestScopeFromContext_Missing(t *testing.T) {
	_, ok := scope.ScopeFromContext(context.Background())
	assert.False(t, ok)
}

func TestScopeFromContext_NilTenant(t *testing.T) {
	s := scope.Scope{} // TenantID = uuid.Nil
	ctx := scope.WithScope(context.Background(), s)

	_, ok := scope.ScopeFromContext(ctx)
	assert.False(t, ok)
}

func TestMustScopeFromContext_Panic(t *testing.T) {
	assert.Panics(t, func() {
		scope.MustScopeFromContext(context.Background())
	})
}

func TestMustScopeFromContext_Success(t *testing.T) {
	s := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	ctx := scope.WithScope(context.Background(), s)

	got := scope.MustScopeFromContext(ctx)
	assert.Equal(t, s.TenantID, got.TenantID)
}

// ── ConfigScopeResolver (single-tenant mode) ───────────────────────────────────

func TestConfigScopeResolver_Resolve(t *testing.T) {
	fixed := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	resolver := scope.NewConfigScopeResolver(fixed)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got, err := resolver.Resolve(req)
	require.NoError(t, err)
	assert.Equal(t, fixed.TenantID, got.TenantID)
	assert.Equal(t, fixed.CompanyID, got.CompanyID)
}

func TestConfigScopeResolver_Resolve_AlwaysSame(t *testing.T) {
	fixed := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	resolver := scope.NewConfigScopeResolver(fixed)

	req1 := httptest.NewRequest(http.MethodGet, "/a", nil)
	req2 := httptest.NewRequest(http.MethodPost, "/b", nil)

	got1, _ := resolver.Resolve(req1)
	got2, _ := resolver.Resolve(req2)

	assert.Equal(t, got1.TenantID, got2.TenantID)
	assert.Equal(t, got1.CompanyID, got2.CompanyID)
}

// ── JWTScopeResolver (multi-tenant mode) ───────────────────────────────────────

// mockClaims mengimplementasi scopeClaimsReader untuk testing.
type mockClaims struct {
	tenantID    string
	companyID   *string
	branchID    *string
	warehouseID *string
}

func (m *mockClaims) GetTenantID() string     { return m.tenantID }
func (m *mockClaims) GetCompanyID() *string   { return m.companyID }
func (m *mockClaims) GetBranchID() *string    { return m.branchID }
func (m *mockClaims) GetWarehouseID() *string { return m.warehouseID }

func TestJWTScopeResolver_Resolve_FullScope(t *testing.T) {
	tenantID := uuid.New()
	companyID := uuid.New()
	branchID := uuid.New()
	warehouseID := uuid.New()

	tenantStr := tenantID.String()
	companyStr := companyID.String()
	branchStr := branchID.String()
	warehouseStr := warehouseID.String()

	claimsKey := struct{}{}
	ctx := context.WithValue(context.Background(), claimsKey, &mockClaims{
		tenantID:    tenantStr,
		companyID:   &companyStr,
		branchID:    &branchStr,
		warehouseID: &warehouseStr,
	})

	resolver := scope.NewJWTScopeResolver(claimsKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	got, err := resolver.Resolve(req)
	require.NoError(t, err)
	assert.Equal(t, tenantID, got.TenantID)
	assert.Equal(t, companyID, got.CompanyID)
	require.NotNil(t, got.BranchID)
	assert.Equal(t, branchID, *got.BranchID)
	require.NotNil(t, got.WarehouseID)
	assert.Equal(t, warehouseID, *got.WarehouseID)
}

func TestJWTScopeResolver_Resolve_TenantOnly(t *testing.T) {
	tenantID := uuid.New()
	tenantStr := tenantID.String()

	claimsKey := struct{}{}
	ctx := context.WithValue(context.Background(), claimsKey, &mockClaims{
		tenantID: tenantStr,
	})

	resolver := scope.NewJWTScopeResolver(claimsKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	got, err := resolver.Resolve(req)
	require.NoError(t, err)
	assert.Equal(t, tenantID, got.TenantID)
	assert.Equal(t, uuid.Nil, got.CompanyID) // tidak ada company
	assert.Nil(t, got.BranchID)
	assert.Nil(t, got.WarehouseID)
}

func TestJWTScopeResolver_Resolve_NoClaims(t *testing.T) {
	claimsKey := struct{}{}
	resolver := scope.NewJWTScopeResolver(claimsKey)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// context tanpa claims
	_, err := resolver.Resolve(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT claims tidak ditemukan")
}

func TestJWTScopeResolver_Resolve_EmptyTenant(t *testing.T) {
	claimsKey := struct{}{}
	ctx := context.WithValue(context.Background(), claimsKey, &mockClaims{
		tenantID: "",
	})

	resolver := scope.NewJWTScopeResolver(claimsKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	_, err := resolver.Resolve(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id kosong")
}

func TestJWTScopeResolver_Resolve_InvalidTenantUUID(t *testing.T) {
	claimsKey := struct{}{}
	ctx := context.WithValue(context.Background(), claimsKey, &mockClaims{
		tenantID: "not-a-uuid",
	})

	resolver := scope.NewJWTScopeResolver(claimsKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	_, err := resolver.Resolve(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id di JWT bukan UUID valid")
}

func TestJWTScopeResolver_Resolve_InvalidCompanyUUID(t *testing.T) {
	tenantID := uuid.New()
	invalidCompany := "not-a-uuid"

	claimsKey := struct{}{}
	ctx := context.WithValue(context.Background(), claimsKey, &mockClaims{
		tenantID:  tenantID.String(),
		companyID: &invalidCompany,
	})

	resolver := scope.NewJWTScopeResolver(claimsKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	_, err := resolver.Resolve(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "company_id di JWT bukan UUID valid")
}

func TestJWTScopeResolver_Resolve_WrongTypeInContext(t *testing.T) {
	claimsKey := struct{}{}
	ctx := context.WithValue(context.Background(), claimsKey, "not-a-claims-reader")

	resolver := scope.NewJWTScopeResolver(claimsKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	_, err := resolver.Resolve(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak mengimplementasi scopeClaimsReader")
}

func TestJWTScopeResolver_Resolve_InvalidBranchUUID(t *testing.T) {
	tenantID := uuid.New()
	companyID := uuid.New()
	invalidBranch := "not-a-uuid"

	claimsKey := struct{}{}
	ctx := context.WithValue(context.Background(), claimsKey, &mockClaims{
		tenantID:  tenantID.String(),
		companyID: ptrStr(companyID.String()),
		branchID:  &invalidBranch,
	})

	resolver := scope.NewJWTScopeResolver(claimsKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	_, err := resolver.Resolve(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch_id di JWT bukan UUID valid")
}

func TestJWTScopeResolver_Resolve_InvalidWarehouseUUID(t *testing.T) {
	tenantID := uuid.New()
	companyID := uuid.New()
	invalidWarehouse := "not-a-uuid"

	claimsKey := struct{}{}
	ctx := context.WithValue(context.Background(), claimsKey, &mockClaims{
		tenantID:    tenantID.String(),
		companyID:   ptrStr(companyID.String()),
		warehouseID: &invalidWarehouse,
	})

	resolver := scope.NewJWTScopeResolver(claimsKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	_, err := resolver.Resolve(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "warehouse_id di JWT bukan UUID valid")
}

// ── ResolveScope Middleware ─────────────────────────────────────────────────────

func TestResolveScope_Success(t *testing.T) {
	fixed := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	resolver := scope.NewConfigScopeResolver(fixed)

	called := false
	handler := scope.ResolveScope(resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		s, ok := scope.ScopeFromContext(r.Context())
		require.True(t, ok)
		assert.Equal(t, fixed.TenantID, s.TenantID)
		assert.Equal(t, fixed.CompanyID, s.CompanyID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestResolveScope_FailOpen(t *testing.T) {
	// Resolver yang selalu error
	claimsKey := struct{}{}
	resolver := scope.NewJWTScopeResolver(claimsKey) // tidak ada claims di context

	called := false
	handler := scope.ResolveScope(resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// Scope tidak ada di context karena fail-open
		_, ok := scope.ScopeFromContext(r.Context())
		assert.False(t, ok)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called) // request tetap diteruskan
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── RequireScope Middleware ─────────────────────────────────────────────────────

func TestRequireScope_TenantAndCompany_Success(t *testing.T) {
	s := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	ctx := scope.WithScope(context.Background(), s)

	called := false
	handler := scope.RequireScope(scope.LevelTenant, scope.LevelCompany)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireScope_NoScope(t *testing.T) {
	called := false
	handler := scope.RequireScope(scope.LevelTenant)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil) // tanpa scope
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireScope_MissingCompany(t *testing.T) {
	s := scope.Scope{
		TenantID: uuid.New(),
		// CompanyID kosong
	}
	ctx := scope.WithScope(context.Background(), s)

	called := false
	handler := scope.RequireScope(scope.LevelTenant, scope.LevelCompany)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireScope_MissingBranch(t *testing.T) {
	s := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
		// BranchID kosong
	}
	ctx := scope.WithScope(context.Background(), s)

	called := false
	handler := scope.RequireScope(scope.LevelTenant, scope.LevelCompany, scope.LevelBranch)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireScope_BranchPresent(t *testing.T) {
	branchID := uuid.New()
	s := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
		BranchID:  &branchID,
	}
	ctx := scope.WithScope(context.Background(), s)

	called := false
	handler := scope.RequireScope(scope.LevelTenant, scope.LevelCompany, scope.LevelBranch)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireScope_Warehouse(t *testing.T) {
	warehouseID := uuid.New()
	s := scope.Scope{
		TenantID:    uuid.New(),
		CompanyID:   uuid.New(),
		WarehouseID: &warehouseID,
	}
	ctx := scope.WithScope(context.Background(), s)

	called := false
	handler := scope.RequireScope(scope.LevelTenant, scope.LevelCompany, scope.LevelWarehouse)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireScope_MissingWarehouse(t *testing.T) {
	s := scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	ctx := scope.WithScope(context.Background(), s)

	called := false
	handler := scope.RequireScope(scope.LevelTenant, scope.LevelCompany, scope.LevelWarehouse)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ── Error Messages ─────────────────────────────────────────────────────────────

func TestErrorMessages(t *testing.T) {
	assert.Equal(t, "tenant tidak teridentifikasi", scope.ErrMissingTenant.Error())
	assert.Contains(t, scope.ErrMissingCompany.Error(), "company")
	assert.Contains(t, scope.ErrMissingBranch.Error(), "branch")
	assert.Contains(t, scope.ErrMissingWarehouse.Error(), "warehouse")
}

// helper
func ptrStr(s string) *string { return &s }
