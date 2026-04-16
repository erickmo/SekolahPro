// Package assign_collector menangani command untuk menugaskan collector ke collection case.
package assign_collector

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "assign_collector"

// Command berisi data untuk menugaskan collector ke collection case.
type Command struct {
	ID         uuid.UUID
	CollectorID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command assign_collector.
type Handler struct {
	readRepo  collection.CaseReadRepository
	writeRepo collection.CaseWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr collection.CaseReadRepository, wr collection.CaseWriteRepository) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr}
}

// Handle mengambil case, mengupdate assignment collector, lalu menyimpan.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}
	if cmd.CollectorID == uuid.Nil {
		return fmt.Errorf("collector_id tidak boleh kosong")
	}

	entity, err := h.readRepo.GetCaseByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get collection case for assignment: %w", err)
	}

	now := time.Now().UTC()
	entity.AssignedCollectorID = &cmd.CollectorID
	entity.AssignedAt = &now

	if err := h.writeRepo.UpdateCase(ctx, s, entity); err != nil {
		return fmt.Errorf("update collection case assignment: %w", err)
	}

	return nil
}
