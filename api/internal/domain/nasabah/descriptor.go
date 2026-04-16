// Package nasabah adalah domain Vernon untuk nasabah koperasi sekolah.
//
// Nasabah adalah entity root yang merepresentasikan anggota koperasi.
// Dibedakan berdasarkan type: siswa, guru, staff, umum.
// Lifecycle: active → inactive/blacklisted.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package nasabah

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldNIK             = "nik"
	FieldFullName        = "full_name"
	FieldNoNasabah       = "no_nasabah"
	FieldType            = "type"
	FieldGender          = "gender"
	FieldBirthPlace      = "birth_place"
	FieldBirthDate       = "birth_date"
	FieldPhone           = "phone"
	FieldEmail           = "email"
	FieldAddress         = "address"
	FieldStatus          = "status"
	FieldJoinDate        = "join_date"
	FieldMotherMaidenName = "mother_maiden_name"
)

// Type constants.
const (
	TypeSiswa = "siswa"
	TypeGuru  = "guru"
	TypeStaff = "staff"
	TypeUmum  = "umum"
)

// Gender constants.
const (
	GenderL = "L"
	GenderP = "P"
)

// Status constants.
const (
	StatusActive      = "active"
	StatusInactive    = "inactive"
	StatusBlacklisted = "blacklisted"
)

// validTypes berisi semua type yang valid.
var validTypes = map[string]bool{
	TypeSiswa: true, TypeGuru: true, TypeStaff: true, TypeUmum: true,
}

// validGenders berisi semua gender yang valid.
var validGenders = map[string]bool{
	GenderL: true, GenderP: true,
}

// validStatuses berisi semua status yang valid.
var validStatuses = map[string]bool{
	StatusActive: true, StatusInactive: true, StatusBlacklisted: true,
}

// nikRegex validasi NIK 16 digit.
var nikRegex = regexp.MustCompile(`^\d{16}$`)

// Descriptor mengimplementasi vernon.DomainDescriptor untuk nasabah.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "nasabah" }

// DefaultRels mendefinisikan relasi domain ini.
// nasabah adalah entity root — tidak punya belongs_to.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateNIK(data); err != nil {
		return err
	}
	if err := validateType(data); err != nil {
		return err
	}
	if err := validateGender(data); err != nil {
		return err
	}
	return validateStatus(data)
}

// validateNIK memeriksa format NIK.
func validateNIK(data map[string]any) error {
	nik, _ := data[FieldNIK].(string)
	if nik == "" {
		return errors.New("nik wajib diisi")
	}
	if !nikRegex.MatchString(nik) {
		return fmt.Errorf("nik tidak valid: %q (harus 16 digit)", nik)
	}
	return nil
}

// validateType memeriksa type nasabah.
func validateType(data map[string]any) error {
	typ, _ := data[FieldType].(string)
	if typ == "" {
		return errors.New("type wajib diisi")
	}
	if !validTypes[typ] {
		return fmt.Errorf("type tidak valid: %q (harus siswa/guru/staff/umum)", typ)
	}
	return nil
}

// validateGender memeriksa gender.
func validateGender(data map[string]any) error {
	gender, _ := data[FieldGender].(string)
	if gender == "" {
		return errors.New("gender wajib diisi")
	}
	if !validGenders[gender] {
		return fmt.Errorf("gender tidak valid: %q (harus L/P)", gender)
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
		return fmt.Errorf("status tidak valid: %q (harus active/inactive/blacklisted)", status)
	}
	return nil
}
