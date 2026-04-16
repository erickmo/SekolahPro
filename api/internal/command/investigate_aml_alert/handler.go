// Package investigate_aml_alert menangani command untuk memulai investigasi alert.
package investigate_aml_alert

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_alert"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "investigate_aml_alert"

// Command berisi data yang dibutuhkan untuk memulai investigasi alert.
type Command struct {
	ID               uuid.UUID
	InvestigatedBy   uuid.UUID
	InvestigationNotes string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command investigate_aml_alert.
type Handler struct {
	readRepo  aml_alert.ReadRepository
	writeRepo aml_alert.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr aml_alert.ReadRepository, wr aml_alert.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi dan memulai investigasi alert.
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
		return fmt.Errorf("get aml alert for investigate: %w", err)
	}

	if entity.Status != aml_alert.AlertStatusNew && entity.Status != aml_alert.AlertStatusEscalated {
		return fmt.Errorf("alert dengan status %s tidak dapat diinvestigasi", entity.Status)
	}

	entity.Status = aml_alert.AlertStatusInvestigating
	entity.InvestigatedBy = &cmd.InvestigatedBy
	if cmd.InvestigationNotes != "" {
		entity.InvestigationNotes = &cmd.InvestigationNotes
	}
	if err := h.writeRepo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("investigate aml alert: %w", err)
	}

	if err := h.eventBus.Publish(ctx, aml_alert.AlertInvestigatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
	}); err != nil {
		return fmt.Errorf("publish aml_alert.investigated: %w", err)
	}

	return nil
}
