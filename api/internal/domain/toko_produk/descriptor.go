// Package toko_produk adalah domain Vernon untuk katalog produk toko/kantin.
//
// Memiliki 1 BelongsTo: category (dari toko_kategori) — TIDAK autoload per ADR.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_produk

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field constants ──────────────────────────────────────────────────────────

const (
	FieldName            = "name"
	FieldSku             = "sku"
	FieldBarcode         = "barcode"
	FieldCategoryID      = "category_id"
	FieldCostPrice       = "cost_price"
	FieldSellPrice       = "sell_price"
	FieldProductType     = "product_type"
	FieldIsActive        = "is_active"
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

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelCategory = "category"
)

// ── Descriptor ───────────────────────────────────────────────────────────────

// Descriptor mengimplementasi vernon.DomainDescriptor untuk toko_produk.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "toko_produk" }

// DefaultRels mendefinisikan relasi toko_produk: category (NOT autoload per ADR — only detail).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelCategory: {
			Domain:     "toko_kategori",
			Type:       vernon.RelBelongsTo,
			FK:         FieldCategoryID,
			LocalKey:   FieldCategoryID,
			ForeignKey: "id",
			IsAutoload: false,
			Fields:     []string{"name", "product_type"},
		},
	}
}

// Validate memvalidasi invariant toko_produk.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldName, "nama produk"); err != nil {
		return err
	}
	if err := requireString(data, FieldSku, "SKU"); err != nil {
		return err
	}
	if err := requireString(data, FieldCategoryID, "kategori"); err != nil {
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
