// Package produk_akad adalah domain Vernon untuk produk akad koperasi.
//
// Produk akad mendefinisikan jenis simpanan dan pembiayaan yang ditawarkan koperasi.
// Entity root — tidak punya belongs_to ke domain lain.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package produk_akad

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldName        = "name"
	FieldCode        = "code"
	FieldType        = "type"
	FieldAkadType    = "akad_type"
	FieldMinAmount   = "min_amount"
	FieldMaxAmount   = "max_amount"
	FieldMinTenor    = "min_tenor"
	FieldMaxTenor    = "max_tenor"
	FieldTenorUnit   = "tenor_unit"
	FieldInterestRate = "interest_rate"
	FieldDescription = "description"
	FieldStatus      = "status"
)

// Type constants.
const (
	TypeSimpanan   = "simpanan"
	TypePembiayaan = "pembiayaan"
)

// AkadType constants.
const (
	AkadWadiah     = "wadiah"
	AkadMudharabah = "mudharabah"
	AkadMusyarakah = "musyarakah"
	AkadMurabahah  = "murabahah"
	AkadIjarah     = "ijarah"
	AkadQard       = "qard"
)

var validTypes = map[string]bool{
	TypeSimpanan: true, TypePembiayaan: true,
}

var validAkadTypes = map[string]bool{
	AkadWadiah: true, AkadMudharabah: true, AkadMusyarakah: true,
	AkadMurabahah: true, AkadIjarah: true, AkadQard: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk produk_akad.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "produk_akad" }

// DefaultRels mendefinisikan relasi domain ini.
// produk_akad adalah entity root — tidak punya belongs_to.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateRequired(data); err != nil {
		return err
	}
	if err := validateEnums(data); err != nil {
		return err
	}
	return validateAmounts(data)
}

// validateRequired memeriksa field wajib.
func validateRequired(data map[string]any) error {
	name, _ := data[FieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}
	code, _ := data[FieldCode].(string)
	if code == "" {
		return errors.New("code wajib diisi")
	}
	return nil
}

// validateEnums memeriksa enum fields.
func validateEnums(data map[string]any) error {
	typ, _ := data[FieldType].(string)
	if typ == "" {
		return errors.New("type wajib diisi")
	}
	if !validTypes[typ] {
		return fmt.Errorf("type tidak valid: %q (harus simpanan/pembiayaan)", typ)
	}

	akadType, _ := data[FieldAkadType].(string)
	if akadType == "" {
		return errors.New("akad_type wajib diisi")
	}
	if !validAkadTypes[akadType] {
		return fmt.Errorf("akad_type tidak valid: %q", akadType)
	}
	return nil
}

// validateAmounts memeriksa invariant amount.
func validateAmounts(data map[string]any) error {
	minAmount, minOK := data[FieldMinAmount].(float64)
	maxAmount, maxOK := data[FieldMaxAmount].(float64)

	if minOK && minAmount < 0 {
		return errors.New("min_amount tidak boleh negatif")
	}
	if maxOK && maxAmount < 0 {
		return errors.New("max_amount tidak boleh negatif")
	}
	if minOK && maxOK && maxAmount < minAmount {
		return errors.New("max_amount harus >= min_amount")
	}
	return nil
}
