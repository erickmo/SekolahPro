// Package rekening adalah domain Vernon untuk rekening koperasi sekolah.
//
// Rekening menghubungkan nasabah dengan produk akad. Memiliki 2 BelongsTo:
// nasabah dan produk_akad — keduanya autoload untuk read performance.
//
// Aturan layer Vernon:
//   - Autoload hanya di sisi "many" (rekening), bukan sisi "one" (nasabah/produk_akad).
//   - Descriptor tidak boleh import infrastructure/database.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package rekening

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldNasabahID    = "nasabah_id"
	FieldProdukAkadID = "produk_akad_id"
	FieldNoRekening   = "no_rekening"
	FieldBalance      = "balance"
	FieldHoldBalance  = "hold_balance"
	FieldStatus       = "status"
	FieldOpenDate     = "open_date"
	FieldCloseDate    = "close_date"
)

// Status constants.
const (
	StatusActive = "active"
	StatusFrozen = "frozen"
	StatusClosed = "closed"
)

// Relation name constants.
const (
	RelNasabah    = "nasabah"
	RelProdukAkad = "produk_akad"
)

var validStatuses = map[string]bool{
	StatusActive: true, StatusFrozen: true, StatusClosed: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk rekening.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "rekening" }

// DefaultRels mendefinisikan relasi domain ini.
// rekening belongs_to nasabah dan produk_akad — keduanya autoload.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         FieldNasabahID,
			LocalKey:   FieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "no_nasabah", "type"},
		},
		RelProdukAkad: {
			Domain:     "produk_akad",
			Type:       vernon.RelBelongsTo,
			FK:         FieldProdukAkadID,
			LocalKey:   FieldProdukAkadID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code", "type", "akad_type"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateBalance(data); err != nil {
		return err
	}
	if err := validateHoldBalance(data); err != nil {
		return err
	}
	return validateStatus(data)
}

// validateBalance memeriksa balance >= 0.
func validateBalance(data map[string]any) error {
	balance, ok := data[FieldBalance].(float64)
	if ok && balance < 0 {
		return fmt.Errorf("balance tidak boleh negatif")
	}
	return nil
}

// validateHoldBalance memeriksa hold_balance <= balance.
func validateHoldBalance(data map[string]any) error {
	hold, holdOK := data[FieldHoldBalance].(float64)
	balance, balanceOK := data[FieldBalance].(float64)
	if holdOK && balanceOK && hold > balance {
		return fmt.Errorf("hold_balance tidak boleh lebih besar dari balance")
	}
	return nil
}

// validateStatus memeriksa status jika disediakan.
func validateStatus(data map[string]any) error {
	status, _ := data[FieldStatus].(string)
	if status == "" {
		return nil
	}
	if !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus active/frozen/closed)", status)
	}
	return nil
}
