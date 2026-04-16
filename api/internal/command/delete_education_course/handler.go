// Package delete_education_course menangani command untuk soft-delete EducationCourse.
package delete_education_course

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/education"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "delete_education_course"

// Command berisi data untuk menghapus EducationCourse.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command delete_education_course.
type Handler struct {
	repo education.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo education.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle melakukan soft delete dan memvalidasi scope.
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
	return h.repo.Delete(ctx, s, cmd.ID)
}
