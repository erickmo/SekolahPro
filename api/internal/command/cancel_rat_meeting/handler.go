// Package cancel_rat_meeting menangani command untuk membatalkan RAT meeting.
package cancel_rat_meeting

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/rat"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "cancel_rat_meeting"

// Command berisi data yang dibutuhkan untuk membatalkan RAT meeting.
type Command struct {
	ID              uuid.UUID
	CancelledReason *string
	RescheduledToID *uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command cancel_rat_meeting.
type Handler struct {
	readRepo  rat.ReadRepository
	writeRepo rat.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr rat.ReadRepository, wr rat.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi dan membatalkan RAT meeting.
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
		return fmt.Errorf("get rat meeting for cancel: %w", err)
	}

	if entity.Status == rat.MeetingStatusCompleted || entity.Status == rat.MeetingStatusCancelled {
		return fmt.Errorf("tidak dapat membatalkan meeting dengan status %s", entity.Status)
	}

	entity.Status = rat.MeetingStatusCancelled
	entity.CancelledReason = cmd.CancelledReason
	entity.RescheduledToID = cmd.RescheduledToID
	entity.UpdatedAt = time.Now().UTC()

	if cmd.RescheduledToID != nil {
		entity.Status = rat.MeetingStatusRescheduled
	}

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("cancel rat meeting: %w", err)
	}

	if err := h.eventBus.Publish(ctx, rat.MeetingCancelledEvent{
		TenantID:        s.TenantID,
		CompanyID:       s.CompanyID,
		ID:              entity.ID,
		CancelledReason: cmd.CancelledReason,
	}); err != nil {
		return fmt.Errorf("publish rat_meeting.cancelled: %w", err)
	}

	return nil
}
