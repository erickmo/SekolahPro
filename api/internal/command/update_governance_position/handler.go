// Package update_governance_position menangani command untuk mengupdate posisi governance.
package update_governance_position

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/governance"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_governance_position"

// Command berisi data yang dibutuhkan untuk mengupdate posisi governance.
type Command struct {
	ID                uuid.UUID
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
	UpdatedBy         uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_governance_position.
type Handler struct {
	readRepo  governance.ReadRepository
	writeRepo governance.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr governance.ReadRepository, wr governance.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengambil entity yang ada, mengupdate, dan mempublish event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get governance position for update: %w", err)
	}

	if cmd.PositionType != "" {
		entity.PositionType = cmd.PositionType
	}
	if cmd.PositionLevel != "" {
		entity.PositionLevel = cmd.PositionLevel
	}
	if !cmd.TermStart.IsZero() {
		entity.TermStart = cmd.TermStart
	}
	if !cmd.TermEnd.IsZero() {
		entity.TermEnd = cmd.TermEnd
	}
	if cmd.TermNumber > 0 {
		entity.TermNumber = cmd.TermNumber
	}
	if cmd.AppointedBy != "" {
		entity.AppointedBy = cmd.AppointedBy
	}
	entity.AppointmentDocID = cmd.AppointmentDocID
	entity.MaxApprovalAmount = cmd.MaxApprovalAmount
	entity.CanDisburse = cmd.CanDisburse
	entity.CanReverse = cmd.CanReverse
	entity.CanWaivePenalty = cmd.CanWaivePenalty
	entity.CanWriteOff = cmd.CanWriteOff
	entity.UpdatedAt = time.Now().UTC()
	if cmd.UpdatedBy != uuid.Nil {
		entity.UpdatedBy = cmd.UpdatedBy
	}

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update governance position: %w", err)
	}

	if err := h.eventBus.Publish(ctx, governance.GovernancePositionUpdatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
	}); err != nil {
		return fmt.Errorf("publish governance_position.updated: %w", err)
	}

	return nil
}
