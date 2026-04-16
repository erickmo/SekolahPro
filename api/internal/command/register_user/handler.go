// Package register_user menangani command untuk registrasi user baru.
package register_user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/boilerplate/internal/domain/user"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
)

const (
	commandName     = "register_user"
	bcryptCost      = 12
	minPasswordLen  = 8
)

// Command berisi data yang dibutuhkan untuk registrasi user baru.
type Command struct {
	Email    string
	Password string
	FullName string
	Phone    string
}

func (c Command) CommandName() string { return commandName }

// UserRegisteredEvent dipublish setelah user berhasil didaftarkan.
type UserRegisteredEvent struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
}

func (e UserRegisteredEvent) EventName() string    { return "user.registered" }
func (e UserRegisteredEvent) AggregateID() string { return e.UserID.String() }

// Handler menangani Command register_user.
type Handler struct {
	userWriteRepo user.WriteRepository
	userReadRepo  user.ReadRepository
	eventBus      eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(
	userWriteRepo user.WriteRepository,
	userReadRepo user.ReadRepository,
	eb eventbus.EventBus,
) *Handler {
	return &Handler{
		userWriteRepo: userWriteRepo,
		userReadRepo:  userReadRepo,
		eventBus:      eb,
	}
}

// Handle memvalidasi data registrasi, membuat user baru, dan mempublikasikan event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if cmd.Email == "" {
		return user.ErrEmailEmpty
	}
	if cmd.FullName == "" {
		return user.ErrFullNameEmpty
	}
	if len(cmd.Password) < minPasswordLen {
		return user.ErrPasswordWeak
	}

	existing, err := h.userReadRepo.GetByEmail(ctx, cmd.Email)
	if err == nil && existing != nil {
		return user.ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	id := uuid.New()

	u := &user.User{
		ID:           id,
		Email:        cmd.Email,
		PasswordHash: string(hash),
		FullName:     cmd.FullName,
		Phone:        cmd.Phone,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.userWriteRepo.Save(ctx, u); err != nil {
		return fmt.Errorf("save user: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, UserRegisteredEvent{
		UserID:   id,
		Email:    cmd.Email,
		FullName: cmd.FullName,
	}); err != nil {
		return fmt.Errorf("publish user.registered: %w", err)
	}

	return nil
}
