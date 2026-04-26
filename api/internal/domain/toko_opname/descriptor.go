// Package toko_opname adalah domain Vernon untuk stock opname toko/kantin.
//
// Terdiri dari 2 tabel: toko_opname, toko_opname_item.
// toko_opname tidak memiliki autoload.
// toko_opname_item autoloads: opname, produk.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package toko_opname

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Opname field constants ──────────────────────────────────────────────────

const (
	FieldLokasiID   = "lokasi_id"
	FieldOpnameDate = "opname_date"
	FieldStatus     = "status"
)

// ── Item field constants ────────────────────────────────────────────────────

const (
	ItemFieldOpnameID    = "opname_id"
	ItemFieldProdukID    = "produk_id"
	ItemFieldSystemStock = "system_stock"
	ItemFieldPhysicalStock = "physical_stock"
)

// ── Opname enum constants ───────────────────────────────────────────────────

const (
	StatusDraft       = "draft"
	StatusInProgress  = "in_progress"
	StatusCompleted   = "completed"
	StatusCancelled   = "cancelled"
)

var validStatuses = map[string]bool{
	StatusDraft: true, StatusInProgress: true,
	StatusCompleted: true, StatusCancelled: true,
}

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelOpname = "opname"
	RelProduk = "produk"
)

// ── OpnameDescriptor ─────────────────────────────────────────────────────────

// OpnameDescriptor mengimplementasi vernon.DomainDescriptor untuk toko_opname.
type OpnameDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *OpnameDescriptor) TableName() string { return "toko_opname" }

// DefaultRels — toko_opname tidak memiliki autoload.
func (d *OpnameDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant toko_opname.
func (d *OpnameDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldLokasiID, "lokasi"); err != nil {
		return err
	}
	if err := requireString(data, FieldOpnameDate, "tanggal opname"); err != nil {
		return err
	}
	return requireEnum(data, FieldStatus, "status", validStatuses)
}

// ── OpnameItemDescriptor ─────────────────────────────────────────────────────

// OpnameItemDescriptor mengimplementasi vernon.DomainDescriptor untuk toko_opname_item.
type OpnameItemDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *OpnameItemDescriptor) TableName() string { return "toko_opname_item" }

// DefaultRels mendefinisikan relasi toko_opname_item: opname + produk.
func (d *OpnameItemDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelOpname: {
			Domain:     "toko_opname",
			Type:       vernon.RelBelongsTo,
			FK:         ItemFieldOpnameID,
			LocalKey:   ItemFieldOpnameID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"opname_date", "status"},
		},
		RelProduk: {
			Domain:     "toko_produk",
			Type:       vernon.RelBelongsTo,
			FK:         ItemFieldProdukID,
			LocalKey:   ItemFieldProdukID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "sku"},
		},
	}
}

// Validate memvalidasi invariant toko_opname_item.
func (d *OpnameItemDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, ItemFieldOpnameID, "opname"); err != nil {
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
