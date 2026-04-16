// Package create_collection_case menangani command untuk membuat CollectionCase baru.
package create_collection_case

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_collection_case"

// Command berisi data yang dibutuhkan untuk membuat CollectionCase baru.
type Command struct {
	PinjamanID               uuid.UUID
	NasabahID                uuid.UUID
	CaseNumber               string
	CurrentDPD               int
	TotalOverdueAmount       int64
	TotalOverdueInstallments int
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_collection_case.
type Handler struct {
	repo     collection.CaseWriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo collection.CaseWriteRepository, eb eventbus.EventBus) *Handler {
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

	if cmd.CaseNumber == "" {
		return collection.ErrCaseNumberEmpty
	}
	if cmd.CurrentDPD < 0 {
		return collection.ErrNegativeDPD
	}
	if cmd.TotalOverdueAmount < 0 {
		return collection.ErrNegativeOverdue
	}

	id := uuid.New()
	now := time.Now().UTC()
	bucket := collection.CalculateAgingBucket(cmd.CurrentDPD)

	entity := &collection.CollectionCase{
		ID:                       id,
		CaseNumber:               cmd.CaseNumber,
		CurrentDPD:               cmd.CurrentDPD,
		AgingBucket:              bucket,
		TotalOverdueAmount:       cmd.TotalOverdueAmount,
		TotalOverdueInstallments: cmd.TotalOverdueInstallments,
		PinjamanID:               cmd.PinjamanID,
		NasabahID:                cmd.NasabahID,
		EscalationLevel:          0,
		Status:                   collection.CaseStatusActive,
		OpenedAt:                 now,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	if err := h.repo.SaveCase(ctx, s, entity); err != nil {
		return fmt.Errorf("save collection case: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, collection.CollectionCaseCreatedEvent{
		TenantID:    s.TenantID,
		CompanyID:   s.CompanyID,
		ID:          id,
		CaseNumber:  cmd.CaseNumber,
		PinjamanID:  cmd.PinjamanID,
		NasabahID:   cmd.NasabahID,
		AgingBucket: string(bucket),
	}); err != nil {
		return fmt.Errorf("publish collection_case.created: %w", err)
	}

	return nil
}
