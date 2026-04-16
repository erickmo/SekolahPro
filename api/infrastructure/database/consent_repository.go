package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/consent"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// ConsentRepository adalah concrete implementation dari consent.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type ConsentRepository struct {
	db *sqlx.DB
}

// NewConsentRepository membuat instance baru ConsentRepository.
func NewConsentRepository(db *sqlx.DB) *ConsentRepository {
	return &ConsentRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity ConsentRecord baru ke database.
func (r *ConsentRepository) Save(ctx context.Context, s scope.Scope, e *consent.ConsentRecord) error {
	const q = `
		INSERT INTO consent_records (
			id, tenant_id, company_id, nasabah_id,
			consent_type, consent_purpose, legal_basis,
			consent_text, consent_version, consent_given, consent_method,
			withdrawn, withdrawn_at, withdrawal_reason,
			parent_id, parent_relationship, parent_consent_given,
			given_at, ip_address, user_agent, witness_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID, e.NasabahID,
		e.ConsentType, e.ConsentPurpose, e.LegalBasis,
		e.ConsentText, e.ConsentVersion, e.ConsentGiven, e.ConsentMethod,
		e.Withdrawn, e.WithdrawnAt, e.WithdrawalReason,
		e.ParentID, e.ParentRelationship, e.ParentConsentGiven,
		e.GivenAt, e.IPAddress, e.UserAgent, e.WitnessID, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save consent record: %w", err)
	}
	return nil
}

// Update mengupdate ConsentRecord yang sudah ada berdasarkan ID + scope.
func (r *ConsentRepository) Update(ctx context.Context, s scope.Scope, e *consent.ConsentRecord) error {
	const q = `
		UPDATE consent_records
		SET nasabah_id = $4, consent_type = $5, consent_purpose = $6, legal_basis = $7,
		    consent_text = $8, consent_version = $9, consent_given = $10, consent_method = $11,
		    withdrawn = $12, withdrawn_at = $13, withdrawal_reason = $14,
		    parent_id = $15, parent_relationship = $16, parent_consent_given = $17,
		    given_at = $18, ip_address = $19, user_agent = $20, witness_id = $21
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.NasabahID, e.ConsentType, e.ConsentPurpose, e.LegalBasis,
		e.ConsentText, e.ConsentVersion, e.ConsentGiven, e.ConsentMethod,
		e.Withdrawn, e.WithdrawnAt, e.WithdrawalReason,
		e.ParentID, e.ParentRelationship, e.ParentConsentGiven,
		e.GivenAt, e.IPAddress, e.UserAgent, e.WitnessID,
	)
	if err != nil {
		return fmt.Errorf("update consent record: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return consent.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *ConsentRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE consent_records SET deleted_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete consent record: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return consent.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu ConsentRecord berdasarkan ID + scope.
func (r *ConsentRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*consent.ConsentRecord, error) {
	const q = `
		SELECT id, nasabah_id, consent_type, consent_purpose, legal_basis,
		       consent_text, consent_version, consent_given, consent_method,
		       withdrawn, withdrawn_at, withdrawal_reason,
		       parent_id, parent_relationship, parent_consent_given,
		       given_at, ip_address, user_agent, witness_id, created_at, deleted_at
		FROM consent_records
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e consent.ConsentRecord
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.NasabahID, &e.ConsentType, &e.ConsentPurpose, &e.LegalBasis,
		&e.ConsentText, &e.ConsentVersion, &e.ConsentGiven, &e.ConsentMethod,
		&e.Withdrawn, &e.WithdrawnAt, &e.WithdrawalReason,
		&e.ParentID, &e.ParentRelationship, &e.ParentConsentGiven,
		&e.GivenAt, &e.IPAddress, &e.UserAgent, &e.WitnessID, &e.CreatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, consent.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get consent record by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar ConsentRecord dengan pagination, sorting, dan filter.
func (r *ConsentRepository) List(ctx context.Context, s scope.Scope, filter consent.ListFilter, limit, offset int, sortBy, order string) ([]*consent.ConsentRecord, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countConsents(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectConsents(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *ConsentRepository) buildWhereClause(s scope.Scope, f consent.ListFilter) (string, []any) {
	conditions := []string{"tenant_id = $1", "company_id = $2", "deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	if f.ConsentType != nil {
		conditions = append(conditions, fmt.Sprintf("consent_type = $%d", argIdx))
		args = append(args, *f.ConsentType)
		argIdx++
	}
	if f.LegalBasis != nil {
		conditions = append(conditions, fmt.Sprintf("legal_basis = $%d", argIdx))
		args = append(args, *f.LegalBasis)
		argIdx++
	}
	if f.ConsentGiven != nil {
		conditions = append(conditions, fmt.Sprintf("consent_given = $%d", argIdx))
		args = append(args, *f.ConsentGiven)
		argIdx++
	}
	if f.Withdrawn != nil {
		conditions = append(conditions, fmt.Sprintf("withdrawn = $%d", argIdx))
		args = append(args, *f.Withdrawn)
		argIdx++
	}
	if f.NasabahID != nil {
		conditions = append(conditions, fmt.Sprintf("nasabah_id = $%d", argIdx))
		args = append(args, *f.NasabahID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	return where, args
}

// countConsents menghitung total ConsentRecord aktif yang cocok dengan filter.
func (r *ConsentRepository) countConsents(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM consent_records WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count consent records: %w", err)
	}
	return total, nil
}

// selectConsents mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *ConsentRepository) selectConsents(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*consent.ConsentRecord, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":    "created_at",
		"given_at":      "given_at",
		"consent_type":  "consent_type",
		"legal_basis":   "legal_basis",
		"consent_given": "consent_given",
		"withdrawn":     "withdrawn",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, nasabah_id, consent_type, consent_purpose, legal_basis,
		        consent_text, consent_version, consent_given, consent_method,
		        withdrawn, withdrawn_at, withdrawal_reason,
		        parent_id, parent_relationship, parent_consent_given,
		        given_at, ip_address, user_agent, witness_id, created_at, deleted_at
		 FROM consent_records
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list consent records: %w", err)
	}
	defer rows.Close()
	return scanConsents(rows)
}

func scanConsents(rows *sql.Rows) ([]*consent.ConsentRecord, error) {
	var results []*consent.ConsentRecord
	for rows.Next() {
		var e consent.ConsentRecord
		if err := rows.Scan(
			&e.ID, &e.NasabahID, &e.ConsentType, &e.ConsentPurpose, &e.LegalBasis,
			&e.ConsentText, &e.ConsentVersion, &e.ConsentGiven, &e.ConsentMethod,
			&e.Withdrawn, &e.WithdrawnAt, &e.WithdrawalReason,
			&e.ParentID, &e.ParentRelationship, &e.ParentConsentGiven,
			&e.GivenAt, &e.IPAddress, &e.UserAgent, &e.WitnessID, &e.CreatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// Compile-time interface check.
var (
	_ consent.WriteRepository = (*ConsentRepository)(nil)
	_ consent.ReadRepository  = (*ConsentRepository)(nil)
)

// suppressUnusedImport mencegah error unused import untuk time package.
var _ = time.Time{}
