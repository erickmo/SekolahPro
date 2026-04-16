// Package withdraw_consent menangani command untuk mencabut consent.
package withdraw_consent

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/consent"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "withdraw_consent"

// Command berisi data yang dibutuhkan untuk mencabut consent.
type Command struct {
	ID               uuid.UUID
	WithdrawalReason *string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command withdraw_consent.
type Handler struct {
	readRepo  consent.ReadRepository
	writeRepo consent.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr consent.ReadRepository, wr consent.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi dan mencabut consent.
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
		return fmt.Errorf("get consent record for withdraw: %w", err)
	}

	if entity.Withdrawn {
		return consent.ErrAlreadyWithdrawn
	}

	now := time.Now().UTC()
	entity.Withdrawn = true
	entity.WithdrawnAt = &now
	entity.WithdrawalReason = cmd.WithdrawalReason
	entity.ConsentGiven = false

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("withdraw consent: %w", err)
	}

	if err := h.eventBus.Publish(ctx, consent.ConsentWithdrawnEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
	}); err != nil {
		return fmt.Errorf("publish consent_record.withdrawn: %w", err)
	}

	return nil
}
