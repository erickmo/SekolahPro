// Package update_aging_bucket menangani command untuk mengupdate DPD dan aging bucket.
package update_aging_bucket

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_aging_bucket"

// Command berisi data untuk mengupdate DPD dan menghitung ulang aging bucket.
type Command struct {
	ID                 uuid.UUID
	NewDPD             int
	TotalOverdueAmount int64
	OverdueInstallments int
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_aging_bucket.
type Handler struct {
	readRepo  collection.CaseReadRepository
	writeRepo collection.CaseWriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr collection.CaseReadRepository, wr collection.CaseWriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle mengambil case, menghitung ulang DPD dan bucket, lalu menyimpan.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}
	if cmd.NewDPD < 0 {
		return collection.ErrNegativeDPD
	}

	entity, err := h.readRepo.GetCaseByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get collection case for aging update: %w", err)
	}

	if entity.Status != collection.CaseStatusActive &&
		entity.Status != collection.CaseStatusRestructured {
		return collection.ErrCaseAlreadyResolved
	}

	oldBucket := entity.AgingBucket
	newBucket := collection.CalculateAgingBucket(cmd.NewDPD)

	entity.CurrentDPD = cmd.NewDPD
	entity.AgingBucket = newBucket
	entity.TotalOverdueAmount = cmd.TotalOverdueAmount
	entity.TotalOverdueInstallments = cmd.OverdueInstallments

	if err := h.writeRepo.UpdateCase(ctx, s, entity); err != nil {
		return fmt.Errorf("update collection case aging: %w", err)
	}

	if oldBucket != newBucket {
		if err := h.eventBus.Publish(ctx, collection.AgingBucketChangedEvent{
			TenantID:   s.TenantID,
			CompanyID:  s.CompanyID,
			ID:         entity.ID,
			OldBucket:  string(oldBucket),
			NewBucket:  string(newBucket),
			CurrentDPD: cmd.NewDPD,
		}); err != nil {
			return fmt.Errorf("publish collection_case.aging_bucket_changed: %w", err)
		}
	}

	return nil
}
