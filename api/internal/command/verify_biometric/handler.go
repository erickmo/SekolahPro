// Package verify_biometric menangani command untuk verifikasi biometric.
package verify_biometric

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/biometric"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const (
	commandName       = "verify_biometric"
	maxFailedAttempts = 3
)

// Command berisi data yang dibutuhkan untuk verifikasi biometric.
type Command struct {
	ID          uuid.UUID
	IsVerified  bool
	VerifiedBy  uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command verify_biometric.
type Handler struct {
	readRepo  biometric.ReadRepository
	writeRepo biometric.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr biometric.ReadRepository, wr biometric.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengambil enrollment, melakukan verifikasi, dan mempublish event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}
	if cmd.ID == uuid.Nil {
		return fmt.Errorf("ID tidak boleh kosong")
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get biometric enrollment for verify: %w", err)
	}

	if !entity.IsActive {
		return biometric.ErrNotActive
	}

	now := time.Now().UTC()
	entity.LastAttemptAt = &now

	if cmd.IsVerified {
		entity.VerifiedAt = &now
		entity.VerifiedBy = &cmd.VerifiedBy
		entity.FailedAttempts = 0
	} else {
		entity.FailedAttempts++
		if entity.FailedAttempts >= maxFailedAttempts {
			entity.IsActive = false
		}
	}

	entity.UpdatedAt = now

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update biometric enrollment: %w", err)
	}

	if cmd.IsVerified {
		if err := h.eventBus.Publish(ctx, biometric.BiometricVerifiedEvent{
			TenantID:      s.TenantID,
			CompanyID:     s.CompanyID,
			ID:            entity.ID,
			NasabahID:     entity.NasabahID,
			BiometricType: entity.BiometricType,
		}); err != nil {
			return fmt.Errorf("publish biometric.verified: %w", err)
		}
	}

	return nil
}
