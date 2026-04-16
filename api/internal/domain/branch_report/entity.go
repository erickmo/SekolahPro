// Package branch_report mendefinisikan domain untuk laporan keuangan multi-cabang koperasi.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package branch_report

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// PeriodType mendefinisikan jenis periode laporan.
type PeriodType string

const (
	PeriodDaily     PeriodType = "daily"
	PeriodWeekly    PeriodType = "weekly"
	PeriodMonthly   PeriodType = "monthly"
	PeriodQuarterly PeriodType = "quarterly"
	PeriodYearly    PeriodType = "yearly"
)

// ValidPeriodTypes memetakan string ke PeriodType yang valid.
var ValidPeriodTypes = map[string]PeriodType{
	string(PeriodDaily):     PeriodDaily,
	string(PeriodWeekly):    PeriodWeekly,
	string(PeriodMonthly):   PeriodMonthly,
	string(PeriodQuarterly): PeriodQuarterly,
	string(PeriodYearly):    PeriodYearly,
}

// RuleType mendefinisikan jenis elimination rule.
type RuleType string

const (
	RuleInterBranchTransfer RuleType = "inter_branch_transfer"
	RuleLoanParticipation   RuleType = "loan_participation"
	RuleProfitElimination   RuleType = "profit_elimination"
)

// ValidRuleTypes memetakan string ke RuleType yang valid.
var ValidRuleTypes = map[string]RuleType{
	string(RuleInterBranchTransfer): RuleInterBranchTransfer,
	string(RuleLoanParticipation):   RuleLoanParticipation,
	string(RuleProfitElimination):   RuleProfitElimination,
}

// BranchFinancialSummary adalah entity utama untuk laporan keuangan cabang.
// branch_id = null berarti laporan konsolidasi (agregasi semua cabang).
type BranchFinancialSummary struct {
	ID uuid.UUID `db:"id"`

	// Period
	BranchID    *uuid.UUID `db:"branch_id"`
	PeriodType  PeriodType `db:"period_type"`
	PeriodStart time.Time  `db:"period_start"`
	PeriodEnd   time.Time  `db:"period_end"`

	// P&L fields (dalam satuan terkecil, misalnya Rupiah tanpa desimal)
	PendapatanOperasional  int64 `db:"pendapatan_operasional"`
	PendapatanBungaMargin  int64 `db:"pendapatan_bunga_margin"`
	PendapatanLain         int64 `db:"pendapatan_lain"`
	TotalPendapatan        int64 `db:"total_pendapatan"`
	BiayaOperasional       int64 `db:"biaya_operasional"`
	BiayaPersonel          int64 `db:"biaya_personel"`
	BiayaAdministrasi      int64 `db:"biaya_administrasi"`
	BebanPPAP              int64 `db:"beban_ppap"`
	TotalBiaya             int64 `db:"total_biaya"`
	LabaRugiBersih         int64 `db:"laba_rugi_bersih"`

	// Balance Sheet fields
	TotalAset       int64 `db:"total_aset"`
	KasDanBank      int64 `db:"kas_dan_bank"`
	PinjamanDiberikan int64 `db:"pinjaman_diberikan"`
	SimpananDiterima int64  `db:"simpanan_diterima"`
	TotalKewajiban   int64 `db:"total_kewajiban"`
	ModalSendiri     int64 `db:"modal_sendiri"`

	// KPI fields (nullable karena hanya dihitung saat consolidated atau on-demand)
	NPLRatio  *float64 `db:"npl_ratio"`
	BOPORatio *float64 `db:"bopo_ratio"`
	ROA       *float64 `db:"roa"`
	CAR       *float64 `db:"car"`

	// Audit
	CalculatedAt time.Time   `db:"calculated_at"`
	ApprovedBy   *uuid.UUID  `db:"approved_by"`
	CreatedAt    time.Time   `db:"created_at"`
	DeletedAt    *time.Time  `db:"deleted_at"`
}

// EliminationRule mendefinisikan aturan eliminasi untuk consolidated reporting.
type EliminationRule struct {
	ID           uuid.UUID  `db:"id"`
	RuleName     string     `db:"rule_name"`
	RuleType     RuleType   `db:"rule_type"`
	FromBranchID uuid.UUID  `db:"from_branch_id"`
	ToBranchID   uuid.UUID  `db:"to_branch_id"`
	IsActive     bool       `db:"is_active"`
	Description  string     `db:"description"`
	CreatedAt    time.Time  `db:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk BranchFinancialSummary.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *BranchFinancialSummary) error
	Update(ctx context.Context, s scope.Scope, e *BranchFinancialSummary) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
	SetApprovedBy(ctx context.Context, s scope.Scope, id uuid.UUID, approvedBy uuid.UUID) error
	ExistsByPeriod(ctx context.Context, s scope.Scope, branchID *uuid.UUID, periodType PeriodType, periodStart time.Time) (bool, error)
}

// ReadRepository mendefinisikan operasi read untuk BranchFinancialSummary.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*BranchFinancialSummary, error)
	List(ctx context.Context, s scope.Scope, filter ListFilter, limit, offset int, sortBy, order string) ([]*BranchFinancialSummary, int, error)
	GetConsolidated(ctx context.Context, s scope.Scope, periodType PeriodType, periodStart time.Time) (*ConsolidatedRow, error)
}

// EliminationRuleWriteRepository mendefinisikan operasi write untuk EliminationRule.
type EliminationRuleWriteRepository interface {
	SaveRule(ctx context.Context, s scope.Scope, e *EliminationRule) error
	ToggleActive(ctx context.Context, s scope.Scope, id uuid.UUID) (bool, error)
}

// EliminationRuleReadRepository mendefinisikan operasi read untuk EliminationRule.
type EliminationRuleReadRepository interface {
	ListRules(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*EliminationRule, int, error)
}

// ListFilter berisi parameter filter untuk list query.
type ListFilter struct {
	BranchID    *uuid.UUID
	PeriodType  *PeriodType
	PeriodStart *time.Time
	PeriodEnd   *time.Time
	ApprovedBy  *uuid.UUID
}

// ConsolidatedRow adalah struct untuk scan hasil query konsolidasi.
type ConsolidatedRow struct {
	TenantID              uuid.UUID    `db:"tenant_id"`
	CompanyID             uuid.UUID    `db:"company_id"`
	PeriodType            PeriodType   `db:"period_type"`
	PeriodStart           time.Time    `db:"period_start"`
	PeriodEnd             time.Time    `db:"period_end"`
	PendapatanOperasional int64        `db:"pendapatan_operasional"`
	PendapatanBungaMargin int64        `db:"pendapatan_bunga_margin"`
	PendapatanLain        int64        `db:"pendapatan_lain"`
	TotalPendapatan       int64        `db:"total_pendapatan"`
	BiayaOperasional      int64        `db:"biaya_operasional"`
	BiayaPersonel         int64        `db:"biaya_personel"`
	BiayaAdministrasi     int64        `db:"biaya_administrasi"`
	BebanPPAP             int64        `db:"beban_ppap"`
	TotalBiaya            int64        `db:"total_biaya"`
	LabaRugiBersih        int64        `db:"laba_rugi_bersih"`
	TotalAset             int64        `db:"total_aset"`
	KasDanBank            int64        `db:"kas_dan_bank"`
	PinjamanDiberikan     int64        `db:"pinjaman_diberikan"`
	SimpananDiterima      int64        `db:"simpanan_diterima"`
	TotalKewajiban        int64        `db:"total_kewajiban"`
	ModalSendiri          int64        `db:"modal_sendiri"`
	NPLRatio              sql.NullFloat64 `db:"npl_ratio"`
	BOPORatio             sql.NullFloat64 `db:"bopo_ratio"`
	ROA                   sql.NullFloat64 `db:"roa"`
	CAR                   sql.NullFloat64 `db:"car"`
	BranchCount           int          `db:"branch_count"`
}
