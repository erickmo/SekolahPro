// Package product_category adalah contoh domain Vernon — tipe: standalone (tanpa autoload).
//
// Domain ini cocok untuk data referensi yang sering di-lookup oleh domain lain.
// Tidak punya relasi ke domain lain, sehingga DefaultRels() mengembalikan map kosong.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant yang bisa dicek dari data struct sendiri.
//   - Tidak ada DB/API call di dalam package ini.
package product_category

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants — hindari magic string.
const (
	FieldName        = "name"
	FieldSlug        = "slug"
	FieldDescription = "description"
	FieldIsActive    = "is_active"
	FieldSortOrder   = "sort_order"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

// Descriptor mengimplementasi vernon.DomainDescriptor untuk product_categories.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "product_categories" }

// DefaultRels mendefinisikan relasi domain ini.
// product_categories adalah domain sumber — tidak punya belongs_to ke domain lain.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain sebelum write.
// Hanya invariant struktural — tidak ada DB/API call.
func (d *Descriptor) Validate(data map[string]any) error {
	name, _ := data[FieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}

	slug, _ := data[FieldSlug].(string)
	if slug == "" {
		return errors.New("slug wajib diisi")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug hanya boleh mengandung huruf kecil, angka, dan tanda hubung (-), got: %q", slug)
	}

	if sortOrder, ok := data[FieldSortOrder]; ok && sortOrder != nil {
		switch v := sortOrder.(type) {
		case float64:
			if v < 0 {
				return errors.New("sort_order tidak boleh negatif")
			}
		case int:
			if v < 0 {
				return errors.New("sort_order tidak boleh negatif")
			}
		}
	}

	return nil
}
