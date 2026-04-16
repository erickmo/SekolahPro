// Package student_guardian adalah domain Vernon untuk wali/orang tua siswa.
//
// Student guardian belongs_to students dengan autoload. Menyimpan
// data ayah, ibu, atau wali siswa.
//
// Aturan layer Vernon:
//   - Autoload hanya di sisi "many" (student_guardians).
//   - Descriptor tidak boleh import infrastructure/database.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package student_guardian

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID       = "student_id"
	FieldGuardianType    = "guardian_type"
	FieldFullName        = "full_name"
	FieldNIK             = "nik"
	FieldOccupation      = "occupation"
	FieldPhone           = "phone"
	FieldEmail           = "email"
	FieldAddress         = "address"
	FieldIsPrimary       = "is_primary"
	FieldEducationLevel  = "education_level"
)

// Guardian type constants.
const (
	TypeAyah = "ayah"
	TypeIbu  = "ibu"
	TypeWali = "wali"
)

// Relation name constants.
const (
	RelStudent = "student"
)

var validTypes = map[string]bool{
	TypeAyah: true, TypeIbu: true, TypeWali: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk student_guardians.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "student_guardians" }

// DefaultRels mendefinisikan relasi domain ini.
// student_guardians belongs_to students — autoload.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         FieldStudentID,
			LocalKey:   FieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	studentID, _ := data[FieldStudentID].(string)
	if studentID == "" {
		return errors.New("student_id wajib diisi")
	}

	guardianType, _ := data[FieldGuardianType].(string)
	if guardianType == "" {
		return errors.New("guardian_type wajib diisi")
	}
	if !validTypes[guardianType] {
		return fmt.Errorf("guardian_type tidak valid: %q (harus ayah/ibu/wali)", guardianType)
	}

	fullName, _ := data[FieldFullName].(string)
	if fullName == "" {
		return errors.New("full_name wajib diisi")
	}
	return nil
}
