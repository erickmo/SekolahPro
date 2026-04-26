// Package toko_stok_movement adalah domain Vernon untuk pergerakan stok toko/kantin.
//
// Tidak memiliki autoload — relasi bersifat reference only.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_stok_movement

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field constants ──────────────────────────────────────────────────────────

const (
	FieldProdukID      = "produk_id"
	FieldLokasiID      = "lokasi_id"
	FieldMovementType  = "movement_type"
	FieldQuantity      = "quantity"
)

// ── Enum constants ───────────────────────────────────────────────────────────

const (
	MovIn           = "in"
	MovOut          = "out"
	MovAdjustment   = "adjustment"
	MovTransferIn   = "transfer_in"
	MovTransferOut  = "transfer_out"
	MovOpname       = "opname"
)

var validMovementTypes = map[string]bool{
	MovIn: true, MovOut: true, MovAdjustment: true,
	MovTransferIn: true, MovTransferOut: true, MovOpname: true,
}

// ── Descriptor ───────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk toko_stok_movement.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "toko_stok_movement" }

// DefaultRels — NO autoload per ADR.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant toko_stok_movement.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldProdukID, "produk"); err != nil {
		return err
	}
	if err := requireString(data, FieldLokasiID, "lokasi"); err != nil {
		return err
	}
	return requireEnum(data, FieldMovementType, "tipe pergerakan", validMovementTypes)
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
