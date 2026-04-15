// Package class_room adalah domain Vernon untuk kelas (ADR-011).
//
// Kelas adalah unit organisasi utama — hampir semua operasional diorganisir per kelas.
// Di-scope per tahun ajaran. Autoload: academic_year + homeroom_teacher.
//
// Aturan layer Vernon:
//   - Autoload hanya di sisi "many" (class_rooms), bukan sisi "one" (academic_years).
//   - Descriptor tidak boleh import infrastructure/database.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package class_room

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldName               = "name"
	FieldGradeLevel         = "grade_level"
	FieldParallelID         = "parallel_id"
	FieldCapacity           = "capacity"
	FieldCurrentCount       = "current_count"
	FieldMajor              = "major"
	FieldIsActive           = "is_active"
	FieldAcademicYearID     = "academic_year_id"
	FieldHomeroomTeacherID  = "homeroom_teacher_id"
)

// Relation name constants.
const (
	RelAcademicYear    = "academic_year"
	RelHomeroomTeacher = "homeroom_teacher"
)

// Default capacity per kelas (standar Kemendikbud).
const DefaultCapacity = 36

var validGradeLevels = map[string]bool{
	"1": true, "2": true, "3": true, "4": true, "5": true, "6": true,
	"7": true, "8": true, "9": true,
	"10": true, "11": true, "12": true,
}

var validMajors = map[string]bool{
	"ipa": true, "ips": true, "bahasa": true,
	"agama": true, "umum": true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk class_rooms.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "class_rooms" }

// DefaultRels mendefinisikan relasi domain ini.
// class_rooms belongs_to academic_years dan teachers (homeroom).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearID,
			LocalKey:   FieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code", "is_active"},
		},
		RelHomeroomTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         FieldHomeroomTeacherID,
			LocalKey:   FieldHomeroomTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "nuptk"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateRequired(data); err != nil {
		return err
	}
	if err := validateCapacity(data); err != nil {
		return err
	}
	return validateMajor(data)
}

// validateRequired memeriksa field wajib.
func validateRequired(data map[string]any) error {
	name, _ := data[FieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}

	gradeLevel, _ := data[FieldGradeLevel].(string)
	if gradeLevel == "" {
		return errors.New("grade_level wajib diisi")
	}
	if !validGradeLevels[gradeLevel] {
		return fmt.Errorf("grade_level tidak valid: %q (harus 1-12)", gradeLevel)
	}

	yearID, _ := data[FieldAcademicYearID].(string)
	if yearID == "" {
		return errors.New("academic_year_id wajib diisi")
	}
	return nil
}

// validateCapacity memeriksa invariant kapasitas kelas.
func validateCapacity(data map[string]any) error {
	capacity, ok := data[FieldCapacity]
	if ok && capacity != nil {
		switch v := capacity.(type) {
		case float64:
			if v <= 0 {
				return errors.New("capacity harus lebih dari 0")
			}
		}
	}

	currentCount, ok := data[FieldCurrentCount]
	if ok && currentCount != nil {
		switch v := currentCount.(type) {
		case float64:
			if v < 0 {
				return errors.New("current_count tidak boleh negatif")
			}
		}
	}
	return nil
}

// validateMajor memeriksa penjurusan (hanya SMA kelas 10-12).
func validateMajor(data map[string]any) error {
	major, _ := data[FieldMajor].(string)
	if major == "" {
		return nil
	}
	if !validMajors[major] {
		return fmt.Errorf("major tidak valid: %q (harus ipa/ips/bahasa/agama/umum)", major)
	}
	return nil
}
