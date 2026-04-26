// Package ewallet_card adalah domain Vernon untuk kartu belanja uang saku digital.
//
// Autoloads: nasabah, spending_control.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package ewallet_card

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field constants ──────────────────────────────────────────────────────────

const (
	FieldNasabahID          = "nasabah_id"
	FieldSpendingControlID  = "spending_control_id"
	FieldCardType           = "card_type"
	FieldCardNumber         = "card_number"
	FieldStatus             = "status"
)

// ── Enum constants ───────────────────────────────────────────────────────────

const (
	CardTypeNFC        = "nfc"
	CardTypeQRStatic   = "qr_static"
	CardTypeQRDynamic  = "qr_dynamic"
	CardTypeBarcode    = "barcode"
)

var validCardTypes = map[string]bool{
	CardTypeNFC: true, CardTypeQRStatic: true,
	CardTypeQRDynamic: true, CardTypeBarcode: true,
}

const (
	CardStatusActive      = "active"
	CardStatusFrozen      = "frozen"
	CardStatusDeactivated = "deactivated"
	CardStatusLost        = "lost"
)

var validCardStatuses = map[string]bool{
	CardStatusActive: true, CardStatusFrozen: true,
	CardStatusDeactivated: true, CardStatusLost: true,
}

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelNasabah         = "nasabah"
	RelSpendingControl = "spending_control"
)

// ── Descriptor ───────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk ewallet_card.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "ewallet_card" }

// DefaultRels mendefinisikan relasi ewallet_card: nasabah + spending_control.
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
		RelSpendingControl: {
			Domain:     "spending_control",
			Type:       vernon.RelBelongsTo,
			FK:         FieldSpendingControlID,
			LocalKey:   FieldSpendingControlID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"daily_limit", "is_active"},
		},
	}
}

// Validate memvalidasi invariant ewallet_card.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldNasabahID, "nasabah"); err != nil {
		return err
	}
	if err := requireString(data, FieldCardNumber, "nomor kartu"); err != nil {
		return err
	}
	if err := requireEnum(data, FieldCardType, "tipe kartu", validCardTypes); err != nil {
		return err
	}
	return requireEnum(data, FieldStatus, "status kartu", validCardStatuses)
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
