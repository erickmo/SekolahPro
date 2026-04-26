// Package spending_control adalah domain Vernon untuk kontrol belanja uang saku digital.
//
// Autoloads: nasabah, rekening, controlled_by (nasabah).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package spending_control

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field constants ──────────────────────────────────────────────────────────

const (
	FieldNasabahID            = "nasabah_id"
	FieldControlledByID       = "controlled_by_id"
	FieldRekeningID           = "rekening_id"
	FieldDailyLimit           = "daily_limit"
	FieldPerTransactionLimit  = "per_transaction_limit"
	FieldWeeklyLimit          = "weekly_limit"
	FieldMonthlyLimit         = "monthly_limit"
	FieldIsActive             = "is_active"
)

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelNasabah     = "nasabah"
	RelRekening    = "rekening"
	RelControlledBy = "controlled_by"
)

// ── Descriptor ───────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk spending_control.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "spending_control" }

// DefaultRels mendefinisikan relasi spending_control: nasabah + rekening + controlled_by.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         FieldNasabahID,
			LocalKey:   FieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap"},
		},
		RelRekening: {
			Domain:     "rekening",
			Type:       vernon.RelBelongsTo,
			FK:         FieldRekeningID,
			LocalKey:   FieldRekeningID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"account_number", "balance"},
		},
		RelControlledBy: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         FieldControlledByID,
			LocalKey:   FieldControlledByID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap"},
		},
	}
}

// Validate memvalidasi invariant spending_control.
func (d *Descriptor) Validate(data map[string]any) error {
	return requireString(data, FieldNasabahID, "nasabah")
}

// ── shared validation helpers ─────────────────────────────────────────────────

func requireString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}
