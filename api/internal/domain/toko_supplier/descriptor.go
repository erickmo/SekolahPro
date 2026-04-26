// Package toko_supplier adalah domain Vernon untuk supplier toko/kantin.
//
// Root master — tidak memiliki BelongsTo.
// Consumer: toko_pembelian (autoloads supplier dari domain ini).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_supplier

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field constants ──────────────────────────────────────────────────────────

const (
	FieldName       = "name"
	FieldPhone      = "phone"
	FieldIsActive   = "is_active"
)

// ── Descriptor ───────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk toko_supplier.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "toko_supplier" }

// DefaultRels — toko_supplier tidak memiliki relasi (root master).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant toko_supplier.
func (d *Descriptor) Validate(data map[string]any) error {
	return requireString(data, FieldName, "nama supplier")
}

// ── shared validation helpers ─────────────────────────────────────────────────

func requireString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}
