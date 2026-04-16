// Package student_address adalah domain Vernon untuk alamat siswa.
//
// Student address belongs_to students dengan autoload. Menyimpan
// berbagai jenis alamat: domisili, asal, kost.
//
// Aturan layer Vernon:
//   - Autoload hanya di sisi "many" (student_addresses).
//   - Descriptor tidak boleh import infrastructure/database.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package student_address

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID        = "student_id"
	FieldAddressType      = "address_type"
	FieldStreet           = "street"
	FieldRT               = "rt"
	FieldRW               = "rw"
	FieldVillage          = "village"
	FieldDistrict         = "district"
	FieldCity             = "city"
	FieldProvince         = "province"
	FieldPostalCode       = "postal_code"
	FieldCountry          = "country"
	FieldIsPrimary        = "is_primary"
	FieldPreviousSchoolName = "previous_school_name"
)

// Address type constants.
const (
	TypeDomisili = "domisili"
	TypeAsal     = "asal"
	TypeKost     = "kost"
)

// Relation name constants.
const (
	RelStudent = "student"
)

var validTypes = map[string]bool{
	TypeDomisili: true, TypeAsal: true, TypeKost: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk student_addresses.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "student_addresses" }

// DefaultRels mendefinisikan relasi domain ini.
// student_addresses belongs_to students — autoload.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         FieldStudentID,
			LocalKey:   FieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis", "nisn"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	studentID, _ := data[FieldStudentID].(string)
	if studentID == "" {
		return errors.New("student_id wajib diisi")
	}

	addrType, _ := data[FieldAddressType].(string)
	if addrType == "" {
		return errors.New("address_type wajib diisi")
	}
	if !validTypes[addrType] {
		return fmt.Errorf("address_type tidak valid: %q (harus domisili/asal/kost)", addrType)
	}
	return nil
}
