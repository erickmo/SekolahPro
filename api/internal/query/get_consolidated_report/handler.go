// Package get_consolidated_report menangani query untuk mengambil laporan konsolidasi (agregasi semua cabang).
package get_consolidated_report

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_consolidated_report"

// Query berisi parameter untuk mengambil laporan konsolidasi.
type Query struct {
	PeriodType  branch_report.PeriodType
	PeriodStart time.Time
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query konsolidasi.
type Result struct {
	PeriodType             string   `json:"period_type"`
	PeriodStart            string   `json:"period_start"`
	PeriodEnd              string   `json:"period_end"`
	BranchCount            int      `json:"branch_count"`
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
}

// Handler menangani Query get_consolidated_report.
type Handler struct {
	repo branch_report.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo branch_report.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil laporan konsolidasi yang merupakan agregasi semua cabang untuk periode tertentu.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	cr, err := h.repo.GetConsolidated(ctx, s, qry.PeriodType, qry.PeriodStart)
	if err != nil {
		return nil, fmt.Errorf("get consolidated report: %w", err)
	}

	result := &Result{
		PeriodType:             string(cr.PeriodType),
		PeriodStart:            cr.PeriodStart.Format("2006-01-02"),
		PeriodEnd:              cr.PeriodEnd.Format("2006-01-02"),
		BranchCount:            cr.BranchCount,
		PendapatanOperasional:  cr.PendapatanOperasional,
		PendapatanBungaMargin:  cr.PendapatanBungaMargin,
		PendapatanLain:         cr.PendapatanLain,
		TotalPendapatan:        cr.TotalPendapatan,
		BiayaOperasional:       cr.BiayaOperasional,
		BiayaPersonel:          cr.BiayaPersonel,
		BiayaAdministrasi:      cr.BiayaAdministrasi,
		BebanPPAP:              cr.BebanPPAP,
		TotalBiaya:             cr.TotalBiaya,
		LabaRugiBersih:         cr.LabaRugiBersih,
		TotalAset:              cr.TotalAset,
		KasDanBank:             cr.KasDanBank,
		PinjamanDiberikan:      cr.PinjamanDiberikan,
		SimpananDiterima:       cr.SimpananDiterima,
		TotalKewajiban:         cr.TotalKewajiban,
		ModalSendiri:           cr.ModalSendiri,
	}

	if cr.NPLRatio.Valid {
		result.NPLRatio = &cr.NPLRatio.Float64
	}
	if cr.BOPORatio.Valid {
		result.BOPORatio = &cr.BOPORatio.Float64
	}
	if cr.ROA.Valid {
		result.ROA = &cr.ROA.Float64
	}
	if cr.CAR.Valid {
		result.CAR = &cr.CAR.Float64
	}

	return result, nil
}
