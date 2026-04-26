// Package toko_pembelian adalah domain Vernon untuk purchase order toko/kantin.
//
// Terdiri dari 2 tabel: toko_pembelian, toko_pembelian_item.
// toko_pembelian autoloads: supplier.
// toko_pembelian_item autoloads: pembelian, produk.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_pembelian

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Pembelian field constants ────────────────────────────────────────────────

const (
	FieldSupplierID  = "supplier_id"
	FieldPONumber    = "po_number"
	FieldOrderDate   = "order_date"
	FieldStatus      = "status"
)

// ── Item field constants ────────────────────────────────────────────────────

const (
	ItemFieldPembelianID = "pembelian_id"
	ItemFieldProdukID    = "produk_id"
	ItemFieldQuantity    = "quantity"
	ItemFieldUnitPrice   = "unit_price"
)

// ── Pembelian enum constants ────────────────────────────────────────────────

const (
	StatusDraft    = "draft"
	StatusSent     = "sent"
	StatusPartial  = "partial"
	StatusReceived = "received"
	StatusCancelled = "cancelled"
)

var validStatuses = map[string]bool{
	StatusDraft: true, StatusSent: true, StatusPartial: true,
	StatusReceived: true, StatusCancelled: true,
}

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelSupplier  = "supplier"
	RelPembelian = "pembelian"
	RelProduk    = "produk"
)

// ── PembelianDescriptor ──────────────────────────────────────────────────────

// PembelianDescriptor mengimplementasi vernon.DomainDescriptor untuk toko_pembelian.
type PembelianDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *PembelianDescriptor) TableName() string { return "toko_pembelian" }

// DefaultRels mendefinisikan relasi toko_pembelian: supplier (autoload).
func (d *PembelianDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelSupplier: {
			Domain:     "toko_supplier",
			Type:       vernon.RelBelongsTo,
			FK:         FieldSupplierID,
			LocalKey:   FieldSupplierID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "phone"},
		},
	}
}

// Validate memvalidasi invariant toko_pembelian.
func (d *PembelianDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldSupplierID, "supplier"); err != nil {
		return err
	}
	if err := requireString(data, FieldPONumber, "nomor PO"); err != nil {
		return err
	}
	if err := requireString(data, FieldOrderDate, "tanggal order"); err != nil {
		return err
	}
	return requireEnum(data, FieldStatus, "status", validStatuses)
}

// ── PembelianItemDescriptor ──────────────────────────────────────────────────

// PembelianItemDescriptor mengimplementasi vernon.DomainDescriptor untuk toko_pembelian_item.
type PembelianItemDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *PembelianItemDescriptor) TableName() string { return "toko_pembelian_item" }

// DefaultRels mendefinisikan relasi toko_pembelian_item: pembelian + produk.
func (d *PembelianItemDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelPembelian: {
			Domain:     "toko_pembelian",
			Type:       vernon.RelBelongsTo,
			FK:         ItemFieldPembelianID,
			LocalKey:   ItemFieldPembelianID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"po_number", "status"},
		},
		RelProduk: {
			Domain:     "toko_produk",
			Type:       vernon.RelBelongsTo,
			FK:         ItemFieldProdukID,
			LocalKey:   ItemFieldProdukID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "sku", "cost_price"},
		},
	}
}

// Validate memvalidasi invariant toko_pembelian_item.
func (d *PembelianItemDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, ItemFieldPembelianID, "pembelian"); err != nil {
		return err
	}
	return requireString(data, ItemFieldProdukID, "produk")
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
