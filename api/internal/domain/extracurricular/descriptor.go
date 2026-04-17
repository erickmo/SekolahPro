// Package extracurricular adalah domain Vernon untuk kegiatan ekstrakurikuler siswa.
//
// Extracurricular memiliki 2 BelongsTo autoload: student dan teacher (pembina).
// Digunakan untuk tracking partisipasi siswa dalam kegiatan ekstrakurikuler.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package extracurricular

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID        = "student_id"
	FieldTeacherID        = "teacher_id"
	FieldActivityName     = "activity_name"
	FieldActivityType     = "activity_type"
	FieldJoinDate         = "join_date"
	FieldExitDate         = "exit_date"
	FieldRole             = "role"
	FieldAchievement      = "achievement"
	FieldSchedule         = "schedule"
	FieldStatus           = "status"
)

// Activity type constants.
const (
	ActivityTypeSport      = "sport"
	ActivityTypeArt        = "art"
	ActivityTypeAcademic   = "academic"
	ActivityTypeSocial     = "social"
	ActivityTypeReligion   = "religion"
	ActivityTypeTechnology = "technology"
)

// Status constants.
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusCompleted = "completed"
)

// Relation name constants.
const (
	RelStudent = "student"
	RelTeacher = "teacher"
)

var validActivityTypes = map[string]bool{
	ActivityTypeSport: true, ActivityTypeArt: true,
	ActivityTypeAcademic: true, ActivityTypeSocial: true,
	ActivityTypeReligion: true, ActivityTypeTechnology: true,
}

var validStatuses = map[string]bool{
	StatusActive: true, StatusInactive: true, StatusCompleted: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk extracurriculars.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "extracurriculars" }

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

	activityType, _ := data[FieldActivityType].(string)
	if activityType != "" && !validActivityTypes[activityType] {
		return fmt.Errorf("activity_type tidak valid: %q", activityType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
