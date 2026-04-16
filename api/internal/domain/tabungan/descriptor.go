// Package tabungan adalah domain Vernon untuk tabungan/simpanan sukarela.
//
// Tabungan adalah produk simpanan paling aktif dengan fleksibilitas setor/tarik.
// Mendukung multiple product types: regular, education, holiday, qurban, goal-based.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package tabungan

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldNasabahID       = "nasabah_id"
	FieldRekeningID      = "rekening_id"
	FieldProductType     = "product_type"
	FieldBalance         = "balance"
	FieldHoldBalance     = "hold_balance"
	FieldGoalAmount      = "goal_amount"
	FieldGoalDate        = "goal_date"
	FieldMonthlyAutoDebit = "monthly_auto_debit"
	FieldStatus          = "status"
	FieldLastTransaction = "last_transaction_at"
	FieldDormantSince    = "dormant_since"
)

// Product type constants.
const (
	ProductRegular   = "regular"
	ProductEducation = "education"
	ProductHoliday   = "holiday"
	ProductQurban    = "qurban"
	ProductGoal      = "goal"
)

// Status constants.
const (
	StatusActive  = "active"
	StatusDormant = "dormant"
	StatusFrozen  = "frozen"
	StatusClosed  = "closed"
)

// Relation name constants.
const (
	RelNasabah  = "nasabah"
	RelRekening = "rekening"
)

var validProductTypes = map[string]bool{
	ProductRegular: true, ProductEducation: true,
	ProductHoliday: true, ProductQurban: true,
	ProductGoal: true,
}

var validStatuses = map[string]bool{
	StatusActive: true, StatusDormant: true,
	StatusFrozen: true, StatusClosed: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk tabungan.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "tabungan" }

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

	productType, _ := data[FieldProductType].(string)
	if productType == "" {
		return errors.New("product_type wajib diisi")
	}
	return nil
}

// validateEnums memeriksa enum fields dan balance.
func validateEnums(data map[string]any) error {
	productType, _ := data[FieldProductType].(string)
	if productType != "" && !validProductTypes[productType] {
		return fmt.Errorf("product_type tidak valid: %q", productType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus active/dormant/frozen/closed)", status)
	}

	if err := validateBalances(data); err != nil {
		return err
	}

	return validateGoal(data)
}

// validateBalances memeriksa balance >= 0 dan hold_balance <= balance.
func validateBalances(data map[string]any) error {
	balance, balanceOK := data[FieldBalance].(float64)
	if balanceOK && balance < 0 {
		return errors.New("balance tidak boleh negatif")
	}

	hold, holdOK := data[FieldHoldBalance].(float64)
	if holdOK && balanceOK && hold > balance {
		return errors.New("hold_balance tidak boleh lebih besar dari balance")
	}
	return nil
}

// validateGoal memeriksa goal_amount hanya boleh untuk goal product.
func validateGoal(data map[string]any) error {
	productType, _ := data[FieldProductType].(string)
	goalAmount, hasGoal := data[FieldGoalAmount].(float64)

	if hasGoal && goalAmount > 0 && productType != ProductGoal {
		return errors.New("goal_amount hanya untuk product_type goal")
	}
	return nil
}
