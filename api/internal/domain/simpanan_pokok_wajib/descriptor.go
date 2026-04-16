// Package simpanan_pokok_wajib adalah domain Vernon untuk simpanan pokok & wajib.
//
// Simpanan pokok adalah setoran satu kali saat mendaftar sebagai anggota.
// Simpanan wajib adalah setoran rutin bulanan yang wajib dibayar anggota.
// Keduanya adalah kewajiban keanggotaan koperasi dan dapat dikembalikan saat keluar.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package simpanan_pokok_wajib

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldNasabahID    = "nasabah_id"
	FieldRekeningID   = "rekening_id"
	FieldType         = "type"
	FieldAmount       = "amount"
	FieldPeriod       = "period"
	FieldDueDate      = "due_date"
	FieldPaidDate     = "paid_date"
	FieldStatus       = "status"
	FieldPaymentRef   = "payment_ref"
	FieldNotes        = "notes"
)

// Type constants.
const (
	TypePokok = "pokok"
	TypeWajib = "wajib"
)

// Status constants.
const (
	StatusPending  = "pending"
	StatusPaid     = "paid"
	StatusOverdue  = "overdue"
	StatusRefunded = "refunded"
)

// Relation name constants.
const (
	RelNasabah  = "nasabah"
	RelRekening = "rekening"
)

var validTypes = map[string]bool{
	TypePokok: true, TypeWajib: true,
}

var validStatuses = map[string]bool{
	StatusPending: true, StatusPaid: true,
	StatusOverdue: true, StatusRefunded: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk simpanan_pokok_wajib.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "simpanan_pokok_wajib" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo nasabah (autoload) dan rekening (autoload).
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
			Fields:     []string{"no_rekening", "status"},
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

	typeVal, _ := data[FieldType].(string)
	if typeVal == "" {
		return errors.New("type wajib diisi")
	}

	amount, ok := data[FieldAmount].(float64)
	if !ok {
		return errors.New("amount wajib diisi")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih dari 0, got %.2f", amount)
	}
	return nil
}

// validateEnums memeriksa enum fields.
func validateEnums(data map[string]any) error {
	typeVal, _ := data[FieldType].(string)
	if typeVal != "" && !validTypes[typeVal] {
		return fmt.Errorf("type tidak valid: %q (harus pokok/wajib)", typeVal)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus pending/paid/overdue/refunded)", status)
	}

	if typeVal == TypeWajib {
		period, _ := data[FieldPeriod].(string)
		if period == "" {
			return errors.New("period wajib diisi untuk simpanan wajib")
		}
	}
	return nil
}
