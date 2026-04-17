// Package exam_assessment adalah domain Vernon untuk ujian dan penilaian.
//
// ExamAssessment memiliki 3 BelongsTo autoload: subject, teacher, dan class_room.
// Digunakan untuk tracking ujian, kuis, dan penilaian per kelas.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package exam_assessment

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldSubjectID   = "subject_id"
	FieldTeacherID   = "teacher_id"
	FieldClassRoomID = "class_room_id"
	FieldTitle       = "title"
	FieldExamType    = "exam_type"
	FieldDate        = "date"
	FieldMaxScore    = "max_score"
	FieldDescription = "description"
	FieldStatus      = "status"
)

// Exam type constants.
const (
	ExamTypeUTS       = "uts"
	ExamTypeUAS       = "uas"
	ExamTypeQuiz      = "quiz"
	ExamTypeAssignment = "assignment"
	ExamTypeProject   = "project"
	ExamTypePractice  = "practice"
)

// Status constants.
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusCompleted = "completed"
)

// Relation name constants.
const (
	RelSubject   = "subject"
	RelTeacher   = "teacher"
	RelClassRoom = "class_room"
)

var validExamTypes = map[string]bool{
	ExamTypeUTS: true, ExamTypeUAS: true, ExamTypeQuiz: true,
	ExamTypeAssignment: true, ExamTypeProject: true, ExamTypePractice: true,
}

var validStatuses = map[string]bool{
	StatusDraft: true, StatusPublished: true, StatusCompleted: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk exam_assessments.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "exam_assessments" }

// DefaultRels mendefinisikan relasi domain ini.
// 3 BelongsTo autoload: subject, teacher, class_room.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
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
		RelClassRoom: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         FieldClassRoomID,
			LocalKey:   FieldClassRoomID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "grade_level"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	subjectID, _ := data[FieldSubjectID].(string)
	if subjectID == "" {
		return fmt.Errorf("subject_id wajib diisi")
	}

	teacherID, _ := data[FieldTeacherID].(string)
	if teacherID == "" {
		return fmt.Errorf("teacher_id wajib diisi")
	}

	classRoomID, _ := data[FieldClassRoomID].(string)
	if classRoomID == "" {
		return fmt.Errorf("class_room_id wajib diisi")
	}

	title, _ := data[FieldTitle].(string)
	if title == "" {
		return fmt.Errorf("title wajib diisi")
	}

	examType, _ := data[FieldExamType].(string)
	if examType != "" && !validExamTypes[examType] {
		return fmt.Errorf("exam_type tidak valid: %q", examType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
