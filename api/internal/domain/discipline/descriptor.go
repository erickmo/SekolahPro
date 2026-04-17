// Package discipline adalah domain Vernon untuk catatan pelanggaran siswa.
//
// Discipline memiliki 2 BelongsTo autoload: student dan teacher (pelapor).
// Digunakan untuk tracking pelanggaran, sanksi, dan tindak lanjut.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package discipline

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID    = "student_id"
	FieldTeacherID    = "teacher_id"
	FieldViolationType = "violation_type"
	FieldIncidentDate = "incident_date"
	FieldDescription  = "description"
	FieldSanction     = "sanction"
	FieldSanctionDate = "sanction_date"
	FieldPoints       = "points"
	FieldStatus       = "status"
)

// Violation type constants.
const (
	ViolationMinor  = "minor"
	ViolationModerate = "moderate"
	ViolationMajor  = "major"
)

// Status constants.
const (
	StatusReported  = "reported"
	StatusReviewed  = "reviewed"
	StatusSanctioned = "sanctioned"
	StatusResolved  = "resolved"
)

// Relation name constants.
const (
	RelStudent = "student"
	RelTeacher = "teacher"
)

var validViolationTypes = map[string]bool{
	ViolationMinor: true, ViolationModerate: true, ViolationMajor: true,
}

var validStatuses = map[string]bool{
	StatusReported: true, StatusReviewed: true,
	StatusSanctioned: true, StatusResolved: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk disciplines.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "disciplines" }

// DefaultRels mendefinisikan relasi domain ini.
// 2 BelongsTo autoload: student, teacher.
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
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTeacherID,
			LocalKey:   FieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	studentID, _ := data[FieldStudentID].(string)
	if studentID == "" {
		return fmt.Errorf("student_id wajib diisi")
	}

	teacherID, _ := data[FieldTeacherID].(string)
	if teacherID == "" {
		return fmt.Errorf("teacher_id wajib diisi")
	}

	violationType, _ := data[FieldViolationType].(string)
	if violationType != "" && !validViolationTypes[violationType] {
		return fmt.Errorf("violation_type tidak valid: %q", violationType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
