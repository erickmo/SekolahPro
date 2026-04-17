// Package angsuran adalah domain Vernon untuk angsuran/cicilan pinjaman.
//
// Angsuran merepresentasikan setiap cicilan dari sebuah pinjaman (K007),
// termasuk komponen pokok, bunga, dan total yang harus dibayar.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package angsuran

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldPinjamanID       = "pinjaman_id"
	FieldInstallmentNumber = "installment_number"
	FieldDueDate          = "due_date"
	FieldPrincipalAmount  = "principal_amount"
	FieldInterestAmount   = "interest_amount"
	FieldTotalAmount      = "total_amount"
	FieldStatus           = "status"
)

// Status constants.
const (
	StatusScheduled = "scheduled"
	StatusPaid      = "paid"
	StatusPartial   = "partial"
	StatusOverdue   = "overdue"
	StatusWaived    = "waived"
)

// Relation name constants.
const (
	RelPinjaman = "pinjaman"
)

var validStatuses = map[string]bool{
	StatusScheduled: true, StatusPaid: true, StatusPartial: true,
	StatusOverdue: true, StatusWaived: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk angsuran.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "angsuran" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo pinjaman (autoload).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelPinjaman: {
			Domain:     "pinjaman",
			Type:       vernon.RelBelongsTo,
			FK:         FieldPinjamanID,
			LocalKey:   FieldPinjamanID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"loan_number", "akad_type", "status"},
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
	pinjamanID, _ := data[FieldPinjamanID].(string)
	if pinjamanID == "" {
		return errors.New("pinjaman_id wajib diisi")
	}

	installmentNumber, installmentOK := data[FieldInstallmentNumber].(float64)
	if !installmentOK {
		return errors.New("installment_number wajib diisi")
	}
	if installmentNumber < 1 {
		return errors.New("installment_number minimal 1")
	}

	return nil
}

// validateEnums memeriksa enum fields dan amount non-negatif.
func validateEnums(data map[string]any) error {
	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf(
			"status tidak valid: %q (harus scheduled/paid/partial/overdue/waived)",
			status,
		)
	}

	if err := validateAmounts(data); err != nil {
		return err
	}

	return nil
}

// validateAmounts memeriksa amount fields >= 0.
func validateAmounts(data map[string]any) error {
	principal, principalOK := data[FieldPrincipalAmount].(float64)
	if principalOK && principal < 0 {
		return errors.New("principal_amount tidak boleh negatif")
	}

	interest, interestOK := data[FieldInterestAmount].(float64)
	if interestOK && interest < 0 {
		return errors.New("interest_amount tidak boleh negatif")
	}

	total, totalOK := data[FieldTotalAmount].(float64)
	if totalOK && total < 0 {
		return errors.New("total_amount tidak boleh negatif")
	}

	return nil
}
