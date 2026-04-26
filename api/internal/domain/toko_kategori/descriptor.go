// Package toko_kategori adalah domain Vernon untuk kategori produk toko/kantin.
//
// Root master — tidak memiliki BelongsTo.
// Consumer: toko_produk (autoloads category dari domain ini).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_kategori

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field constants ──────────────────────────────────────────────────────────

const (
	FieldName       = "name"
	FieldProductType = "product_type"
	FieldParentID   = "parent_id"
	FieldSortOrder  = "sort_order"
	FieldIsActive   = "is_active"
)

// ── Enum constants ───────────────────────────────────────────────────────────

const (
	PTypeStoreItem   = "store_item"
	PTypeCanteenItem = "canteen_item"
)

var validProductTypes = map[string]bool{
	PTypeStoreItem:   true,
	PTypeCanteenItem: true,
}

// ── Descriptor ───────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk toko_kategori.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "toko_kategori" }

// DefaultRels — toko_kategori tidak memiliki relasi (root master).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant toko_kategori.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldName, "nama kategori"); err != nil {
		return err
	}
	return requireEnum(data, FieldProductType, "tipe produk", validProductTypes)
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
