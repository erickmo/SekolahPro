// Package resolve_collection_case menangani command untuk meresolusi collection case.
package resolve_collection_case

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "resolve_collection_case"

// Command berisi data untuk meresolusi collection case.
type Command struct {
	ID             uuid.UUID
	ResolutionType collection.ResolutionType
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command resolve_collection_case.
type Handler struct {
	readRepo  collection.CaseReadRepository
	writeRepo collection.CaseWriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr collection.CaseReadRepository, wr collection.CaseWriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle mengambil case, mengupdate status dan resolution, lalu menyimpan dan publish event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	entity, err := h.readRepo.GetCaseByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get collection case for resolution: %w", err)
	}

	if entity.Status == collection.CaseStatusSettled ||
		entity.Status == collection.CaseStatusClosed ||
		entity.Status == collection.CaseStatusWrittenOff {
		return collection.ErrCaseAlreadyResolved
	}

	now := time.Now().UTC()

	switch cmd.ResolutionType {
	case collection.ResolutionRestructuring:
		entity.Status = collection.CaseStatusRestructured
	case collection.ResolutionWriteOff:
		entity.Status = collection.CaseStatusWrittenOff
	default:
		entity.Status = collection.CaseStatusSettled
	}

	entity.ResolutionType = &cmd.ResolutionType
	entity.ClosedAt = &now

	if err := h.writeRepo.UpdateCase(ctx, s, entity); err != nil {
		return fmt.Errorf("update collection case resolution: %w", err)
	}

	if err := h.eventBus.Publish(ctx, collection.CollectionCaseResolvedEvent{
		TenantID:       s.TenantID,
		CompanyID:      s.CompanyID,
		ID:             entity.ID,
		ResolutionType: string(cmd.ResolutionType),
	}); err != nil {
		return fmt.Errorf("publish collection_case.resolved: %w", err)
	}

	return nil
}
