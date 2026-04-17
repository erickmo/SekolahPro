// Package subject_grade adalah domain Vernon untuk nilai mata pelajaran siswa.
//
// SubjectGrade memiliki 3 BelongsTo autoload: student, subject, dan teacher.
// Digunakan untuk tracking nilai per mata pelajaran per semester.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package subject_grade

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldStudentID    = "student_id"
	FieldSubjectID    = "subject_id"
	FieldTeacherID    = "teacher_id"
	FieldGrade        = "grade"
	FieldGradeType    = "grade_type"
	FieldExamType     = "exam_type"
	FieldWeight       = "weight"
	FieldDescription  = "description"
)

// Grade type constants.
const (
	GradeTypeAssignment = "assignment"
	GradeTypeMidterm    = "midterm"
	GradeTypeFinal      = "final"
	GradeTypeQuiz       = "quiz"
	GradeTypeProject    = "project"
)

// Relation name constants.
const (
	RelStudent = "student"
	RelSubject = "subject"
	RelTeacher = "teacher"
)

var validGradeTypes = map[string]bool{
	GradeTypeAssignment: true, GradeTypeMidterm: true,
	GradeTypeFinal: true, GradeTypeQuiz: true, GradeTypeProject: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk subject_grades.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "subject_grades" }

// DefaultRels mendefinisikan relasi domain ini.
// 3 BelongsTo autoload: student, subject, teacher.
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
		RelSubject: {
			Domain:     "subjects",
			Type:       vernon.RelBelongsTo,
			FK:         FieldSubjectID,
			LocalKey:   FieldSubjectID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
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

	subjectID, _ := data[FieldSubjectID].(string)
	if subjectID == "" {
		return fmt.Errorf("subject_id wajib diisi")
	}

	teacherID, _ := data[FieldTeacherID].(string)
	if teacherID == "" {
		return fmt.Errorf("teacher_id wajib diisi")
	}

	grade, ok := data[FieldGrade].(float64)
	if !ok {
		return fmt.Errorf("grade wajib diisi dan harus berupa angka")
	}
	if grade < 0 || grade > 100 {
		return fmt.Errorf("grade harus antara 0 dan 100")
	}

	gradeType, _ := data[FieldGradeType].(string)
	if gradeType != "" && !validGradeTypes[gradeType] {
		return fmt.Errorf("grade_type tidak valid: %q", gradeType)
	}
	return nil
}
