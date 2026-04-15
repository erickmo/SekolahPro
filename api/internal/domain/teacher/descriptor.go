// Package teacher adalah domain Vernon untuk guru dan tenaga kependidikan (ADR-012).
//
// Satu tabel menampung semua jenis personel sekolah (guru + staff),
// dibedakan oleh field "role". Teachers adalah entity root — tidak punya
// belongs_to ke domain lain, tapi menjadi sumber _data untuk 7+ domain.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teacher

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldFullName     = "full_name"
	FieldNIP          = "nip"
	FieldNUPTK        = "nuptk"
	FieldGender       = "gender"
	FieldBirthPlace   = "birth_place"
	FieldBirthDate    = "birth_date"
	FieldReligion     = "religion"
	FieldPhone        = "phone"
	FieldEmail        = "email"
	FieldPhotoURL     = "photo_url"
	FieldEmployeeType = "employee_type"
	FieldRole         = "role"
	FieldJoinDate     = "join_date"
	FieldResignDate   = "resign_date"
	FieldStatus       = "status"
	FieldSignatureURL = "signature_url"
	FieldUserID       = "user_id"
)

// Status constants.
const (
	StatusActive   = "active"
	StatusOnLeave  = "on_leave"
	StatusResigned = "resigned"
	StatusRetired  = "retired"
)

// Gender constants.
const (
	GenderMale   = "L"
	GenderFemale = "P"
)

var validGenders = map[string]bool{
	GenderMale: true, GenderFemale: true,
}

var validEmployeeTypes = map[string]bool{
	"pns": true, "p3k": true, "honorer": true,
	"yayasan": true, "kontrak": true,
}

var validRoles = map[string]bool{
	"kepala_sekolah":    true,
	"wakasek_kurikulum": true,
	"wakasek_kesiswaan": true,
	"wakasek_sarana":    true,
	"wakasek_humas":     true,
	"guru_mapel":        true,
	"guru_bk":           true,
	"guru_piket":        true,
	"admin_tu":          true,
	"bendahara":         true,
	"pustakawan":        true,
	"staff_umum":        true,
}

var validStatuses = map[string]bool{
	StatusActive: true, StatusOnLeave: true,
	StatusResigned: true, StatusRetired: true,
}

var validReligions = map[string]bool{
	"islam": true, "kristen": true, "katolik": true,
	"hindu": true, "buddha": true, "konghucu": true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk teachers.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "teachers" }

// DefaultRels mendefinisikan relasi domain ini.
// teachers adalah entity root — tidak punya belongs_to.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateIdentity(data); err != nil {
		return err
	}
	return validateEnums(data)
}

// validateIdentity memeriksa field identitas wajib.
func validateIdentity(data map[string]any) error {
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
	empType, _ := data[FieldEmployeeType].(string)
	if empType == "" {
		return errors.New("employee_type wajib diisi")
	}
	if !validEmployeeTypes[empType] {
		return fmt.Errorf("employee_type tidak valid: %q", empType)
	}

	role, _ := data[FieldRole].(string)
	if role == "" {
		return errors.New("role wajib diisi")
	}
	if !validRoles[role] {
		return fmt.Errorf("role tidak valid: %q", role)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	religion, _ := data[FieldReligion].(string)
	if religion != "" && !validReligions[religion] {
		return fmt.Errorf("religion tidak valid: %q", religion)
	}
	return nil
}
