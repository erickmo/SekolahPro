// Package get_branch_report_by_id menangani query untuk mengambil BranchFinancialSummary berdasarkan ID.
package get_branch_report_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_branch_report_by_id"

// Query berisi parameter untuk mengambil BranchFinancialSummary berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                     string   `json:"id"`
	BranchID               *string  `json:"branch_id"`
	PeriodType             string   `json:"period_type"`
	PeriodStart            string   `json:"period_start"`
	PeriodEnd              string   `json:"period_end"`
	PendapatanOperasional  int64    `json:"pendapatan_operasional"`
	PendapatanBungaMargin  int64    `json:"pendapatan_bunga_margin"`
	PendapatanLain         int64    `json:"pendapatan_lain"`
	TotalPendapatan        int64    `json:"total_pendapatan"`
	BiayaOperasional       int64    `json:"biaya_operasional"`
	BiayaPersonel          int64    `json:"biaya_personel"`
	BiayaAdministrasi      int64    `json:"biaya_administrasi"`
	BebanPPAP              int64    `json:"beban_ppap"`
	TotalBiaya             int64    `json:"total_biaya"`
	LabaRugiBersih         int64    `json:"laba_rugi_bersih"`
	TotalAset              int64    `json:"total_aset"`
	KasDanBank             int64    `json:"kas_dan_bank"`
	PinjamanDiberikan      int64    `json:"pinjaman_diberikan"`
	SimpananDiterima       int64    `json:"simpanan_diterima"`
	TotalKewajiban         int64    `json:"total_kewajiban"`
	ModalSendiri           int64    `json:"modal_sendiri"`
	NPLRatio               *float64 `json:"npl_ratio"`
	BOPORatio              *float64 `json:"bopo_ratio"`
	ROA                    *float64 `json:"roa"`
	CAR                    *float64 `json:"car"`
	CalculatedAt           string   `json:"calculated_at"`
	ApprovedBy             *string  `json:"approved_by"`
	CreatedAt              string   `json:"created_at"`
}

// Handler menangani Query get_branch_report_by_id.
type Handler struct {
	repo branch_report.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo branch_report.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil BranchFinancialSummary berdasarkan ID dengan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	entity, err := h.repo.GetByID(ctx, s, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("get branch report by id: %w", err)
	}

	return toResult(entity), nil
}

func toResult(e *branch_report.BranchFinancialSummary) *Result {
	r := &Result{
		ID:                     e.ID.String(),
		PeriodType:             string(e.PeriodType),
		PeriodStart:            e.PeriodStart.Format("2006-01-02"),
		PeriodEnd:              e.PeriodEnd.Format("2006-01-02"),
		PendapatanOperasional:  e.PendapatanOperasional,
		PendapatanBungaMargin:  e.PendapatanBungaMargin,
		PendapatanLain:         e.PendapatanLain,
		TotalPendapatan:        e.TotalPendapatan,
		BiayaOperasional:       e.BiayaOperasional,
		BiayaPersonel:          e.BiayaPersonel,
		BiayaAdministrasi:      e.BiayaAdministrasi,
		BebanPPAP:              e.BebanPPAP,
		TotalBiaya:             e.TotalBiaya,
		LabaRugiBersih:         e.LabaRugiBersih,
		TotalAset:              e.TotalAset,
		KasDanBank:             e.KasDanBank,
		PinjamanDiberikan:      e.PinjamanDiberikan,
		SimpananDiterima:       e.SimpananDiterima,
		TotalKewajiban:         e.TotalKewajiban,
		ModalSendiri:           e.ModalSendiri,
		NPLRatio:               e.NPLRatio,
		BOPORatio:              e.BOPORatio,
		ROA:                    e.ROA,
		CAR:                    e.CAR,
		CalculatedAt:           e.CalculatedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:              e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if e.BranchID != nil {
		bid := e.BranchID.String()
		r.BranchID = &bid
	}
	if e.ApprovedBy != nil {
		aid := e.ApprovedBy.String()
		r.ApprovedBy = &aid
	}

	return r
}
