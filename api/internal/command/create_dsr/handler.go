// Package create_dsr menangani command untuk membuat DataSubjectRequest baru.
package create_dsr

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dsr"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_dsr"

// Command berisi data yang dibutuhkan untuk membuat DataSubjectRequest baru.
type Command struct {
	RequestType    dsr.RequestType
	RequestorName  string
	RequestorEmail string
	SubjectID      uuid.UUID
	SubjectType    string
	Description    string
	DueDate        time.Time
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_dsr.
type Handler struct {
	repo dsr.WriteRepository
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo dsr.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle memvalidasi command, membuat entity, dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.RequestType == "" {
		return dsr.ErrRequestTypeEmpty
	}
	if cmd.RequestorName == "" {
		return dsr.ErrRequestorEmpty
	}
	if cmd.SubjectID == uuid.Nil {
		return dsr.ErrSubjectEmpty
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &dsr.DataSubjectRequest{
		ID:             id,
		RequestType:    cmd.RequestType,
		Status:         dsr.RequestStatusPending,
		RequestorName:  cmd.RequestorName,
		RequestorEmail: cmd.RequestorEmail,
		SubjectID:      cmd.SubjectID,
		SubjectType:    cmd.SubjectType,
		Description:    cmd.Description,
		DueDate:        cmd.DueDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save dsr: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	return nil
}
