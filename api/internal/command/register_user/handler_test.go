package register_user_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/user"
	registercmd "github.com/yourorg/boilerplate/internal/command/register_user"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Mocks ──────────────────────────────────────────────────────────────────────

type mockUserReadRepo struct {
	byEmail map[string]*user.User
}

func (m *mockUserReadRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return nil, user.ErrNotFound
}

func (m *mockUserReadRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (m *mockUserReadRepo) List(ctx context.Context, s scope.Scope, limit, offset int) ([]*user.User, int, error) {
	return nil, 0, nil
}

type mockUserWriteRepo struct {
	savedUser *user.User
	saveErr   error
}

func (m *mockUserWriteRepo) Save(ctx context.Context, u *user.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUser = u
	return nil
}

func (m *mockUserWriteRepo) Update(ctx context.Context, u *user.User) error               { return nil }
func (m *mockUserWriteRepo) UpdatePassword(ctx context.Context, id uuid.UUID, h string) error { return nil }
func (m *mockUserWriteRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *mockUserWriteRepo) Deactivate(ctx context.Context, id uuid.UUID) error             { return nil }
func (m *mockUserWriteRepo) SoftDelete(ctx context.Context, id uuid.UUID) error             { return nil }

func newTestEventBus(t *testing.T) eventbus.EventBus {
	t.Helper()
	eb, err := eventbus.NewInMemoryEventBus()
	require.NoError(t, err)
	return eb
}

// ── Register Success ───────────────────────────────────────────────────────────

func TestHandle_Success(t *testing.T) {
	readRepo := &mockUserReadRepo{byEmail: map[string]*user.User{}}
	writeRepo := &mockUserWriteRepo{}
	handler := registercmd.NewHandler(writeRepo, readRepo, newTestEventBus(t))

	ctx, holder := commandbus.WithResultID(context.Background())

	err := handler.Handle(ctx, registercmd.Command{
		Email:    "newuser@test.com",
		Password: "password123",
		FullName: "New User",
		Phone:    "08123456789",
	})
	require.NoError(t, err)

	// Verify user was saved
	require.NotNil(t, writeRepo.savedUser)
	assert.Equal(t, "newuser@test.com", writeRepo.savedUser.Email)
	assert.Equal(t, "New User", writeRepo.savedUser.FullName)
	assert.Equal(t, "08123456789", writeRepo.savedUser.Phone)
	assert.True(t, writeRepo.savedUser.IsActive)
	assert.NotEmpty(t, writeRepo.savedUser.PasswordHash)
	assert.NotEqual(t, "password123", writeRepo.savedUser.PasswordHash) // hashed

	// Verify CreatedID was set
	assert.NotEqual(t, uuid.Nil, holder.ID)
	assert.Equal(t, writeRepo.savedUser.ID, holder.ID)
}

func TestHandle_Success_WithoutPhone(t *testing.T) {
	readRepo := &mockUserReadRepo{byEmail: map[string]*user.User{}}
	writeRepo := &mockUserWriteRepo{}
	handler := registercmd.NewHandler(writeRepo, readRepo, newTestEventBus(t))

	err := handler.Handle(context.Background(), registercmd.Command{
		Email:    "nophone@test.com",
		Password: "password123",
		FullName: "No Phone User",
	})
	require.NoError(t, err)

	require.NotNil(t, writeRepo.savedUser)
	assert.Empty(t, writeRepo.savedUser.Phone)
}

// ── Validation Errors ──────────────────────────────────────────────────────────

func TestHandle_EmptyEmail(t *testing.T) {
	handler := registercmd.NewHandler(&mockUserWriteRepo{}, &mockUserReadRepo{}, newTestEventBus(t))

	err := handler.Handle(context.Background(), registercmd.Command{
		Email:    "",
		Password: "password123",
		FullName: "Test User",
	})
	assert.ErrorIs(t, err, user.ErrEmailEmpty)
}

func TestHandle_EmptyFullName(t *testing.T) {
	handler := registercmd.NewHandler(&mockUserWriteRepo{}, &mockUserReadRepo{}, newTestEventBus(t))

	err := handler.Handle(context.Background(), registercmd.Command{
		Email:    "test@example.com",
		Password: "password123",
		FullName: "",
	})
	assert.ErrorIs(t, err, user.ErrFullNameEmpty)
}

func TestHandle_WeakPassword_7Chars(t *testing.T) {
	handler := registercmd.NewHandler(&mockUserWriteRepo{}, &mockUserReadRepo{}, newTestEventBus(t))

	err := handler.Handle(context.Background(), registercmd.Command{
		Email:    "test@example.com",
		Password: "1234567", // 7 chars
		FullName: "Test User",
	})
	assert.ErrorIs(t, err, user.ErrPasswordWeak)
}

func TestHandle_WeakPassword_Exact8(t *testing.T) {
	readRepo := &mockUserReadRepo{byEmail: map[string]*user.User{}}
	writeRepo := &mockUserWriteRepo{}
	handler := registercmd.NewHandler(writeRepo, readRepo, newTestEventBus(t))

	err := handler.Handle(context.Background(), registercmd.Command{
		Email:    "test@example.com",
		Password: "12345678", // exactly 8
		FullName: "Test User",
	})
	assert.NoError(t, err)
}

// ── Duplicate Email ────────────────────────────────────────────────────────────

func TestHandle_DuplicateEmail(t *testing.T) {
	readRepo := &mockUserReadRepo{
		byEmail: map[string]*user.User{
			"existing@test.com": {
				ID:    uuid.New(),
				Email: "existing@test.com",
			},
		},
	}
	handler := registercmd.NewHandler(&mockUserWriteRepo{}, readRepo, newTestEventBus(t))

	err := handler.Handle(context.Background(), registercmd.Command{
		Email:    "existing@test.com",
		Password: "password123",
		FullName: "Duplicate User",
	})
	assert.ErrorIs(t, err, user.ErrEmailExists)
}

// ── CommandName ────────────────────────────────────────────────────────────────

func TestCommand_CommandName(t *testing.T) {
	cmd := registercmd.Command{Email: "a@b.com", Password: "pass", FullName: "Test"}
	assert.Equal(t, "register_user", cmd.CommandName())
}

// ── Event ──────────────────────────────────────────────────────────────────────

func TestUserRegisteredEvent_EventName(t *testing.T) {
	evt := registercmd.UserRegisteredEvent{
		UserID:   uuid.New(),
		Email:    "test@test.com",
		FullName: "Test",
	}
	assert.Equal(t, "user.registered", evt.EventName())
	assert.NotEmpty(t, evt.AggregateID())
}
