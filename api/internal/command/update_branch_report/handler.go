// Package update_branch_report menangani command untuk mengupdate BranchFinancialSummary yang sudah ada.
package update_branch_report

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_branch_report"

// Command berisi data yang dibutuhkan untuk mengupdate BranchFinancialSummary.
type Command struct {
	ID                     uuid.UUID
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

// Handler menangani Command update_branch_report.
type Handler struct {
	readRepo  branch_report.ReadRepository
	writeRepo branch_report.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr branch_report.ReadRepository, wr branch_report.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengambil entity yang ada, mengupdate, dan mempublish event.
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

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get branch report for update: %w", err)
	}

	if entity.ApprovedBy != nil {
		return branch_report.ErrAlreadyApproved
	}

	entity.PendapatanOperasional = cmd.PendapatanOperasional
	entity.PendapatanBungaMargin = cmd.PendapatanBungaMargin
	entity.PendapatanLain = cmd.PendapatanLain
	entity.TotalPendapatan = cmd.TotalPendapatan
	entity.BiayaOperasional = cmd.BiayaOperasional
	entity.BiayaPersonel = cmd.BiayaPersonel
	entity.BiayaAdministrasi = cmd.BiayaAdministrasi
	entity.BebanPPAP = cmd.BebanPPAP
	entity.TotalBiaya = cmd.TotalBiaya
	entity.LabaRugiBersih = cmd.LabaRugiBersih
	entity.TotalAset = cmd.TotalAset
	entity.KasDanBank = cmd.KasDanBank
	entity.PinjamanDiberikan = cmd.PinjamanDiberikan
	entity.SimpananDiterima = cmd.SimpananDiterima
	entity.TotalKewajiban = cmd.TotalKewajiban
	entity.ModalSendiri = cmd.ModalSendiri
	entity.NPLRatio = cmd.NPLRatio
	entity.BOPORatio = cmd.BOPORatio
	entity.ROA = cmd.ROA
	entity.CAR = cmd.CAR
	entity.CalculatedAt = time.Now().UTC()

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update branch report: %w", err)
	}

	return nil
}
