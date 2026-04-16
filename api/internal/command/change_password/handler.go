// Package change_password menangani command untuk mengubah password user.
package change_password

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/boilerplate/internal/domain/user"
)

const (
	commandName    = "change_password"
	bcryptCost     = 12
	minPasswordLen = 8
)

// Command berisi data yang dibutuhkan untuk mengubah password.
type Command struct {
	UserID      uuid.UUID
	OldPassword string
	NewPassword string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command change_password.
type Handler struct {
	userReadRepo  user.ReadRepository
	userWriteRepo user.WriteRepository
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(
	userReadRepo user.ReadRepository,
	userWriteRepo user.WriteRepository,
) *Handler {
	return &Handler{
		userReadRepo:  userReadRepo,
		userWriteRepo: userWriteRepo,
	}
}

// Handle memvalidasi password lama, lalu mengupdate ke password baru.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if cmd.UserID == uuid.Nil {
		return user.ErrNotFound
	}
	if cmd.OldPassword == "" {
		return user.ErrPasswordEmpty
	}
	if len(cmd.NewPassword) < minPasswordLen {
		return user.ErrPasswordWeak
	}

	u, err := h.userReadRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return user.ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash), []byte(cmd.OldPassword),
	); err != nil {
		return user.ErrWrongPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.NewPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := h.userWriteRepo.UpdatePassword(ctx, cmd.UserID, string(hash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
