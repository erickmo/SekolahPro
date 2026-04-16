// Package assign_role menangani command untuk menetapkan role ke user.
package assign_role

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/role"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "assign_role"

// Command berisi data yang dibutuhkan untuk menetapkan role ke user.
type Command struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// RoleAssignedEvent dipublish setelah role berhasil ditetapkan ke user.
type RoleAssignedEvent struct {
	UserID    uuid.UUID `json:"user_id"`
	RoleID    uuid.UUID `json:"role_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
}

func (e RoleAssignedEvent) EventName() string    { return "role.assigned" }
func (e RoleAssignedEvent) AggregateID() string { return e.UserID.String() }

// Handler menangani Command assign_role.
type Handler struct {
	userRoleWriteRepo role.UserRoleWriteRepository
	userRoleReadRepo  role.UserRoleReadRepository
	eventBus          eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(
	userRoleWriteRepo role.UserRoleWriteRepository,
	userRoleReadRepo role.UserRoleReadRepository,
	eb eventbus.EventBus,
) *Handler {
	return &Handler{
		userRoleWriteRepo: userRoleWriteRepo,
		userRoleReadRepo:  userRoleReadRepo,
		eventBus:          eb,
	}
}

// Handle memvalidasi command, memeriksa duplikasi, lalu menetapkan role ke user.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.UserID == uuid.Nil {
		return role.ErrNotFound
	}
	if cmd.RoleID == uuid.Nil {
		return role.ErrNotFound
	}

	existing, err := h.userRoleReadRepo.GetByUserAndCompany(ctx, cmd.UserID, s.CompanyID)
	if err == nil {
		for _, ur := range existing {
			if ur.RoleID == cmd.RoleID {
				return role.ErrAlreadyAssigned
			}
		}
	}

	ur := &role.UserRole{
		ID:         uuid.New(),
		UserID:     cmd.UserID,
		RoleID:     cmd.RoleID,
		TenantID:   s.TenantID,
		CompanyID:  s.CompanyID,
		IsActive:   true,
		AssignedAt: time.Now().UTC(),
	}

	if err := h.userRoleWriteRepo.Assign(ctx, ur); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	if err := h.eventBus.Publish(ctx, RoleAssignedEvent{
		UserID:    cmd.UserID,
		RoleID:    cmd.RoleID,
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
	}); err != nil {
		return fmt.Errorf("publish role.assigned: %w", err)
	}

	return nil
}
