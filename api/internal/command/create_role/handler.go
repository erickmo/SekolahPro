// Package create_role menangani command untuk membuat role baru.
package create_role

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/role"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_role"

// Command berisi data yang dibutuhkan untuk membuat role baru.
type Command struct {
	Name        string
	Code        string
	Description string
	RoleType    string
	Permissions []string
}

func (c Command) CommandName() string { return commandName }

// RoleCreatedEvent dipublish setelah role berhasil dibuat.
type RoleCreatedEvent struct {
	RoleID    uuid.UUID `json:"role_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
}

func (e RoleCreatedEvent) EventName() string    { return "role.created" }
func (e RoleCreatedEvent) AggregateID() string { return e.RoleID.String() }

// Handler menangani Command create_role.
type Handler struct {
	roleWriteRepo role.RoleWriteRepository
	roleReadRepo  role.RoleReadRepository
	eventBus      eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(
	roleWriteRepo role.RoleWriteRepository,
	roleReadRepo role.RoleReadRepository,
	eb eventbus.EventBus,
) *Handler {
	return &Handler{
		roleWriteRepo: roleWriteRepo,
		roleReadRepo:  roleReadRepo,
		eventBus:      eb,
	}
}

// Handle memvalidasi command, membuat role entity, menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.Name == "" {
		return role.ErrNameEmpty
	}
	if cmd.Code == "" {
		return role.ErrCodeEmpty
	}

	existing, err := h.roleReadRepo.GetByCode(ctx, s, cmd.Code)
	if err == nil && existing != nil {
		return role.ErrCodeExists
	}

	now := time.Now().UTC()
	id := uuid.New()

	permissions := cmd.Permissions
	if permissions == nil {
		permissions = []string{}
	}

	rl := &role.Role{
		ID:          id,
		Name:        cmd.Name,
		Code:        cmd.Code,
		Description: cmd.Description,
		RoleType:    cmd.RoleType,
		IsSystem:    false,
		Permissions: permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.roleWriteRepo.Save(ctx, s, rl); err != nil {
		return fmt.Errorf("save role: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, RoleCreatedEvent{
		RoleID:    id,
		Name:      cmd.Name,
		Code:      cmd.Code,
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
	}); err != nil {
		return fmt.Errorf("publish role.created: %w", err)
	}

	return nil
}
