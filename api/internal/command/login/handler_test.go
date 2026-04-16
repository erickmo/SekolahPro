package login_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/boilerplate/internal/domain/user"
	logincmd "github.com/yourorg/boilerplate/internal/command/login"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ── Mocks ──────────────────────────────────────────────────────────────────────

type mockUserReadRepo struct {
	byEmail map[string]*user.User
	byID    map[uuid.UUID]*user.User
}

func (m *mockUserReadRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
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
	lastLoginID *uuid.UUID
}

func (m *mockUserWriteRepo) Save(ctx context.Context, u *user.User) error                { return nil }
func (m *mockUserWriteRepo) Update(ctx context.Context, u *user.User) error               { return nil }
func (m *mockUserWriteRepo) UpdatePassword(ctx context.Context, id uuid.UUID, h string) error { return nil }
func (m *mockUserWriteRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	m.lastLoginID = &id
	return nil
}
func (m *mockUserWriteRepo) Deactivate(ctx context.Context, id uuid.UUID) error           { return nil }
func (m *mockUserWriteRepo) SoftDelete(ctx context.Context, id uuid.UUID) error           { return nil }

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(hash)
}

func newTestJWTService() *jwtpkg.Service {
	return jwtpkg.NewService("test-secret-must-be-32-characters-minimum", 1, "test", 7)
}

// ── Login Success ──────────────────────────────────────────────────────────────

func TestHandle_Success(t *testing.T) {
	jwtSvc := newTestJWTService()
	userID := uuid.New()
	passwordHash := hashPassword(t, "password123")

	readRepo := &mockUserReadRepo{
		byEmail: map[string]*user.User{
			"user@test.com": {
				ID:           userID,
				Email:        "user@test.com",
				PasswordHash: passwordHash,
				IsActive:     true,
			},
		},
	}
	writeRepo := &mockUserWriteRepo{}
	handler := logincmd.NewHandler(readRepo, writeRepo, jwtSvc)

	result, err := handler.Handle(context.Background(), logincmd.Command{
		Email:    "user@test.com",
		Password: "password123",
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotZero(t, result.ExpiresAt)

	// Verify UpdateLastLogin was called
	require.NotNil(t, writeRepo.lastLoginID)
	assert.Equal(t, userID, *writeRepo.lastLoginID)

	// Verify token is valid
	claims, err := jwtSvc.ValidateClaims(result.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, "user@test.com", claims.Email)
}

// ── Validation Errors ──────────────────────────────────────────────────────────

func TestHandle_EmptyEmail(t *testing.T) {
	handler := logincmd.NewHandler(&mockUserReadRepo{}, &mockUserWriteRepo{}, newTestJWTService())

	_, err := handler.Handle(context.Background(), logincmd.Command{
		Email:    "",
		Password: "password123",
	})
	assert.ErrorIs(t, err, user.ErrEmailEmpty)
}

func TestHandle_EmptyPassword(t *testing.T) {
	handler := logincmd.NewHandler(&mockUserReadRepo{}, &mockUserWriteRepo{}, newTestJWTService())

	_, err := handler.Handle(context.Background(), logincmd.Command{
		Email:    "user@test.com",
		Password: "",
	})
	assert.ErrorIs(t, err, user.ErrPasswordEmpty)
}

// ── User Not Found ─────────────────────────────────────────────────────────────

func TestHandle_UserNotFound(t *testing.T) {
	readRepo := &mockUserReadRepo{byEmail: map[string]*user.User{}}
	handler := logincmd.NewHandler(readRepo, &mockUserWriteRepo{}, newTestJWTService())

	_, err := handler.Handle(context.Background(), logincmd.Command{
		Email:    "nonexistent@test.com",
		Password: "password123",
	})
	// User not found should return wrong password to prevent email enumeration
	assert.ErrorIs(t, err, user.ErrWrongPassword)
}

// ── Inactive User ──────────────────────────────────────────────────────────────

func TestHandle_InactiveUser(t *testing.T) {
	readRepo := &mockUserReadRepo{
		byEmail: map[string]*user.User{
			"inactive@test.com": {
				ID:           uuid.New(),
				Email:        "inactive@test.com",
				PasswordHash: hashPassword(t, "password123"),
				IsActive:     false,
			},
		},
	}
	handler := logincmd.NewHandler(readRepo, &mockUserWriteRepo{}, newTestJWTService())

	_, err := handler.Handle(context.Background(), logincmd.Command{
		Email:    "inactive@test.com",
		Password: "password123",
	})
	assert.ErrorIs(t, err, user.ErrInactive)
}

// ── Wrong Password ─────────────────────────────────────────────────────────────

func TestHandle_WrongPassword(t *testing.T) {
	readRepo := &mockUserReadRepo{
		byEmail: map[string]*user.User{
			"user@test.com": {
				ID:           uuid.New(),
				Email:        "user@test.com",
				PasswordHash: hashPassword(t, "correctpassword"),
				IsActive:     true,
			},
		},
	}
	handler := logincmd.NewHandler(readRepo, &mockUserWriteRepo{}, newTestJWTService())

	_, err := handler.Handle(context.Background(), logincmd.Command{
		Email:    "user@test.com",
		Password: "wrongpassword",
	})
	assert.ErrorIs(t, err, user.ErrWrongPassword)
}

// ── CommandName ────────────────────────────────────────────────────────────────

func TestCommand_CommandName(t *testing.T) {
	cmd := logincmd.Command{Email: "a@b.com", Password: "pass"}
	assert.Equal(t, "login", cmd.CommandName())
}

// ── Result fields ──────────────────────────────────────────────────────────────

func TestResult_Fields(t *testing.T) {
	r := &logincmd.Result{
		AccessToken:  "at",
		RefreshToken: "rt",
		ExpiresAt:    time.Now().Unix(),
	}
	assert.Equal(t, "at", r.AccessToken)
	assert.Equal(t, "rt", r.RefreshToken)
	assert.NotZero(t, r.ExpiresAt)
}
