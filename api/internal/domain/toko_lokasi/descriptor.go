// Package toko_lokasi adalah domain Vernon untuk lokasi toko/kantin.
//
// Root master — tidak memiliki BelongsTo.
// Consumer: toko_penjualan (autoloads location), toko_stok_movement, toko_opname.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_lokasi

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field constants ──────────────────────────────────────────────────────────

const (
	FieldName       = "name"
	FieldType       = "location_type"
	FieldIsActive   = "is_active"
)

// ── Enum constants ───────────────────────────────────────────────────────────

const (
	LTypeToko         = "toko"
	LTypeKantin       = "kantin"
	LTypeKantinAsrama = "kantin_asrama"
	LTypeKiosk        = "kiosk"
)

var validLocationTypes = map[string]bool{
	LTypeToko: true, LTypeKantin: true,
	LTypeKantinAsrama: true, LTypeKiosk: true,
}

// ── Descriptor ───────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk toko_lokasi.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "toko_lokasi" }

// DefaultRels — toko_lokasi tidak memiliki relasi (root master).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant toko_lokasi.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldName, "nama lokasi"); err != nil {
		return err
	}
	return requireEnum(data, FieldType, "tipe lokasi", validLocationTypes)
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
