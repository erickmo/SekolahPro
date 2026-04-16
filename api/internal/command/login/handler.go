// Package login menangani command untuk autentikasi user.
package login

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/boilerplate/internal/domain/user"
	jwtpkg "github.com/yourorg/boilerplate/pkg/jwt"
)

const commandName = "login"

// Command berisi data yang dibutuhkan untuk login.
type Command struct {
	Email    string
	Password string
}

func (c Command) CommandName() string { return commandName }

// Result berisi token yang dikembalikan setelah login berhasil.
type Result struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// Handler menangani Command login.
type Handler struct {
	userReadRepo  user.ReadRepository
	userWriteRepo user.WriteRepository
	jwtSvc        *jwtpkg.Service
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(
	userReadRepo user.ReadRepository,
	userWriteRepo user.WriteRepository,
	jwtSvc *jwtpkg.Service,
) *Handler {
	return &Handler{
		userReadRepo:  userReadRepo,
		userWriteRepo: userWriteRepo,
		jwtSvc:        jwtSvc,
	}
}

// Handle memvalidasi credential user, generate token, dan update last login.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.Email == "" {
		return nil, user.ErrEmailEmpty
	}
	if cmd.Password == "" {
		return nil, user.ErrPasswordEmpty
	}

	u, err := h.userReadRepo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, user.ErrWrongPassword
	}

	if !u.IsActive {
		return nil, user.ErrInactive
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash), []byte(cmd.Password),
	); err != nil {
		return nil, user.ErrWrongPassword
	}

	if err := h.userWriteRepo.UpdateLastLogin(ctx, u.ID); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}

	tenantID, err := uuid.Parse("00000000-0000-0000-0000-000000000000")
	if err != nil {
		return nil, fmt.Errorf("parse tenant id: %w", err)
	}

	pair, err := h.jwtSvc.GenerateTokenPair(u.ID, u.Email, "user", tenantID)
	if err != nil {
		return nil, fmt.Errorf("generate token pair: %w", err)
	}

	return &Result{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
	}, nil
}
