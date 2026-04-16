// Package student_class_placement adalah domain Vernon untuk penempatan kelas siswa.
//
// Student class placement memiliki 3 BelongsTo autoload: student, class_room,
// dan academic_year. Digunakan untuk tracking penempatan siswa per semester.
//
// Aturan layer Vernon:
//   - Autoload hanya di sisi "many" (student_class_placements).
//   - Descriptor tidak boleh import infrastructure/database.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package student_class_placement

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID      = "student_id"
	FieldClassRoomID    = "class_room_id"
	FieldAcademicYearID = "academic_year_id"
	FieldSemester       = "semester"
	FieldStatus         = "status"
	FieldEnrollmentDate = "enrollment_date"
	FieldExitDate       = "exit_date"
	FieldExitReason     = "exit_reason"
)

// Status constants.
const (
	StatusActive   = "active"
	StatusMoved    = "moved"
	StatusGraduated = "graduated"
)

// Relation name constants.
const (
	RelStudent      = "student"
	RelClassRoom    = "class_room"
	RelAcademicYear = "academic_year"
)

var validStatuses = map[string]bool{
	StatusActive: true, StatusMoved: true, StatusGraduated: true,
}

var validSemesters = map[string]bool{
	"1": true, "2": true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk student_class_placements.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "student_class_placements" }

// DefaultRels mendefinisikan relasi domain ini.
// 3 BelongsTo autoload: student, class_room, academic_year.
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
		RelClassRoom: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         FieldClassRoomID,
			LocalKey:   FieldClassRoomID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "grade_level"},
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

	classRoomID, _ := data[FieldClassRoomID].(string)
	if classRoomID == "" {
		return fmt.Errorf("class_room_id wajib diisi")
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
		return fmt.Errorf("status tidak valid: %q (harus active/moved/graduated)", status)
	}
	return nil
}
