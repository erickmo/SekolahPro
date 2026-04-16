package change_password_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/boilerplate/internal/domain/user"
	changepwcmd "github.com/yourorg/boilerplate/internal/command/change_password"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Mocks ──────────────────────────────────────────────────────────────────────

type mockReadRepo struct {
	users map[uuid.UUID]*user.User
}

func (m *mockReadRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (m *mockReadRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, user.ErrNotFound
}

func (m *mockReadRepo) List(ctx context.Context, s scope.Scope, limit, offset int) ([]*user.User, int, error) {
	return nil, 0, nil
}

type mockWriteRepo struct {
	updatedPasswordID *uuid.UUID
	updatedHash       string
}

func (m *mockWriteRepo) Save(ctx context.Context, u *user.User) error                    { return nil }
func (m *mockWriteRepo) Update(ctx context.Context, u *user.User) error                   { return nil }
func (m *mockWriteRepo) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	m.updatedPasswordID = &id
	m.updatedHash = hash
	return nil
}
func (m *mockWriteRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error          { return nil }
func (m *mockWriteRepo) Deactivate(ctx context.Context, id uuid.UUID) error               { return nil }
func (m *mockWriteRepo) SoftDelete(ctx context.Context, id uuid.UUID) error               { return nil }

func hashPW(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	require.NoError(t, err)
	return string(h)
}

// ── Change Password Success ───────────────────────────────────────────────────

func TestHandle_Success(t *testing.T) {
	userID := uuid.New()
	oldHash := hashPW(t, "oldpassword123")

	readRepo := &mockReadRepo{
		users: map[uuid.UUID]*user.User{
			userID: {
				ID:           userID,
				PasswordHash: oldHash,
			},
		},
	}
	writeRepo := &mockWriteRepo{}
	handler := changepwcmd.NewHandler(readRepo, writeRepo)

	err := handler.Handle(context.Background(), changepwcmd.Command{
		UserID:      userID,
		OldPassword: "oldpassword123",
		NewPassword: "newpassword456",
	})
	require.NoError(t, err)

	// Verify password was updated
	require.NotNil(t, writeRepo.updatedPasswordID)
	assert.Equal(t, userID, *writeRepo.updatedPasswordID)
	assert.NotEqual(t, oldHash, writeRepo.updatedHash)

	// Verify new password can be validated
	err = bcrypt.CompareHashAndPassword([]byte(writeRepo.updatedHash), []byte("newpassword456"))
	assert.NoError(t, err)
}

// ── Validation Errors ──────────────────────────────────────────────────────────

func TestHandle_NilUserID(t *testing.T) {
	handler := changepwcmd.NewHandler(&mockReadRepo{}, &mockWriteRepo{})

	err := handler.Handle(context.Background(), changepwcmd.Command{
		UserID:      uuid.Nil,
		OldPassword: "old",
		NewPassword: "newpassword123",
	})
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestHandle_EmptyOldPassword(t *testing.T) {
	handler := changepwcmd.NewHandler(&mockReadRepo{}, &mockWriteRepo{})

	err := handler.Handle(context.Background(), changepwcmd.Command{
		UserID:      uuid.New(),
		OldPassword: "",
		NewPassword: "newpassword123",
	})
	assert.ErrorIs(t, err, user.ErrPasswordEmpty)
}

func TestHandle_WeakNewPassword(t *testing.T) {
	handler := changepwcmd.NewHandler(&mockReadRepo{}, &mockWriteRepo{})

	err := handler.Handle(context.Background(), changepwcmd.Command{
		UserID:      uuid.New(),
		OldPassword: "oldpassword123",
		NewPassword: "short",
	})
	assert.ErrorIs(t, err, user.ErrPasswordWeak)
}

func TestHandle_ExactMinLength(t *testing.T) {
	userID := uuid.New()
	readRepo := &mockReadRepo{
		users: map[uuid.UUID]*user.User{
			userID: {ID: userID, PasswordHash: hashPW(t, "oldpassword123")},
		},
	}
	writeRepo := &mockWriteRepo{}
	handler := changepwcmd.NewHandler(readRepo, writeRepo)

	// Password with exactly 8 characters should pass
	err := handler.Handle(context.Background(), changepwcmd.Command{
		UserID:      userID,
		OldPassword: "oldpassword123",
		NewPassword: "12345678", // exactly 8
	})
	assert.NoError(t, err)
}

func TestHandle_SevenCharPassword(t *testing.T) {
	handler := changepwcmd.NewHandler(&mockReadRepo{}, &mockWriteRepo{})

	err := handler.Handle(context.Background(), changepwcmd.Command{
		UserID:      uuid.New(),
		OldPassword: "oldpassword123",
		NewPassword: "1234567", // 7 chars
	})
	assert.ErrorIs(t, err, user.ErrPasswordWeak)
}

// ── User Not Found ─────────────────────────────────────────────────────────────

func TestHandle_UserNotFound(t *testing.T) {
	readRepo := &mockReadRepo{users: map[uuid.UUID]*user.User{}}
	handler := changepwcmd.NewHandler(readRepo, &mockWriteRepo{})

	err := handler.Handle(context.Background(), changepwcmd.Command{
		UserID:      uuid.New(),
		OldPassword: "oldpassword123",
		NewPassword: "newpassword456",
	})
	assert.ErrorIs(t, err, user.ErrNotFound)
}

// ── Wrong Old Password ────────────────────────────────────────────────────────

func TestHandle_WrongOldPassword(t *testing.T) {
	userID := uuid.New()
	readRepo := &mockReadRepo{
		users: map[uuid.UUID]*user.User{
			userID: {ID: userID, PasswordHash: hashPW(t, "correctold")},
		},
	}
	handler := changepwcmd.NewHandler(readRepo, &mockWriteRepo{})

	err := handler.Handle(context.Background(), changepwcmd.Command{
		UserID:      userID,
		OldPassword: "wrongold",
		NewPassword: "newpassword456",
	})
	assert.ErrorIs(t, err, user.ErrWrongPassword)
}

// ── CommandName ────────────────────────────────────────────────────────────────

func TestCommand_CommandName(t *testing.T) {
	cmd := changepwcmd.Command{
		UserID:      uuid.New(),
		OldPassword: "old",
		NewPassword: "newpassword123",
	}
	assert.Equal(t, "change_password", cmd.CommandName())
}
