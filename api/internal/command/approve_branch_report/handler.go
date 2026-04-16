// Package approve_branch_report menangani command untuk menyetujui BranchFinancialSummary.
package approve_branch_report

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "approve_branch_report"

// Command berisi data untuk menyetujui BranchFinancialSummary.
type Command struct {
	ID         uuid.UUID
	ApprovedBy uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command approve_branch_report.
type Handler struct {
	readRepo  branch_report.ReadRepository
	writeRepo branch_report.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr branch_report.ReadRepository, wr branch_report.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengecek approval status, dan menyetujui laporan.
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
	if cmd.ApprovedBy == uuid.Nil {
		return fmt.Errorf("approved_by tidak boleh kosong")
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get branch report for approval: %w", err)
	}

	if entity.ApprovedBy != nil {
		return branch_report.ErrAlreadyApproved
	}

	if entity.BranchID == nil {
		return branch_report.ErrCannotApproveConsolidated
	}

	if err := h.writeRepo.SetApprovedBy(ctx, s, cmd.ID, cmd.ApprovedBy); err != nil {
		return fmt.Errorf("approve branch report: %w", err)
	}

	if err := h.eventBus.Publish(ctx, branch_report.BranchReportApprovedEvent{
		TenantID:   s.TenantID,
		CompanyID:  s.CompanyID,
		ID:         cmd.ID,
		BranchID:   entity.BranchID,
		ApprovedBy: cmd.ApprovedBy,
	}); err != nil {
		return fmt.Errorf("publish branch_report.approved: %w", err)
	}

	return nil
}
