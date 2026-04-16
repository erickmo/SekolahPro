package create_role_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/role"
	createrolecmd "github.com/yourorg/boilerplate/internal/command/create_role"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Mocks ──────────────────────────────────────────────────────────────────────

type mockRoleWriteRepo struct {
	savedRole *role.Role
	saveErr   error
}

func (m *mockRoleWriteRepo) Save(ctx context.Context, s scope.Scope, r *role.Role) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedRole = r
	return nil
}

func (m *mockRoleWriteRepo) Update(ctx context.Context, s scope.Scope, r *role.Role) error {
	return nil
}

func (m *mockRoleWriteRepo) SoftDelete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	return nil
}

type mockRoleReadRepo struct {
	byCode map[string]*role.Role
}

func (m *mockRoleReadRepo) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*role.Role, error) {
	return nil, role.ErrNotFound
}

func (m *mockRoleReadRepo) GetByCode(ctx context.Context, s scope.Scope, code string) (*role.Role, error) {
	r, ok := m.byCode[code]
	if !ok {
		return nil, role.ErrNotFound
	}
	return r, nil
}

func (m *mockRoleReadRepo) List(ctx context.Context, s scope.Scope) ([]*role.Role, error) {
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

// ── Create Role Success ────────────────────────────────────────────────────────

func TestHandle_Success(t *testing.T) {
	s := newTestScope()
	readRepo := &mockRoleReadRepo{byCode: map[string]*role.Role{}}
	writeRepo := &mockRoleWriteRepo{}
	handler := createrolecmd.NewHandler(writeRepo, readRepo, newTestEventBus(t))

	ctx, holder := commandbus.WithResultID(scope.WithScope(context.Background(), s))

	err := handler.Handle(ctx, createrolecmd.Command{
		Name:        "Guru Matematika",
		Code:        "math_teacher",
		Description: "Guru mata pelajaran matematika",
		RoleType:    role.TypeTeacher,
		Permissions: []string{"students:read", "grades:write"},
	})
	require.NoError(t, err)

	// Verify role was saved
	require.NotNil(t, writeRepo.savedRole)
	assert.Equal(t, "Guru Matematika", writeRepo.savedRole.Name)
	assert.Equal(t, "math_teacher", writeRepo.savedRole.Code)
	assert.Equal(t, role.TypeTeacher, writeRepo.savedRole.RoleType)
	assert.False(t, writeRepo.savedRole.IsSystem)
	assert.Equal(t, []string{"students:read", "grades:write"}, writeRepo.savedRole.Permissions)

	// Verify CreatedID was set
	assert.NotEqual(t, uuid.Nil, holder.ID)
	assert.Equal(t, writeRepo.savedRole.ID, holder.ID)
}

func TestHandle_Success_NilPermissions(t *testing.T) {
	s := newTestScope()
	readRepo := &mockRoleReadRepo{byCode: map[string]*role.Role{}}
	writeRepo := &mockRoleWriteRepo{}
	handler := createrolecmd.NewHandler(writeRepo, readRepo, newTestEventBus(t))

	ctx := scope.WithScope(context.Background(), s)

	err := handler.Handle(ctx, createrolecmd.Command{
		Name:     "Custom Role",
		Code:     "custom",
		RoleType: role.TypeStaff,
		// Permissions: nil
	})
	require.NoError(t, err)

	require.NotNil(t, writeRepo.savedRole)
	assert.Equal(t, []string{}, writeRepo.savedRole.Permissions) // nil → empty slice
}

// ── Scope Validation ───────────────────────────────────────────────────────────

func TestHandle_NoScope(t *testing.T) {
	handler := createrolecmd.NewHandler(&mockRoleWriteRepo{}, &mockRoleReadRepo{}, newTestEventBus(t))

	err := handler.Handle(context.Background(), createrolecmd.Command{
		Name: "Role",
		Code: "code",
	})
	assert.ErrorIs(t, err, scope.ErrMissingTenant)
}

func TestHandle_MissingCompany(t *testing.T) {
	s := scope.Scope{TenantID: uuid.New()} // CompanyID kosong
	ctx := scope.WithScope(context.Background(), s)
	handler := createrolecmd.NewHandler(&mockRoleWriteRepo{}, &mockRoleReadRepo{}, newTestEventBus(t))

	err := handler.Handle(ctx, createrolecmd.Command{
		Name: "Role",
		Code: "code",
	})
	assert.ErrorIs(t, err, scope.ErrMissingCompany)
}

// ── Validation Errors ──────────────────────────────────────────────────────────

func TestHandle_EmptyName(t *testing.T) {
	s := newTestScope()
	ctx := scope.WithScope(context.Background(), s)
	handler := createrolecmd.NewHandler(&mockRoleWriteRepo{}, &mockRoleReadRepo{}, newTestEventBus(t))

	err := handler.Handle(ctx, createrolecmd.Command{
		Name: "",
		Code: "code",
	})
	assert.ErrorIs(t, err, role.ErrNameEmpty)
}

func TestHandle_EmptyCode(t *testing.T) {
	s := newTestScope()
	ctx := scope.WithScope(context.Background(), s)
	handler := createrolecmd.NewHandler(&mockRoleWriteRepo{}, &mockRoleReadRepo{}, newTestEventBus(t))

	err := handler.Handle(ctx, createrolecmd.Command{
		Name: "Role Name",
		Code: "",
	})
	assert.ErrorIs(t, err, role.ErrCodeEmpty)
}

// ── Duplicate Code ─────────────────────────────────────────────────────────────

func TestHandle_DuplicateCode(t *testing.T) {
	s := newTestScope()
	readRepo := &mockRoleReadRepo{
		byCode: map[string]*role.Role{
			"existing_code": {ID: uuid.New(), Code: "existing_code"},
		},
	}
	handler := createrolecmd.NewHandler(&mockRoleWriteRepo{}, readRepo, newTestEventBus(t))

	ctx := scope.WithScope(context.Background(), s)

	err := handler.Handle(ctx, createrolecmd.Command{
		Name: "Duplicate",
		Code: "existing_code",
	})
	assert.ErrorIs(t, err, role.ErrCodeExists)
}

// ── CommandName ────────────────────────────────────────────────────────────────

func TestCommand_CommandName(t *testing.T) {
	cmd := createrolecmd.Command{Name: "n", Code: "c"}
	assert.Equal(t, "create_role", cmd.CommandName())
}

// ── Event ──────────────────────────────────────────────────────────────────────

func TestRoleCreatedEvent_EventName(t *testing.T) {
	evt := createrolecmd.RoleCreatedEvent{
		RoleID:    uuid.New(),
		Name:      "Test",
		Code:      "test",
		TenantID:  uuid.New(),
		CompanyID: uuid.New(),
	}
	assert.Equal(t, "role.created", evt.EventName())
	assert.NotEmpty(t, evt.AggregateID())
}
