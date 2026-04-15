// Package product adalah contoh domain Vernon — tipe: belongs_to dengan autoload.
//
// Domain ini mendemonstrasikan fitur autoload Vernon:
// saat products dibuat/diupdate, field name+slug dari product_categories
// otomatis di-embed ke dalam _data["category"], sehingga GET products
// tidak perlu JOIN ke tabel categories.
//
// Aturan layer Vernon:
//   - Autoload hanya di sisi "many" (products), bukan sisi "one" (categories).
//   - Descriptor tidak boleh import infrastructure/database.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package product

import (
	"errors"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldName       = "name"
	FieldSKU        = "sku"
	FieldPrice      = "price"
	FieldStock      = "stock"
	FieldIsActive   = "is_active"
	FieldImageURL   = "image_url"
	FieldCategoryID = "category_id"
)

// Relation name constants.
const RelCategory = "category"

// Descriptor mengimplementasi vernon.DomainDescriptor untuk products.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "products" }

// DefaultRels mendefinisikan relasi domain ini.
// products belongs_to product_categories — autoload name dan slug.
// Setiap call mengembalikan map literal baru (defensive copy).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelCategory: {
			Domain:     "product_categories",
			Type:       vernon.RelBelongsTo,
			FK:         FieldCategoryID, // field di _data products yang menyimpan UUID category
			LocalKey:   FieldCategoryID, // key di sisi products
			ForeignKey: "id",            // key di sisi product_categories
			IsAutoload: true,            // embed name+slug ke _data["category"] saat write
			Fields:     []string{FieldName, FieldSKU},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	name, _ := data[FieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}

	sku, _ := data[FieldSKU].(string)
	if sku == "" {
		return errors.New("sku wajib diisi")
	}

	// price adalah invariant: produk valid harus punya harga >= 0.
	switch v := data[FieldPrice].(type) {
	case float64:
		if v < 0 {
			return errors.New("price tidak boleh negatif")
		}
	case nil:
		return errors.New("price wajib diisi")
	}

	// stock >= 0 adalah invariant logistik.
	if stock, ok := data[FieldStock]; ok && stock != nil {
		switch v := stock.(type) {
		case float64:
			if v < 0 {
				return errors.New("stock tidak boleh negatif")
			}
		}
	}

	return nil
}
