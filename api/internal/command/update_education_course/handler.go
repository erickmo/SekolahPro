// Package update_education_course menangani command untuk mengupdate EducationCourse yang sudah ada.
package update_education_course

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/education"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_education_course"

// Command berisi data yang dibutuhkan untuk mengupdate EducationCourse.
type Command struct {
	ID            uuid.UUID
	Title         string
	Description   string
	CourseType    education.CourseType
	Category      string
	DurationHours int
	ContentURL    string
	IsActive      bool
	PassingScore  float64
	MaxAttempts   int
	MandatoryFor  []string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_education_course.
type Handler struct {
	readRepo  education.ReadRepository
	writeRepo education.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr education.ReadRepository, wr education.WriteRepository, eb eventbus.EventBus) *Handler {
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
	if cmd.Title == "" {
		return education.ErrTitleEmpty
	}
	if len(cmd.Title) > 255 {
		return education.ErrTitleTooLong
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get education course for update: %w", err)
	}

	entity.Title = cmd.Title
	entity.Description = cmd.Description
	entity.CourseType = cmd.CourseType
	entity.Category = cmd.Category
	entity.DurationHours = cmd.DurationHours
	entity.ContentURL = cmd.ContentURL
	entity.IsActive = cmd.IsActive
	entity.PassingScore = cmd.PassingScore
	entity.MaxAttempts = cmd.MaxAttempts
	entity.MandatoryFor = cmd.MandatoryFor
	entity.UpdatedAt = time.Now().UTC()

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update education course: %w", err)
	}

	if err := h.eventBus.Publish(ctx, education.CourseUpdatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
		Title:     entity.Title,
	}); err != nil {
		return fmt.Errorf("publish education_course.updated: %w", err)
	}

	return nil
}
