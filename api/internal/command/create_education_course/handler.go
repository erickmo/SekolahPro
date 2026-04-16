// Package create_education_course menangani command untuk membuat EducationCourse baru.
package create_education_course

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/education"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_education_course"

// Command berisi data yang dibutuhkan untuk membuat EducationCourse baru.
type Command struct {
	Title         string
	Description   string
	CourseType    education.CourseType
	Category      string
	DurationHours int
	ContentURL    string
	PassingScore  float64
	MaxAttempts   int
	MandatoryFor  []string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_education_course.
type Handler struct {
	repo     education.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo education.WriteRepository, eb eventbus.EventBus) *Handler {
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

	if cmd.Title == "" {
		return education.ErrTitleEmpty
	}
	if len(cmd.Title) > 255 {
		return education.ErrTitleTooLong
	}
	if !isValidCourseType(cmd.CourseType) {
		return education.ErrInvalidCourseType
	}
	if cmd.PassingScore < 0 || cmd.PassingScore > 100 {
		return education.ErrInvalidScoreRange
	}
	if cmd.MaxAttempts <= 0 {
		return education.ErrInvalidAttempts
	}

	id := uuid.New()
	now := time.Now().UTC()

	mandatoryFor := cmd.MandatoryFor
	if mandatoryFor == nil {
		mandatoryFor = []string{}
	}

	entity := &education.EducationCourse{
		ID:            id,
		Title:         cmd.Title,
		Description:   cmd.Description,
		CourseType:    cmd.CourseType,
		Category:      cmd.Category,
		DurationHours: cmd.DurationHours,
		ContentURL:    cmd.ContentURL,
		IsActive:      true,
		PassingScore:  cmd.PassingScore,
		MaxAttempts:   cmd.MaxAttempts,
		MandatoryFor:  mandatoryFor,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save education course: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, education.CourseCreatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        id,
		Title:     cmd.Title,
	}); err != nil {
		return fmt.Errorf("publish education_course.created: %w", err)
	}

	return nil
}

func isValidCourseType(t education.CourseType) bool {
	switch t {
	case education.CourseTypeMandatory, education.CourseTypeElective, education.CourseTypeCertification:
		return true
	default:
		return false
	}
}
