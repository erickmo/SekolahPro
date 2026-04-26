// Package payroll_deduction adalah domain Vernon untuk potongan payroll per employee.
//
// Terdiri dari 4 tabel: payroll_authorization, payroll_batch, payroll_deduction, payroll_mapping.
// payroll_authorization autoloads: nasabah, rekening.
// payroll_batch: NO autoload.
// payroll_deduction autoloads: batch, nasabah, rekening, authorization.
// payroll_mapping: NO autoload.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package payroll_deduction

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Authorization field constants ────────────────────────────────────────────

const (
	AuthFieldNasabahID       = "nasabah_id"
	AuthFieldRekeningID      = "rekening_id"
	AuthFieldAuthorizationType = "authorization_type"
	AuthFieldDeductionType   = "deduction_type"
	AuthFieldEffectiveDate   = "effective_date"
	AuthFieldStatus          = "status"
)

// ── Batch field constants ────────────────────────────────────────────────────

const (
	BatchFieldBatchNumber = "batch_number"
	BatchFieldPeriodMonth = "period_month"
	BatchFieldPeriodYear  = "period_year"
	BatchFieldUploadSource = "upload_source"
	BatchFieldStatus      = "status"
)

// ── Deduction field constants ────────────────────────────────────────────────

const (
	DeductFieldBatchID         = "batch_id"
	DeductFieldNasabahID       = "nasabah_id"
	DeductFieldRekeningID      = "rekening_id"
	DeductFieldDeductionType   = "deduction_type"
	DeductFieldMatchStatus     = "match_status"
	DeductFieldDeductionStatus = "deduction_status"
)

// ── Mapping field constants ──────────────────────────────────────────────────

const (
	MapFieldEmployeeID     = "employee_id"
	MapFieldEmployeeNumber = "employee_number"
	MapFieldEmployeeName   = "employee_name"
	MapFieldNasabahID      = "nasabah_id"
	MapFieldRekeningID     = "rekening_id"
)

// ── Enum constants ───────────────────────────────────────────────────────────

const (
	AuthTypePayrollDeduction = "payroll_deduction"
	AuthTypeAutoDebit        = "auto_debit"
	AuthTypeStandingOrder    = "standing_order"
)

var validAuthTypes = map[string]bool{
	AuthTypePayrollDeduction: true, AuthTypeAutoDebit: true, AuthTypeStandingOrder: true,
}

const (
	DeductTypeSimpanan = "simpanan"
	DeductTypePinjaman = "pinjaman"
	DeductTypeSPP      = "spp"
	DeductTypeKantin   = "kantin"
	DeductTypeOther    = "other"
)

var validDeductionTypes = map[string]bool{
	DeductTypeSimpanan: true, DeductTypePinjaman: true,
	DeductTypeSPP: true, DeductTypeKantin: true, DeductTypeOther: true,
}

const (
	AuthStatusActive  = "active"
	AuthStatusRevoked = "revoked"
	AuthStatusExpired = "expired"
)

var validAuthStatuses = map[string]bool{
	AuthStatusActive: true, AuthStatusRevoked: true, AuthStatusExpired: true,
}

const (
	BatchStatusUploaded        = "uploaded"
	BatchStatusCalculated      = "calculated"
	BatchStatusPendingApproval = "pending_approval"
	BatchStatusApproved        = "approved"
	BatchStatusExecuting       = "executing"
	BatchStatusExecuted        = "executed"
	BatchStatusCompleted       = "completed"
	BatchStatusFailed          = "failed"
)

var validBatchStatuses = map[string]bool{
	BatchStatusUploaded: true, BatchStatusCalculated: true,
	BatchStatusPendingApproval: true, BatchStatusApproved: true,
	BatchStatusExecuting: true, BatchStatusExecuted: true,
	BatchStatusCompleted: true, BatchStatusFailed: true,
}

const (
	MatchStatusMatched   = "matched"
	MatchStatusUnmatched = "unmatched"
	MatchStatusSkipped   = "skipped"
)

var validMatchStatuses = map[string]bool{
	MatchStatusMatched: true, MatchStatusUnmatched: true, MatchStatusSkipped: true,
}

const (
	DeductStatusPending = "pending"
	DeductStatusExecuted = "executed"
	DeductStatusFailed  = "failed"
	DeductStatusSkipped = "skipped"
)

var validDeductStatuses = map[string]bool{
	DeductStatusPending: true, DeductStatusExecuted: true,
	DeductStatusFailed: true, DeductStatusSkipped: true,
}

const (
	UploadSourceCSV = "csv"
	UploadSourceXLSX = "xlsx"
	UploadSourceAPI = "api"
)

var validUploadSources = map[string]bool{
	UploadSourceCSV: true, UploadSourceXLSX: true, UploadSourceAPI: true,
}

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelNasabah       = "nasabah"
	RelRekening      = "rekening"
	RelBatch         = "batch"
	RelAuthorization = "authorization"
)

// ── AuthorizationDescriptor ──────────────────────────────────────────────────

// AuthorizationDescriptor mengimplementasi vernon.DomainDescriptor untuk payroll_authorization.
type AuthorizationDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *AuthorizationDescriptor) TableName() string { return "payroll_authorization" }

// DefaultRels mendefinisikan relasi payroll_authorization: nasabah + rekening.
func (d *AuthorizationDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         AuthFieldNasabahID,
			LocalKey:   AuthFieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap"},
		},
		RelRekening: {
			Domain:     "rekening",
			Type:       vernon.RelBelongsTo,
			FK:         AuthFieldRekeningID,
			LocalKey:   AuthFieldRekeningID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"account_number"},
		},
	}
}

// Validate memvalidasi invariant payroll_authorization.
func (d *AuthorizationDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, AuthFieldNasabahID, "nasabah"); err != nil {
		return err
	}
	if err := requireString(data, AuthFieldRekeningID, "rekening"); err != nil {
		return err
	}
	if err := requireEnum(data, AuthFieldAuthorizationType, "tipe otorisasi", validAuthTypes); err != nil {
		return err
	}
	if err := requireEnum(data, AuthFieldDeductionType, "tipe potongan", validDeductionTypes); err != nil {
		return err
	}
	if err := requireString(data, AuthFieldEffectiveDate, "tanggal efektif"); err != nil {
		return err
	}
	return requireEnum(data, AuthFieldStatus, "status otorisasi", validAuthStatuses)
}

// ── BatchDescriptor ──────────────────────────────────────────────────────────

// BatchDescriptor mengimplementasi vernon.DomainDescriptor untuk payroll_batch.
type BatchDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *BatchDescriptor) TableName() string { return "payroll_batch" }

// DefaultRels — NO autoload per ADR.
func (d *BatchDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant payroll_batch.
func (d *BatchDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, BatchFieldBatchNumber, "nomor batch"); err != nil {
		return err
	}
	return requireEnum(data, BatchFieldStatus, "status batch", validBatchStatuses)
}

// ── DeductionDescriptor ──────────────────────────────────────────────────────

// DeductionDescriptor mengimplementasi vernon.DomainDescriptor untuk payroll_deduction.
type DeductionDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *DeductionDescriptor) TableName() string { return "payroll_deduction" }

// DefaultRels mendefinisikan relasi payroll_deduction: batch + nasabah + rekening + authorization.
func (d *DeductionDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelBatch: {
			Domain:     "payroll_batch",
			Type:       vernon.RelBelongsTo,
			FK:         DeductFieldBatchID,
			LocalKey:   DeductFieldBatchID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"batch_number", "status"},
		},
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         DeductFieldNasabahID,
			LocalKey:   DeductFieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap"},
		},
		RelRekening: {
			Domain:     "rekening",
			Type:       vernon.RelBelongsTo,
			FK:         DeductFieldRekeningID,
			LocalKey:   DeductFieldRekeningID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"account_number"},
		},
		RelAuthorization: {
			Domain:     "payroll_authorization",
			Type:       vernon.RelBelongsTo,
			FK:         "authorization_id",
			LocalKey:   "authorization_id",
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"authorization_type", "deduction_type", "status"},
		},
	}
}

// Validate memvalidasi invariant payroll_deduction.
func (d *DeductionDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, DeductFieldBatchID, "batch"); err != nil {
		return err
	}
	if err := requireString(data, DeductFieldNasabahID, "nasabah"); err != nil {
		return err
	}
	if err := requireString(data, DeductFieldRekeningID, "rekening"); err != nil {
		return err
	}
	if err := requireEnum(data, DeductFieldDeductionType, "tipe potongan", validDeductionTypes); err != nil {
		return err
	}
	return nil
}

// ── MappingDescriptor ────────────────────────────────────────────────────────

// MappingDescriptor mengimplementasi vernon.DomainDescriptor untuk payroll_mapping.
type MappingDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *MappingDescriptor) TableName() string { return "payroll_mapping" }

// DefaultRels — NO autoload per ADR.
func (d *MappingDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant payroll_mapping.
func (d *MappingDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, MapFieldEmployeeID, "employee ID"); err != nil {
		return err
	}
	if err := requireString(data, MapFieldEmployeeNumber, "nomor employee"); err != nil {
		return err
	}
	if err := requireString(data, MapFieldEmployeeName, "nama employee"); err != nil {
		return err
	}
	if err := requireString(data, MapFieldNasabahID, "nasabah"); err != nil {
		return err
	}
	return requireString(data, MapFieldRekeningID, "rekening")
}

// ── shared validation helpers ─────────────────────────────────────────────────

func requireString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}

func requireEnum(data map[string]any, field, label string, valid map[string]bool) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	if !valid[val] {
		return fmt.Errorf("%s tidak valid: %q", label, val)
	}
	return nil
}
