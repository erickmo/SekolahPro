// Package student adalah domain Vernon untuk siswa.
//
// Student adalah entity root — referensi utama untuk hampir semua domain sekolah.
// Direferensikan oleh 15+ domain (kehadiran, nilai, kesehatan, disiplin, dll).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package student

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldNIS           = "nis"
	FieldNISN          = "nisn"
	FieldFullName      = "full_name"
	FieldGender        = "gender"
	FieldBirthPlace    = "birth_place"
	FieldBirthDate     = "birth_date"
	FieldReligion      = "religion"
	FieldAddress       = "address"
	FieldPhone         = "phone"
	FieldEmail         = "email"
	FieldPhotoURL      = "photo_url"
	FieldBloodType     = "blood_type"
	FieldEnrollmentDate = "enrollment_date"
	FieldStatus        = "status"
	FieldPreviousSchool = "previous_school"
	FieldAdmissionType = "admission_type"
	FieldNotes         = "notes"
)

// Status constants.
const (
	StatusActive     = "active"
	StatusInactive   = "inactive"
	StatusGraduated  = "graduated"
	StatusTransferred = "transferred"
	StatusDroppedOut = "dropped_out"
)

// Gender constants.
const (
	GenderL = "L"
	GenderP = "P"
)

var validStatuses = map[string]bool{
	StatusActive: true, StatusInactive: true, StatusGraduated: true,
	StatusTransferred: true, StatusDroppedOut: true,
}

var validGenders = map[string]bool{
	GenderL: true, GenderP: true,
}

var validBloodTypes = map[string]bool{
	"A": true, "B": true, "AB": true, "O": true,
}

var validReligions = map[string]bool{
	"islam": true, "kristen": true, "katolik": true,
	"hindu": true, "buddha": true, "konghucu": true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk students.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "students" }

// DefaultRels mendefinisikan relasi domain ini.
// students adalah entity root — tidak punya belongs_to.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateRequired(data); err != nil {
		return err
	}
	return validateEnums(data)
}

// validateRequired memeriksa field wajib.
func validateRequired(data map[string]any) error {
	fullName, _ := data[FieldFullName].(string)
	if fullName == "" {
		return errors.New("full_name wajib diisi")
	}

	gender, _ := data[FieldGender].(string)
	if gender == "" {
		return errors.New("gender wajib diisi")
	}
	if !validGenders[gender] {
		return fmt.Errorf("gender tidak valid: %q (harus L/P)", gender)
	}
	return nil
}

// validateEnums memeriksa enum fields.
func validateEnums(data map[string]any) error {
	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	bloodType, _ := data[FieldBloodType].(string)
	if bloodType != "" && !validBloodTypes[bloodType] {
		return fmt.Errorf("blood_type tidak valid: %q (harus A/B/AB/O)", bloodType)
	}

	religion, _ := data[FieldReligion].(string)
	if religion != "" && !validReligions[religion] {
		return fmt.Errorf("religion tidak valid: %q", religion)
	}
	return nil
}
