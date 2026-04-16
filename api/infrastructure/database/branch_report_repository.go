package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// BranchReportRepository adalah concrete implementation dari branch_report.WriteRepository,
// branch_report.ReadRepository, branch_report.EliminationRuleWriteRepository,
// dan branch_report.EliminationRuleReadRepository.
type BranchReportRepository struct {
	db *sqlx.DB
}

// NewBranchReportRepository membuat instance baru BranchReportRepository.
func NewBranchReportRepository(db *sqlx.DB) *BranchReportRepository {
	return &BranchReportRepository{db: db}
}

// ── BranchFinancialSummary WriteRepository ──────────────────────────────────

// Save menyimpan entity BranchFinancialSummary baru ke database.
func (r *BranchReportRepository) Save(ctx context.Context, s scope.Scope, e *branch_report.BranchFinancialSummary) error {
	const q = `
		INSERT INTO branch_financial_summaries (
			id, tenant_id, company_id, branch_id,
			period_type, period_start, period_end,
			pendapatan_operasional, pendapatan_bunga_margin, pendapatan_lain, total_pendapatan,
			biaya_operasional, biaya_personel, biaya_administrasi, beban_ppap, total_biaya,
			laba_rugi_bersih,
			total_aset, kas_dan_bank, pinjaman_diberikan, simpanan_diterima,
			total_kewajiban, modal_sendiri,
			npl_ratio, bopo_ratio, roa, car,
			calculated_at, approved_by, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID, e.BranchID,
		e.PeriodType, e.PeriodStart, e.PeriodEnd,
		e.PendapatanOperasional, e.PendapatanBungaMargin, e.PendapatanLain, e.TotalPendapatan,
		e.BiayaOperasional, e.BiayaPersonel, e.BiayaAdministrasi, e.BebanPPAP, e.TotalBiaya,
		e.LabaRugiBersih,
		e.TotalAset, e.KasDanBank, e.PinjamanDiberikan, e.SimpananDiterima,
		e.TotalKewajiban, e.ModalSendiri,
		e.NPLRatio, e.BOPORatio, e.ROA, e.CAR,
		e.CalculatedAt, e.ApprovedBy, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save branch report: %w", err)
	}
	return nil
}

// Update mengupdate BranchFinancialSummary yang sudah ada berdasarkan ID + scope.
func (r *BranchReportRepository) Update(ctx context.Context, s scope.Scope, e *branch_report.BranchFinancialSummary) error {
	const q = `
		UPDATE branch_financial_summaries SET
			pendapatan_operasional=$4, pendapatan_bunga_margin=$5, pendapatan_lain=$6, total_pendapatan=$7,
			biaya_operasional=$8, biaya_personel=$9, biaya_administrasi=$10, beban_ppap=$11, total_biaya=$12,
			laba_rugi_bersih=$13,
			total_aset=$14, kas_dan_bank=$15, pinjaman_diberikan=$16, simpanan_diterima=$17,
			total_kewajiban=$18, modal_sendiri=$19,
			npl_ratio=$20, bopo_ratio=$21, roa=$22, car=$23,
			calculated_at=$24
		WHERE tenant_id=$1 AND company_id=$2 AND id=$3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.PendapatanOperasional, e.PendapatanBungaMargin, e.PendapatanLain, e.TotalPendapatan,
		e.BiayaOperasional, e.BiayaPersonel, e.BiayaAdministrasi, e.BebanPPAP, e.TotalBiaya,
		e.LabaRugiBersih,
		e.TotalAset, e.KasDanBank, e.PinjamanDiberikan, e.SimpananDiterima,
		e.TotalKewajiban, e.ModalSendiri,
		e.NPLRatio, e.BOPORatio, e.ROA, e.CAR,
		e.CalculatedAt,
	)
	if err != nil {
		return fmt.Errorf("update branch report: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return branch_report.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *BranchReportRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE branch_financial_summaries SET deleted_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete branch report: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return branch_report.ErrNotFound
	}
	return nil
}

// SetApprovedBy mengisi approved_by pada BranchFinancialSummary.
func (r *BranchReportRepository) SetApprovedBy(ctx context.Context, s scope.Scope, id uuid.UUID, approvedBy uuid.UUID) error {
	const q = `
		UPDATE branch_financial_summaries SET approved_by = $4
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL AND approved_by IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id, approvedBy)
	if err != nil {
		return fmt.Errorf("approve branch report: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return branch_report.ErrNotFound
	}
	return nil
}

// ExistsByPeriod mengecek apakah laporan untuk tenant+branch+period sudah ada.
func (r *BranchReportRepository) ExistsByPeriod(ctx context.Context, s scope.Scope, branchID *uuid.UUID, periodType branch_report.PeriodType, periodStart time.Time) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(
		SELECT 1 FROM branch_financial_summaries
		WHERE tenant_id = $1 AND company_id = $2
			AND branch_id IS NOT DISTINCT FROM $3
			AND period_type = $4 AND period_start = $5
			AND deleted_at IS NULL
	)`
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, branchID, periodType, periodStart).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check exists by period: %w", err)
	}
	return exists, nil
}

// ── BranchFinancialSummary ReadRepository ───────────────────────────────────

// GetByID mengambil satu BranchFinancialSummary berdasarkan ID + scope.
func (r *BranchReportRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*branch_report.BranchFinancialSummary, error) {
	const q = `
		SELECT id, branch_id, period_type, period_start, period_end,
			pendapatan_operasional, pendapatan_bunga_margin, pendapatan_lain, total_pendapatan,
			biaya_operasional, biaya_personel, biaya_administrasi, beban_ppap, total_biaya,
			laba_rugi_bersih,
			total_aset, kas_dan_bank, pinjaman_diberikan, simpanan_diterima,
			total_kewajiban, modal_sendiri,
			npl_ratio, bopo_ratio, roa, car,
			calculated_at, approved_by, created_at, deleted_at
		FROM branch_financial_summaries
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e branch_report.BranchFinancialSummary
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.BranchID, &e.PeriodType, &e.PeriodStart, &e.PeriodEnd,
		&e.PendapatanOperasional, &e.PendapatanBungaMargin, &e.PendapatanLain, &e.TotalPendapatan,
		&e.BiayaOperasional, &e.BiayaPersonel, &e.BiayaAdministrasi, &e.BebanPPAP, &e.TotalBiaya,
		&e.LabaRugiBersih,
		&e.TotalAset, &e.KasDanBank, &e.PinjamanDiberikan, &e.SimpananDiterima,
		&e.TotalKewajiban, &e.ModalSendiri,
		&e.NPLRatio, &e.BOPORatio, &e.ROA, &e.CAR,
		&e.CalculatedAt, &e.ApprovedBy, &e.CreatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, branch_report.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get branch report by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar BranchFinancialSummary dengan filter, pagination + sorting.
func (r *BranchReportRepository) List(ctx context.Context, s scope.Scope, filter branch_report.ListFilter, limit, offset int, sortBy, order string) ([]*branch_report.BranchFinancialSummary, int, error) {
	total, err := r.countReports(ctx, s, filter)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectReports(ctx, s, filter, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// GetConsolidated mengambil agregasi SUM semua cabang untuk periode tertentu.
func (r *BranchReportRepository) GetConsolidated(ctx context.Context, s scope.Scope, periodType branch_report.PeriodType, periodStart time.Time) (*branch_report.ConsolidatedRow, error) {
	const q = `
		SELECT
			$1::uuid AS tenant_id,
			$2::uuid AS company_id,
			$3::text AS period_type,
			$4::date AS period_start,
			MAX(period_end) AS period_end,
			COALESCE(SUM(pendapatan_operasional), 0) AS pendapatan_operasional,
			COALESCE(SUM(pendapatan_bunga_margin), 0) AS pendapatan_bunga_margin,
			COALESCE(SUM(pendapatan_lain), 0) AS pendapatan_lain,
			COALESCE(SUM(total_pendapatan), 0) AS total_pendapatan,
			COALESCE(SUM(biaya_operasional), 0) AS biaya_operasional,
			COALESCE(SUM(biaya_personel), 0) AS biaya_personel,
			COALESCE(SUM(biaya_administrasi), 0) AS biaya_administrasi,
			COALESCE(SUM(beban_ppap), 0) AS beban_ppap,
			COALESCE(SUM(total_biaya), 0) AS total_biaya,
			COALESCE(SUM(laba_rugi_bersih), 0) AS laba_rugi_bersih,
			COALESCE(SUM(total_aset), 0) AS total_aset,
			COALESCE(SUM(kas_dan_bank), 0) AS kas_dan_bank,
			COALESCE(SUM(pinjaman_diberikan), 0) AS pinjaman_diberikan,
			COALESCE(SUM(simpanan_diterima), 0) AS simpanan_diterima,
			COALESCE(SUM(total_kewajiban), 0) AS total_kewajiban,
			COALESCE(SUM(modal_sendiri), 0) AS modal_sendiri,
			AVG(npl_ratio) AS npl_ratio,
			AVG(bopo_ratio) AS bopo_ratio,
			AVG(roa) AS roa,
			AVG(car) AS car,
			COUNT(*) AS branch_count
		FROM branch_financial_summaries
		WHERE tenant_id = $1 AND company_id = $2
			AND branch_id IS NOT NULL
			AND period_type = $3 AND period_start = $4
			AND deleted_at IS NULL`

	var cr branch_report.ConsolidatedRow
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, periodType, periodStart).Scan(
		&cr.TenantID, &cr.CompanyID,
		&cr.PeriodType, &cr.PeriodStart, &cr.PeriodEnd,
		&cr.PendapatanOperasional, &cr.PendapatanBungaMargin, &cr.PendapatanLain, &cr.TotalPendapatan,
		&cr.BiayaOperasional, &cr.BiayaPersonel, &cr.BiayaAdministrasi, &cr.BebanPPAP, &cr.TotalBiaya,
		&cr.LabaRugiBersih,
		&cr.TotalAset, &cr.KasDanBank, &cr.PinjamanDiberikan, &cr.SimpananDiterima,
		&cr.TotalKewajiban, &cr.ModalSendiri,
		&cr.NPLRatio, &cr.BOPORatio, &cr.ROA, &cr.CAR,
		&cr.BranchCount,
	)
	if err == sql.ErrNoRows || cr.BranchCount == 0 {
		return nil, branch_report.ErrNoBranchesForConsolidation
	}
	if err != nil {
		return nil, fmt.Errorf("get consolidated report: %w", err)
	}

	return &cr, nil
}

// countReports menghitung total laporan dengan filter.
func (r *BranchReportRepository) countReports(ctx context.Context, s scope.Scope, filter branch_report.ListFilter) (int, error) {
	base := `SELECT COUNT(*) FROM branch_financial_summaries WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	where, args := r.buildFilterClauses(filter, 3)
	q := base + where

	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count branch reports: %w", err)
	}
	return total, nil
}

// selectReports mengambil baris dengan filter + ORDER BY + LIMIT/OFFSET.
func (r *BranchReportRepository) selectReports(ctx context.Context, s scope.Scope, filter branch_report.ListFilter, limit, offset int, sortBy, order string) ([]*branch_report.BranchFinancialSummary, error) {
	col := safeColumn(sortBy, map[string]string{
		"period_start": "period_start",
		"created_at":   "created_at",
		"period_type":  "period_type",
	}, "created_at")
	dir := safeOrder(order)

	base := `SELECT id, branch_id, period_type, period_start, period_end,
		pendapatan_operasional, pendapatan_bunga_margin, pendapatan_lain, total_pendapatan,
		biaya_operasional, biaya_personel, biaya_administrasi, beban_ppap, total_biaya,
		laba_rugi_bersih,
		total_aset, kas_dan_bank, pinjaman_diberikan, simpanan_diterima,
		total_kewajiban, modal_sendiri,
		npl_ratio, bopo_ratio, roa, car,
		calculated_at, approved_by, created_at, deleted_at
		FROM branch_financial_summaries
		WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`

	where, args := r.buildFilterClauses(filter, 3)
	q := fmt.Sprintf("%s%s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		base, where, col, dir, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list branch reports: %w", err)
	}
	defer rows.Close()
	return scanBranchReports(rows)
}

// buildFilterClauses membangun WHERE clause dinamis berdasarkan filter.
func (r *BranchReportRepository) buildFilterClauses(filter branch_report.ListFilter, startIdx int) (string, []interface{}) {
	var clauses []string
	var args []interface{}
	idx := startIdx

	if filter.BranchID != nil {
		clauses = append(clauses, fmt.Sprintf(" AND branch_id IS NOT DISTINCT FROM $%d", idx))
		args = append(args, *filter.BranchID)
		idx++
	}
	if filter.PeriodType != nil {
		clauses = append(clauses, fmt.Sprintf(" AND period_type = $%d", idx))
		args = append(args, *filter.PeriodType)
		idx++
	}
	if filter.PeriodStart != nil {
		clauses = append(clauses, fmt.Sprintf(" AND period_start >= $%d", idx))
		args = append(args, *filter.PeriodStart)
		idx++
	}
	if filter.PeriodEnd != nil {
		clauses = append(clauses, fmt.Sprintf(" AND period_end <= $%d", idx))
		args = append(args, *filter.PeriodEnd)
		idx++
	}
	if filter.ApprovedBy != nil {
		clauses = append(clauses, fmt.Sprintf(" AND approved_by IS NOT DISTINCT FROM $%d", idx))
		args = append(args, *filter.ApprovedBy)
		idx++
	}

	return strings.Join(clauses, ""), args
}

func scanBranchReports(rows *sql.Rows) ([]*branch_report.BranchFinancialSummary, error) {
	var results []*branch_report.BranchFinancialSummary
	for rows.Next() {
		var e branch_report.BranchFinancialSummary
		if err := rows.Scan(
			&e.ID, &e.BranchID, &e.PeriodType, &e.PeriodStart, &e.PeriodEnd,
			&e.PendapatanOperasional, &e.PendapatanBungaMargin, &e.PendapatanLain, &e.TotalPendapatan,
			&e.BiayaOperasional, &e.BiayaPersonel, &e.BiayaAdministrasi, &e.BebanPPAP, &e.TotalBiaya,
			&e.LabaRugiBersih,
			&e.TotalAset, &e.KasDanBank, &e.PinjamanDiberikan, &e.SimpananDiterima,
			&e.TotalKewajiban, &e.ModalSendiri,
			&e.NPLRatio, &e.BOPORatio, &e.ROA, &e.CAR,
			&e.CalculatedAt, &e.ApprovedBy, &e.CreatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// ── EliminationRule WriteRepository ─────────────────────────────────────────

// SaveRule menyimpan entity EliminationRule baru ke database.
func (r *BranchReportRepository) SaveRule(ctx context.Context, s scope.Scope, e *branch_report.EliminationRule) error {
	const q = `
		INSERT INTO elimination_rules (
			id, tenant_id, company_id, rule_name, rule_type,
			from_branch_id, to_branch_id, is_active, description, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID, e.RuleName, e.RuleType,
		e.FromBranchID, e.ToBranchID, e.IsActive, e.Description, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save elimination rule: %w", err)
	}
	return nil
}

// ToggleActive membalik status is_active pada EliminationRule.
// Mengembalikan status baru setelah toggle.
func (r *BranchReportRepository) ToggleActive(ctx context.Context, s scope.Scope, id uuid.UUID) (bool, error) {
	const q = `
		UPDATE elimination_rules SET is_active = NOT is_active
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL
		RETURNING is_active`

	var newActive bool
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(&newActive)
	if err == sql.ErrNoRows {
		return false, branch_report.ErrRuleNotFound
	}
	if err != nil {
		return false, fmt.Errorf("toggle elimination rule: %w", err)
	}
	return newActive, nil
}

// ── EliminationRule ReadRepository ──────────────────────────────────────────

// ListRules mengambil daftar EliminationRule dengan pagination + sorting.
func (r *BranchReportRepository) ListRules(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*branch_report.EliminationRule, int, error) {
	total, err := r.countRules(ctx, s)
	if err != nil {
		return nil, 0, err
	}

	col := safeColumn(sortBy, map[string]string{
		"rule_name":  "rule_name",
		"rule_type":  "rule_type",
		"created_at": "created_at",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, rule_name, rule_type, from_branch_id, to_branch_id, is_active, description, created_at, deleted_at
		 FROM elimination_rules
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
		 ORDER BY %s %s LIMIT $3 OFFSET $4`, col, dir,
	)

	rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list elimination rules: %w", err)
	}
	defer rows.Close()

	var results []*branch_report.EliminationRule
	for rows.Next() {
		var e branch_report.EliminationRule
		if err := rows.Scan(
			&e.ID, &e.RuleName, &e.RuleType,
			&e.FromBranchID, &e.ToBranchID, &e.IsActive,
			&e.Description, &e.CreatedAt, &e.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		results = append(results, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// countRules menghitung total EliminationRule aktif milik scope ini.
func (r *BranchReportRepository) countRules(ctx context.Context, s scope.Scope) (int, error) {
	const q = `SELECT COUNT(*) FROM elimination_rules WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count elimination rules: %w", err)
	}
	return total, nil
}
