package assign_role_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/role"
	assignrolecmd "github.com/yourorg/boilerplate/internal/command/assign_role"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Mocks ──────────────────────────────────────────────────────────────────────

type mockUserRoleWriteRepo struct {
	assigned *role.UserRole
	assignFn func(ctx context.Context, ur *role.UserRole) error
}

func (m *mockUserRoleWriteRepo) Assign(ctx context.Context, ur *role.UserRole) error {
	if m.assignFn != nil {
		return m.assignFn(ctx, ur)
	}
	m.assigned = ur
	return nil
}

func (m *mockUserRoleWriteRepo) Revoke(ctx context.Context, s scope.Scope, userID, roleID uuid.UUID) error {
	return nil
}

type mockUserRoleReadRepo struct {
	existingRoles []*role.UserRole
}

func (m *mockUserRoleReadRepo) GetByUserAndCompany(ctx context.Context, userID, companyID uuid.UUID) ([]*role.UserRole, error) {
	return m.existingRoles, nil
}

func (m *mockUserRoleReadRepo) GetRolesForUser(ctx context.Context, userID, companyID uuid.UUID) ([]*role.Role, error) {
	return nil, nil
}

func newTestScope() scope.Scope {
	return scope.Scope{
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
}

func newTestEventBus(t *testing.T) eventbus.EventBus {
	t.Helper()
	eb, err := eventbus.NewInMemoryEventBus()
	require.NoError(t, err)
	return eb
}

// ── Assign Role Success ────────────────────────────────────────────────────────

func TestHandle_Success(t *testing.T) {
	s := newTestScope()
	userID := uuid.New()
	roleID := uuid.New()

	writeRepo := &mockUserRoleWriteRepo{}
	readRepo := &mockUserRoleReadRepo{}
	handler := assignrolecmd.NewHandler(writeRepo, readRepo, newTestEventBus(t))

	ctx := scope.WithScope(context.Background(), s)

	err := handler.Handle(ctx, assignrolecmd.Command{
		UserID: userID,
		RoleID: roleID,
	})
	require.NoError(t, err)

	// Verify assignment was created
	require.NotNil(t, writeRepo.assigned)
	assert.Equal(t, userID, writeRepo.assigned.UserID)
	assert.Equal(t, roleID, writeRepo.assigned.RoleID)
	assert.Equal(t, s.TenantID, writeRepo.assigned.TenantID)
	assert.Equal(t, s.CompanyID, writeRepo.assigned.CompanyID)
	assert.True(t, writeRepo.assigned.IsActive)
	assert.NotEqual(t, uuid.Nil, writeRepo.assigned.ID)
}

// ── Scope Validation ───────────────────────────────────────────────────────────

func TestHandle_NoScope(t *testing.T) {
	handler := assignrolecmd.NewHandler(&mockUserRoleWriteRepo{}, &mockUserRoleReadRepo{}, newTestEventBus(t))

	err := handler.Handle(context.Background(), assignrolecmd.Command{
		UserID: uuid.New(),
		RoleID: uuid.New(),
	})
	assert.ErrorIs(t, err, scope.ErrMissingTenant)
}

func TestHandle_MissingCompany(t *testing.T) {
	s := scope.Scope{TenantID: uuid.New()}
	ctx := scope.WithScope(context.Background(), s)
	handler := assignrolecmd.NewHandler(&mockUserRoleWriteRepo{}, &mockUserRoleReadRepo{}, newTestEventBus(t))

	err := handler.Handle(ctx, assignrolecmd.Command{
		UserID: uuid.New(),
		RoleID: uuid.New(),
	})
	assert.ErrorIs(t, err, scope.ErrMissingCompany)
}

// ── Validation Errors ──────────────────────────────────────────────────────────

func TestHandle_NilUserID(t *testing.T) {
	s := newTestScope()
	ctx := scope.WithScope(context.Background(), s)
	handler := assignrolecmd.NewHandler(&mockUserRoleWriteRepo{}, &mockUserRoleReadRepo{}, newTestEventBus(t))

	err := handler.Handle(ctx, assignrolecmd.Command{
		UserID: uuid.Nil,
		RoleID: uuid.New(),
	})
	assert.ErrorIs(t, err, role.ErrNotFound)
}

func TestHandle_NilRoleID(t *testing.T) {
	s := newTestScope()
	ctx := scope.WithScope(context.Background(), s)
	handler := assignrolecmd.NewHandler(&mockUserRoleWriteRepo{}, &mockUserRoleReadRepo{}, newTestEventBus(t))

	err := handler.Handle(ctx, assignrolecmd.Command{
		UserID: uuid.New(),
		RoleID: uuid.Nil,
	})
	assert.ErrorIs(t, err, role.ErrNotFound)
}

// ── Already Assigned ───────────────────────────────────────────────────────────

func TestHandle_AlreadyAssigned(t *testing.T) {
	s := newTestScope()
	userID := uuid.New()
	roleID := uuid.New()

	readRepo := &mockUserRoleReadRepo{
		existingRoles: []*role.UserRole{
			{UserID: userID, RoleID: roleID, CompanyID: s.CompanyID},
		},
	}
	handler := assignrolecmd.NewHandler(&mockUserRoleWriteRepo{}, readRepo, newTestEventBus(t))

	ctx := scope.WithScope(context.Background(), s)

	err := handler.Handle(ctx, assignrolecmd.Command{
		UserID: userID,
		RoleID: roleID,
	})
	assert.ErrorIs(t, err, role.ErrAlreadyAssigned)
}

func TestHandle_DifferentRole_SameUser(t *testing.T) {
	s := newTestScope()
	userID := uuid.New()
	existingRoleID := uuid.New()
	newRoleID := uuid.New()

	readRepo := &mockUserRoleReadRepo{
		existingRoles: []*role.UserRole{
			{UserID: userID, RoleID: existingRoleID, CompanyID: s.CompanyID},
		},
	}
	writeRepo := &mockUserRoleWriteRepo{}
	handler := assignrolecmd.NewHandler(writeRepo, readRepo, newTestEventBus(t))

	ctx := scope.WithScope(context.Background(), s)

	err := handler.Handle(ctx, assignrolecmd.Command{
		UserID: userID,
		RoleID: newRoleID,
	})
	require.NoError(t, err)
	require.NotNil(t, writeRepo.assigned)
	assert.Equal(t, newRoleID, writeRepo.assigned.RoleID)
}

// ── CommandName ────────────────────────────────────────────────────────────────

func TestCommand_CommandName(t *testing.T) {
	cmd := assignrolecmd.Command{UserID: uuid.New(), RoleID: uuid.New()}
	assert.Equal(t, "assign_role", cmd.CommandName())
}

// ── Event ──────────────────────────────────────────────────────────────────────

func TestRoleAssignedEvent_EventName(t *testing.T) {
	evt := assignrolecmd.RoleAssignedEvent{
		UserID:    uuid.New(),
		RoleID:    uuid.New(),
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	assert.Equal(t, "role.assigned", evt.EventName())
	assert.NotEmpty(t, evt.AggregateID())
}
