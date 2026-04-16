// Package enroll_biometric menangani command untuk membuat biometric enrollment baru.
package enroll_biometric

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/biometric"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "enroll_biometric"

// Command berisi data yang dibutuhkan untuk membuat biometric enrollment baru.
type Command struct {
	NasabahID     uuid.UUID
	BiometricType biometric.BiometricType
	DeviceInfo    string
	TemplateHash  string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command enroll_biometric.
type Handler struct {
	repo     biometric.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo biometric.WriteRepository, eb eventbus.EventBus) *Handler {
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
		return biometric.ErrNasabahEmpty
	}
	if cmd.TemplateHash == "" {
		return biometric.ErrTemplateHashEmpty
	}
	if !isValidBiometricType(cmd.BiometricType) {
		return biometric.ErrInvalidBiometricType
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &biometric.BiometricEnrollment{
		ID:             id,
		NasabahID:      cmd.NasabahID,
		BiometricType:  cmd.BiometricType,
		DeviceInfo:     cmd.DeviceInfo,
		TemplateHash:   cmd.TemplateHash,
		IsActive:       true,
		FailedAttempts: 0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save biometric enrollment: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, biometric.BiometricEnrolledEvent{
		TenantID:      s.TenantID,
		CompanyID:     s.CompanyID,
		ID:            id,
		NasabahID:     cmd.NasabahID,
		BiometricType: cmd.BiometricType,
	}); err != nil {
		return fmt.Errorf("publish biometric.enrolled: %w", err)
	}

	return nil
}

func isValidBiometricType(t biometric.BiometricType) bool {
	switch t {
	case biometric.BiometricTypeFingerprint, biometric.BiometricTypeFace, biometric.BiometricTypeVoice:
		return true
	default:
		return false
	}
}
