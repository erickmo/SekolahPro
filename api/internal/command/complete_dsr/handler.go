// Package complete_dsr menangani command untuk menyelesaikan DataSubjectRequest.
package complete_dsr

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dsr"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "complete_dsr"

// Command berisi data untuk menyelesaikan DataSubjectRequest.
type Command struct {
	ID           uuid.UUID
	CompletedBy  uuid.UUID
	ResponseData string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command complete_dsr.
type Handler struct {
	repo dsr.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo dsr.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle menyelesaikan DataSubjectRequest dengan menyimpan response data.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	readRepo, ok := h.repo.(dsr.ReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement ReadRepository")
	}

	entity, err := readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get dsr: %w", err)
	}

	if entity.Status == dsr.RequestStatusCompleted {
		return dsr.ErrAlreadyCompleted
	}

	now := time.Now().UTC()
	entity.Status = dsr.RequestStatusCompleted
	entity.CompletedBy = &cmd.CompletedBy
	entity.CompletedAt = &now
	entity.ResponseData = cmd.ResponseData
	entity.UpdatedAt = now

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("complete dsr: %w", err)
	}

	return nil
}
