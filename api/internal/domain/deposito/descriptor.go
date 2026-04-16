// Package deposito adalah domain Vernon untuk deposito/simpanan berjangka.
//
// Deposito adalah simpanan berjangka dengan tenor tetap dan bunga lebih tinggi.
// Mendukung auto-rollover, pencairan awal dengan penalti, dan agunan pinjaman.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package deposito

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldNasabahID       = "nasabah_id"
	FieldRekeningID      = "rekening_id"
	FieldPrincipal       = "principal"
	FieldTenorMonths     = "tenor_months"
	FieldRate            = "rate"
	FieldMaturityDate    = "maturity_date"
	FieldStatus          = "status"
	FieldAutoRollover    = "auto_rollover"
	FieldPaymentMethod   = "payment_method"
	FieldOnHold          = "on_hold"
	FieldHoldReference   = "hold_reference"
	FieldStartDate       = "start_date"
	FieldRolloverCount   = "rollover_count"
)

// Status constants.
const (
	StatusActive       = "active"
	StatusMatured      = "matured"
	StatusRolledOver   = "rolled_over"
	StatusEarlyWithdraw = "early_withdrawn"
	StatusClosed       = "closed"
)

// Payment method constants.
const (
	PaymentMonthly = "monthly"
	PaymentMaturity = "at_maturity"
	PaymentCompound = "compounded"
)

// Relation name constants.
const (
	RelNasabah  = "nasabah"
	RelRekening = "rekening"
)

var validStatuses = map[string]bool{
	StatusActive: true, StatusMatured: true,
	StatusRolledOver: true, StatusEarlyWithdraw: true,
	StatusClosed: true,
}

var validPaymentMethods = map[string]bool{
	PaymentMonthly: true, PaymentMaturity: true, PaymentCompound: true,
}

var validTenors = map[int]bool{
	1: true, 3: true, 6: true, 12: true, 24: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk deposito.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "deposito" }

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

	principal, ok := data[FieldPrincipal].(float64)
	if !ok {
		return errors.New("principal wajib diisi")
	}
	if principal <= 0 {
		return fmt.Errorf("principal harus lebih dari 0, got %.2f", principal)
	}

	tenor, ok := data[FieldTenorMonths].(float64)
	if !ok {
		return errors.New("tenor_months wajib diisi")
	}
	if !validTenors[int(tenor)] {
		return fmt.Errorf("tenor_months tidak valid: %v (harus 1/3/6/12/24)", int(tenor))
	}

	rate, ok := data[FieldRate].(float64)
	if !ok {
		return errors.New("rate wajib diisi")
	}
	if rate < 0 {
		return errors.New("rate tidak boleh negatif")
	}
	return nil
}

// validateEnums memeriksa enum fields.
func validateEnums(data map[string]any) error {
	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	method, _ := data[FieldPaymentMethod].(string)
	if method != "" && !validPaymentMethods[method] {
		return fmt.Errorf("payment_method tidak valid: %q", method)
	}
	return nil
}
