// Package pinjaman adalah domain Vernon untuk pinjaman/kredit.
//
// Pinjaman adalah entitas transaksional dengan lifecycle: pending → approved → active → paid_off.
// Mendukung berbagai status lifecycle termasuk rejected, written_off, dan overdue.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package pinjaman

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldNasabahID       = "nasabah_id"
	FieldRekeningID      = "rekening_id"
	FieldProdukAkadID    = "produk_akad_id"
	FieldPrincipalAmount = "principal_amount"
	FieldTenorMonths     = "tenor_months"
	FieldInterestRate    = "interest_rate"
	FieldStatus          = "status"
	FieldDisbursementMethod = "disbursement_method"
)

// Status constants.
const (
	StatusPending     = "pending"
	StatusApproved    = "approved"
	StatusActive      = "active"
	StatusPaidOff     = "paid_off"
	StatusRejected    = "rejected"
	StatusWrittenOff  = "written_off"
	StatusOverdue     = "overdue"
)

// Disbursement method constants.
const (
	DisbursementTransfer  = "transfer"
	DisbursementCash      = "cash"
	DisbursementDeduction = "deduction"
)

// Relation name constants.
const (
	RelNasabah    = "nasabah"
	RelRekening   = "rekening"
	RelProdukAkad = "produk_akad"
)

var validStatuses = map[string]bool{
	StatusPending: true, StatusApproved: true, StatusActive: true,
	StatusPaidOff: true, StatusRejected: true, StatusWrittenOff: true,
	StatusOverdue: true,
}

var validDisbursementMethods = map[string]bool{
	DisbursementTransfer: true, DisbursementCash: true, DisbursementDeduction: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk pinjaman.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "pinjaman" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo nasabah (autoload), rekening (autoload), dan produk_akad (autoload).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         FieldNasabahID,
			LocalKey:   FieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap", "no_nasabah", "type"},
		},
		RelRekening: {
			Domain:     "rekening",
			Type:       vernon.RelBelongsTo,
			FK:         FieldRekeningID,
			LocalKey:   FieldRekeningID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"no_rekening", "category"},
		},
		RelProdukAkad: {
			Domain:     "produk_akad",
			Type:       vernon.RelBelongsTo,
			FK:         FieldProdukAkadID,
			LocalKey:   FieldProdukAkadID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "akad_type"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateRequired(data); err != nil {
		return err
	}
	return validateEnums(data)
}

// validateRequired memeriksa field wajib.
func validateRequired(data map[string]any) error {
	nasabahID, _ := data[FieldNasabahID].(string)
	if nasabahID == "" {
		return errors.New("nasabah_id wajib diisi")
	}

	rekeningID, _ := data[FieldRekeningID].(string)
	if rekeningID == "" {
		return errors.New("rekening_id wajib diisi")
	}

	produkAkadID, _ := data[FieldProdukAkadID].(string)
	if produkAkadID == "" {
		return errors.New("produk_akad_id wajib diisi")
	}

	principalAmount, principalOK := data[FieldPrincipalAmount].(float64)
	if !principalOK {
		return errors.New("principal_amount wajib diisi")
	}
	if principalAmount <= 0 {
		return errors.New("principal_amount harus lebih dari 0")
	}

	tenorMonths, tenorOK := data[FieldTenorMonths].(float64)
	if !tenorOK {
		return errors.New("tenor_months wajib diisi")
	}
	if tenorMonths < 1 {
		return errors.New("tenor_months minimal 1")
	}

	return nil
}

// validateEnums memeriksa enum fields dan interest_rate.
func validateEnums(data map[string]any) error {
	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf(
			"status tidak valid: %q (harus pending/approved/active/paid_off/rejected/written_off/overdue)",
			status,
		)
	}

	disbursementMethod, _ := data[FieldDisbursementMethod].(string)
	if disbursementMethod != "" && !validDisbursementMethods[disbursementMethod] {
		return fmt.Errorf(
			"disbursement_method tidak valid: %q (harus transfer/cash/deduction)",
			disbursementMethod,
		)
	}

	interestRate, interestOK := data[FieldInterestRate].(float64)
	if interestOK && interestRate < 0 {
		return errors.New("interest_rate tidak boleh negatif")
	}

	return nil
}
