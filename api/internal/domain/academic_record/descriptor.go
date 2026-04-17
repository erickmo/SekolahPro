// Package academic_record adalah domain Vernon untuk catatan akademik siswa.
//
// AcademicRecord memiliki 2 BelongsTo autoload: student dan academic_year.
// Digunakan untuk tracking nilai semester, rata-rata, dan peringkat siswa.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package academic_record

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID      = "student_id"
	FieldAcademicYearID = "academic_year_id"
	FieldSemester       = "semester"
	FieldGPA            = "gpa"
	FieldRank           = "rank"
	FieldTotalScore     = "total_score"
	FieldStatus         = "status"
	FieldNotes          = "notes"
)

// Status constants.
const (
	StatusPass    = "pass"
	StatusFail    = "fail"
	StatusRemedial = "remedial"
)

// Relation name constants.
const (
	RelStudent      = "student"
	RelAcademicYear = "academic_year"
)

var validStatuses = map[string]bool{
	StatusPass: true, StatusFail: true, StatusRemedial: true,
}

var validSemesters = map[string]bool{
	"1": true, "2": true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk academic_records.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "academic_records" }

// DefaultRels mendefinisikan relasi domain ini.
// 2 BelongsTo autoload: student, academic_year.
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
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearID,
			LocalKey:   FieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	studentID, _ := data[FieldStudentID].(string)
	if studentID == "" {
		return fmt.Errorf("student_id wajib diisi")
	}

	academicYearID, _ := data[FieldAcademicYearID].(string)
	if academicYearID == "" {
		return fmt.Errorf("academic_year_id wajib diisi")
	}

	semester, _ := data[FieldSemester].(string)
	if semester != "" && !validSemesters[semester] {
		return fmt.Errorf("semester tidak valid: %q (harus 1 atau 2)", semester)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
