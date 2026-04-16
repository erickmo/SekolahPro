// Package create_governance_position menangani command untuk membuat posisi governance baru.
package create_governance_position

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/governance"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_governance_position"

// Command berisi data yang dibutuhkan untuk membuat posisi governance baru.
type Command struct {
	NasabahID         uuid.UUID
	PositionType      string
	PositionLevel     string
	TermStart         time.Time
	TermEnd           time.Time
	TermNumber        int
	AppointedBy       string
	AppointmentDocID  *uuid.UUID
	MaxApprovalAmount int64
	CanDisburse       bool
	CanReverse        bool
	CanWaivePenalty   bool
	CanWriteOff       bool
	CreatedBy         uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_governance_position.
type Handler struct {
	repo     governance.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo governance.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{repo: repo, eventBus: eb}
}

// Handle memvalidasi command, membuat entity, menyimpan ke repository,
// dan mempublikasikan domain event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.NasabahID == uuid.Nil {
		return fmt.Errorf("nasabah_id tidak boleh kosong")
	}
	if cmd.PositionType == "" {
		return fmt.Errorf("position_type tidak boleh kosong")
	}
	if cmd.PositionLevel == "" {
		return fmt.Errorf("position_level tidak boleh kosong")
	}
	if cmd.TermStart.IsZero() {
		return fmt.Errorf("term_start tidak boleh kosong")
	}
	if cmd.TermEnd.IsZero() {
		return fmt.Errorf("term_end tidak boleh kosong")
	}
	if cmd.AppointedBy == "" {
		return fmt.Errorf("appointed_by tidak boleh kosong")
	}
	if cmd.CreatedBy == uuid.Nil {
		return fmt.Errorf("created_by tidak boleh kosong")
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &governance.GovernancePosition{
		ID:                id,
		NasabahID:         cmd.NasabahID,
		PositionType:      cmd.PositionType,
		PositionLevel:     cmd.PositionLevel,
		TermStart:         cmd.TermStart,
		TermEnd:           cmd.TermEnd,
		TermNumber:        cmd.TermNumber,
		Status:            governance.StatusActive,
		AppointedBy:       cmd.AppointedBy,
		AppointmentDocID:  cmd.AppointmentDocID,
		MaxApprovalAmount: cmd.MaxApprovalAmount,
		CanDisburse:       cmd.CanDisburse,
		CanReverse:        cmd.CanReverse,
		CanWaivePenalty:   cmd.CanWaivePenalty,
		CanWriteOff:       cmd.CanWriteOff,
		CreatedAt:         now,
		UpdatedAt:         now,
		CreatedBy:         cmd.CreatedBy,
		UpdatedBy:         cmd.CreatedBy,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save governance position: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, governance.GovernancePositionCreatedEvent{
		TenantID:      s.TenantID,
		CompanyID:     s.CompanyID,
		ID:            id,
		NasabahID:     cmd.NasabahID,
		PositionType:  cmd.PositionType,
		PositionLevel: cmd.PositionLevel,
	}); err != nil {
		return fmt.Errorf("publish governance_position.created: %w", err)
	}

	return nil
}
