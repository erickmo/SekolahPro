// Package create_branch_report menangani command untuk membuat BranchFinancialSummary baru.
package create_branch_report

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_branch_report"

// Command berisi data yang dibutuhkan untuk membuat BranchFinancialSummary baru.
type Command struct {
	BranchID               *uuid.UUID
	PeriodType             string
	PeriodStart            time.Time
	PeriodEnd              time.Time
	PendapatanOperasional  int64
	PendapatanBungaMargin  int64
	PendapatanLain         int64
	TotalPendapatan        int64
	BiayaOperasional       int64
	BiayaPersonel          int64
	BiayaAdministrasi      int64
	BebanPPAP              int64
	TotalBiaya             int64
	LabaRugiBersih         int64
	TotalAset              int64
	KasDanBank             int64
	PinjamanDiberikan      int64
	SimpananDiterima       int64
	TotalKewajiban         int64
	ModalSendiri           int64
	NPLRatio               *float64
	BOPORatio              *float64
	ROA                    *float64
	CAR                    *float64
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_branch_report.
type Handler struct {
	repo     branch_report.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo branch_report.WriteRepository, eb eventbus.EventBus) *Handler {
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

	pt, ok := branch_report.ValidPeriodTypes[cmd.PeriodType]
	if !ok {
		return branch_report.ErrInvalidPeriodType
	}

	if !cmd.PeriodStart.Before(cmd.PeriodEnd) {
		return branch_report.ErrInvalidPeriodRange
	}

	exists, err := h.repo.ExistsByPeriod(ctx, s, cmd.BranchID, pt, cmd.PeriodStart)
	if err != nil {
		return fmt.Errorf("check duplicate period: %w", err)
	}
	if exists {
		return branch_report.ErrDuplicatePeriod
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &branch_report.BranchFinancialSummary{
		ID:                     id,
		BranchID:               cmd.BranchID,
		PeriodType:             pt,
		PeriodStart:            cmd.PeriodStart,
		PeriodEnd:              cmd.PeriodEnd,
		PendapatanOperasional:  cmd.PendapatanOperasional,
		PendapatanBungaMargin:  cmd.PendapatanBungaMargin,
		PendapatanLain:         cmd.PendapatanLain,
		TotalPendapatan:        cmd.TotalPendapatan,
		BiayaOperasional:       cmd.BiayaOperasional,
		BiayaPersonel:          cmd.BiayaPersonel,
		BiayaAdministrasi:      cmd.BiayaAdministrasi,
		BebanPPAP:              cmd.BebanPPAP,
		TotalBiaya:             cmd.TotalBiaya,
		LabaRugiBersih:         cmd.LabaRugiBersih,
		TotalAset:              cmd.TotalAset,
		KasDanBank:             cmd.KasDanBank,
		PinjamanDiberikan:      cmd.PinjamanDiberikan,
		SimpananDiterima:       cmd.SimpananDiterima,
		TotalKewajiban:         cmd.TotalKewajiban,
		ModalSendiri:           cmd.ModalSendiri,
		NPLRatio:               cmd.NPLRatio,
		BOPORatio:              cmd.BOPORatio,
		ROA:                    cmd.ROA,
		CAR:                    cmd.CAR,
		CalculatedAt:           now,
		CreatedAt:              now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save branch report: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, branch_report.BranchReportGeneratedEvent{
		TenantID:   s.TenantID,
		CompanyID:  s.CompanyID,
		ID:         id,
		BranchID:   cmd.BranchID,
		PeriodType: pt,
	}); err != nil {
		return fmt.Errorf("publish branch_report.generated: %w", err)
	}

	return nil
}
