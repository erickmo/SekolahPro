// Package toko_penjualan adalah domain Vernon untuk penjualan POS toko/kantin.
//
// Terdiri dari 2 tabel: toko_penjualan, toko_penjualan_item.
// toko_penjualan autoloads: location, nasabah (if nasabah_id present).
// toko_penjualan_item autoloads: penjualan, produk.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_penjualan

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Penjualan field constants ────────────────────────────────────────────────

const (
	FieldLocationID   = "location_id"
	FieldReceiptNumber = "receipt_number"
	FieldSaleDate     = "sale_date"
	FieldSaleType     = "sale_type"
	FieldNasabahID    = "nasabah_id"
	FieldPaymentMethod = "payment_method"
	FieldTotalAmount  = "total_amount"
	FieldStatus       = "status"
)

// ── Item field constants ────────────────────────────────────────────────────

const (
	ItemFieldPenjualanID = "penjualan_id"
	ItemFieldProdukID    = "produk_id"
	ItemFieldQuantity    = "quantity"
	ItemFieldUnitPrice   = "unit_price"
	ItemFieldTotalAmount = "total_amount"
)

// ── Penjualan enum constants ────────────────────────────────────────────────

const (
	SaleTypeStore   = "store"
	SaleTypeCanteen = "canteen"
)

const (
	PaymentCash           = "cash"
	PaymentTabunganDebit  = "tabungan_debit"
	PaymentEwalletDebit   = "ewallet_debit"
	PaymentMixed          = "mixed"
)

const (
	StatusCompleted = "completed"
	StatusVoided    = "voided"
)

var validSaleTypes = map[string]bool{
	SaleTypeStore: true, SaleTypeCanteen: true,
}

var validPaymentMethods = map[string]bool{
	PaymentCash: true, PaymentTabunganDebit: true,
	PaymentEwalletDebit: true, PaymentMixed: true,
}

var validStatuses = map[string]bool{
	StatusCompleted: true, StatusVoided: true,
}

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelLocation = "location"
	RelNasabah  = "nasabah"
	RelPenjualan = "penjualan"
	RelProduk    = "produk"
)

// ── PenjualanDescriptor ──────────────────────────────────────────────────────

// PenjualanDescriptor mengimplementasi vernon.DomainDescriptor untuk toko_penjualan.
type PenjualanDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *PenjualanDescriptor) TableName() string { return "toko_penjualan" }

// DefaultRels mendefinisikan relasi toko_penjualan: location + nasabah.
func (d *PenjualanDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelLocation: {
			Domain:     "toko_lokasi",
			Type:       vernon.RelBelongsTo,
			FK:         FieldLocationID,
			LocalKey:   FieldLocationID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "location_type"},
		},
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         FieldNasabahID,
			LocalKey:   FieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap"},
		},
	}
}

// Validate memvalidasi invariant toko_penjualan.
func (d *PenjualanDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldLocationID, "lokasi"); err != nil {
		return err
	}
	if err := requireString(data, FieldReceiptNumber, "nomor receipt"); err != nil {
		return err
	}
	if err := requireEnum(data, FieldSaleType, "tipe penjualan", validSaleTypes); err != nil {
		return err
	}
	return requireEnum(data, FieldPaymentMethod, "metode pembayaran", validPaymentMethods)
}

// ── ItemDescriptor ───────────────────────────────────────────────────────────

// ItemDescriptor mengimplementasi vernon.DomainDescriptor untuk toko_penjualan_item.
type ItemDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ItemDescriptor) TableName() string { return "toko_penjualan_item" }

// DefaultRels mendefinisikan relasi toko_penjualan_item: penjualan + produk.
func (d *ItemDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelPenjualan: {
			Domain:     "toko_penjualan",
			Type:       vernon.RelBelongsTo,
			FK:         ItemFieldPenjualanID,
			LocalKey:   ItemFieldPenjualanID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"receipt_number", "sale_type", "status"},
		},
		RelProduk: {
			Domain:     "toko_produk",
			Type:       vernon.RelBelongsTo,
			FK:         ItemFieldProdukID,
			LocalKey:   ItemFieldProdukID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "sku", "sell_price"},
		},
	}
}

// Validate memvalidasi invariant toko_penjualan_item.
func (d *ItemDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, ItemFieldPenjualanID, "penjualan"); err != nil {
		return err
	}
	if err := requireString(data, ItemFieldProdukID, "produk"); err != nil {
		return err
	}
	return nil
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
