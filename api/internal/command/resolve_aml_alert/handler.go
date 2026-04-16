// Package resolve_aml_alert menangani command untuk menyelesaikan alert.
package resolve_aml_alert

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_alert"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "resolve_aml_alert"

// Command berisi data yang dibutuhkan untuk menyelesaikan alert.
type Command struct {
	ID                   uuid.UUID
	InvestigationOutcome string
	InvestigationNotes   string
	LTKMFiled           bool
	LTKMReference        *string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command resolve_aml_alert.
type Handler struct {
	readRepo  aml_alert.ReadRepository
	writeRepo aml_alert.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr aml_alert.ReadRepository, wr aml_alert.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi dan menyelesaikan alert.
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
		return fmt.Errorf("get aml alert for resolve: %w", err)
	}

	if entity.Status == aml_alert.AlertStatusResolved {
		return aml_alert.ErrAlreadyResolved
	}
	if entity.Status != aml_alert.AlertStatusInvestigating && entity.Status != aml_alert.AlertStatusEscalated {
		return fmt.Errorf("alert dengan status %s tidak dapat diresolusi", entity.Status)
	}

	if cmd.InvestigationOutcome == "" {
		return fmt.Errorf("investigation_outcome tidak boleh kosong")
	}

	now := time.Now().UTC()
	entity.Status = aml_alert.AlertStatusResolved
	outcome := cmd.InvestigationOutcome
	entity.InvestigationOutcome = &outcome
	if cmd.InvestigationNotes != "" {
		entity.InvestigationNotes = &cmd.InvestigationNotes
	}
	entity.LTKMFiled = cmd.LTKMFiled
	entity.LTKMReference = cmd.LTKMReference
	entity.ResolvedAt = &now

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("resolve aml alert: %w", err)
	}

	if err := h.eventBus.Publish(ctx, aml_alert.AlertResolvedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
		Outcome:   cmd.InvestigationOutcome,
	}); err != nil {
		return fmt.Errorf("publish aml_alert.resolved: %w", err)
	}

	return nil
}
