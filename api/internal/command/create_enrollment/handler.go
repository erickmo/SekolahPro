// Package create_enrollment menangani command untuk membuat EducationEnrollment baru.
package create_enrollment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/enrollment"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_enrollment"

// Command berisi data yang dibutuhkan untuk membuat EducationEnrollment baru.
type Command struct {
	CourseID  uuid.UUID
	NasabahID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_enrollment.
type Handler struct {
	repo     enrollment.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo enrollment.WriteRepository, eb eventbus.EventBus) *Handler {
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

	if cmd.CourseID == uuid.Nil {
		return enrollment.ErrCourseRequired
	}
	if cmd.NasabahID == uuid.Nil {
		return enrollment.ErrNasabahRequired
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &enrollment.EducationEnrollment{
		ID:         id,
		CourseID:   cmd.CourseID,
		NasabahID:  cmd.NasabahID,
		Status:     enrollment.StatusEnrolled,
		EnrolledAt: now,
		Score:      0,
		Attempts:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.repo.SaveEnrollment(ctx, s, entity); err != nil {
		return fmt.Errorf("save enrollment: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, enrollment.EnrollmentCreatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        id,
		CourseID:  cmd.CourseID,
		NasabahID: cmd.NasabahID,
	}); err != nil {
		return fmt.Errorf("publish education_enrollment.created: %w", err)
	}

	return nil
}
