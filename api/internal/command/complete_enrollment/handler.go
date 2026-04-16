// Package complete_enrollment menangani command untuk menyelesaikan EducationEnrollment.
package complete_enrollment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/enrollment"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "complete_enrollment"

// Command berisi data yang dibutuhkan untuk menyelesaikan EducationEnrollment.
type Command struct {
	ID             uuid.UUID
	Score          float64
	CertificateURL string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command complete_enrollment.
type Handler struct {
	readRepo  enrollment.ReadRepository
	writeRepo enrollment.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr enrollment.ReadRepository, wr enrollment.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengambil entity yang ada, mengupdate status, dan mempublish event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}
	if cmd.Score < 0 || cmd.Score > 100 {
		return enrollment.ErrInvalidScore
	}

	entity, err := h.readRepo.GetEnrollmentByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get enrollment for completion: %w", err)
	}

	now := time.Now().UTC()
	if cmd.Score >= 70 {
		entity.Status = enrollment.StatusCompleted
	} else {
		entity.Status = enrollment.StatusFailed
	}
	entity.Score = cmd.Score
	entity.CertificateURL = cmd.CertificateURL
	entity.CompletedAt = &now
	entity.UpdatedAt = now

	if err := h.writeRepo.UpdateEnrollment(ctx, s, entity); err != nil {
		return fmt.Errorf("complete enrollment: %w", err)
	}

	if err := h.eventBus.Publish(ctx, enrollment.EnrollmentCompletedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
		CourseID:  entity.CourseID,
		NasabahID: entity.NasabahID,
		Score:     entity.Score,
	}); err != nil {
		return fmt.Errorf("publish education_enrollment.completed: %w", err)
	}

	return nil
}
