// Package complete_dissolution menangani command untuk menyelesaikan proses pembubaran.
package complete_dissolution

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "complete_dissolution"

// Command berisi data yang dibutuhkan untuk menyelesaikan proses pembubaran.
type Command struct {
	ID              uuid.UUID
	TotalAssets     int64
	TotalLiabilities int64
	MemberCount     int
	FinalReportDocID *uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command complete_dissolution.
type Handler struct {
	readRepo  dissolution.ReadRepository
	writeRepo dissolution.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr dissolution.ReadRepository, wr dissolution.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle menghitung distribusi per anggota, mengupdate entity ke status completed.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get dissolution for completion: %w", err)
	}

	if entity.Status == dissolution.StatusCompleted {
		return dissolution.ErrAlreadyClosed
	}
	if entity.Status == dissolution.StatusCancelled {
		return dissolution.ErrCannotCancelCompleted
	}

	netEquity := cmd.TotalAssets - cmd.TotalLiabilities
	if cmd.MemberCount <= 0 {
		return dissolution.ErrMemberCountRequired
	}

	distributionPerMember := netEquity / int64(cmd.MemberCount)
	now := time.Now().UTC()

	entity.TotalAssets = &cmd.TotalAssets
	entity.TotalLiabilities = &cmd.TotalLiabilities
	entity.NetEquity = &netEquity
	entity.DistributionPerMember = &distributionPerMember
	entity.FinalReportDocID = cmd.FinalReportDocID
	entity.Stage = dissolution.StageClosed
	entity.Status = dissolution.StatusCompleted
	entity.ClosedAt = &now

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("complete dissolution: %w", err)
	}

	if err := h.eventBus.Publish(ctx, dissolution.DissolutionCompletedEvent{
		TenantID:              s.TenantID,
		CompanyID:             s.CompanyID,
		ID:                    entity.ID,
		NetEquity:             netEquity,
		DistributionPerMember: distributionPerMember,
	}); err != nil {
		return fmt.Errorf("publish dissolution.completed: %w", err)
	}

	return nil
}
