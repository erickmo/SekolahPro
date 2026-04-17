// Package rapor_generation adalah domain Vernon untuk generasi rapor siswa.
//
// Terdiri dari 2 sub-descriptor:
//   - TemplateDescriptor: template layout rapor (tanpa BelongsTo).
//   - RecordDescriptor: rapor per siswa per semester (4 BelongsTo: student, academic_year, class_room, template).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package rapor_generation

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// Template — Template layout rapor
// ============================================================================

// Template field constants.
const (
	TplFieldName           = "name"
	TplFieldCurriculumType = "curriculum_type"
	TplFieldGradeLevels    = "grade_levels"
	TplFieldDescription    = "description"
	TplFieldLayoutConfig   = "layout_config"
	TplFieldHeaderConfig   = "header_config"
	TplFieldIsActive       = "is_active"
)

// Curriculum type constants.
const (
	CurriculumMerdeka = "merdeka"
	CurriculumK13     = "k13"
	CurriculumDiniyah = "diniyah"
	CurriculumCustom  = "custom"
)

var validCurriculumTypes = map[string]bool{
	CurriculumMerdeka: true, CurriculumK13: true,
	CurriculumDiniyah: true, CurriculumCustom: true,
}

// TemplateDescriptor mengimplementasi vernon.DomainDescriptor untuk rapor_templates.
type TemplateDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *TemplateDescriptor) TableName() string { return "rapor_templates" }

// DefaultRels — rapor_templates tidak memiliki relasi (standalone config).
func (d *TemplateDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant rapor_templates.
func (d *TemplateDescriptor) Validate(data map[string]any) error {
	name, _ := data[TplFieldName].(string)
	if name == "" {
		return fmt.Errorf("name wajib diisi")
	}

	curriculumType, _ := data[TplFieldCurriculumType].(string)
	if curriculumType == "" {
		return fmt.Errorf("curriculum_type wajib diisi")
	}
	if !validCurriculumTypes[curriculumType] {
		return fmt.Errorf("curriculum_type tidak valid: %q", curriculumType)
	}

	gradeLevels := data[TplFieldGradeLevels]
	if gradeLevels == nil {
		return fmt.Errorf("grade_levels wajib diisi")
	}

	return nil
}

// ============================================================================
// Record — Rapor per siswa per semester
// ============================================================================

// Record field constants.
const (
	RecFieldStudentID            = "student_id"
	RecFieldAcademicYearID       = "academic_year_id"
	RecFieldClassRoomID          = "class_room_id"
	RecFieldTemplateID           = "template_id"
	RecFieldSemester             = "semester"
	RecFieldStatus               = "status"
	RecFieldStudentSnapshot      = "student_snapshot"
	RecFieldGradesSnapshot       = "grades_snapshot"
	RecFieldAttendanceSnapshot   = "attendance_snapshot"
	RecFieldPdfUrl               = "pdf_url"
	RecFieldGeneratedAt          = "generated_at"
	RecFieldFinalizedAt          = "finalized_at"
	RecFieldFinalizedBy          = "finalized_by"
	RecFieldNotes                = "notes"
)

// Semester constants.
const (
	SemesterGanjil = "ganjil"
	SemesterGenap  = "genap"
)

// Record status constants.
const (
	RecordStatusDraft     = "draft"
	RecordStatusGenerated = "generated"
	RecordStatusReviewed  = "reviewed"
	RecordStatusFinalized = "finalized"
)

// Record relation name constants.
const (
	RecRelStudent      = "student"
	RecRelAcademicYear = "academic_year"
	RecRelClassRoom    = "class_room"
	RecRelTemplate     = "template"
)

var validSemesters = map[string]bool{
	SemesterGanjil: true, SemesterGenap: true,
}

var validRecordStatuses = map[string]bool{
	RecordStatusDraft: true, RecordStatusGenerated: true,
	RecordStatusReviewed: true, RecordStatusFinalized: true,
}

// RecordDescriptor mengimplementasi vernon.DomainDescriptor untuk rapor_records.
type RecordDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *RecordDescriptor) TableName() string { return "rapor_records" }

// DefaultRels mendefinisikan 4 BelongsTo autoload: student, academic_year, class_room, template.
func (d *RecordDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RecRelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         RecFieldStudentID,
			LocalKey:   RecFieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
		RecRelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         RecFieldAcademicYearID,
			LocalKey:   RecFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
		RecRelClassRoom: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         RecFieldClassRoomID,
			LocalKey:   RecFieldClassRoomID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "grade_level"},
		},
		RecRelTemplate: {
			Domain:     "rapor_templates",
			Type:       vernon.RelBelongsTo,
			FK:         RecFieldTemplateID,
			LocalKey:   RecFieldTemplateID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "curriculum_type"},
		},
	}
}

// Validate memvalidasi invariant rapor_records.
func (d *RecordDescriptor) Validate(data map[string]any) error {
	if err := validateRecordRequired(data); err != nil {
		return err
	}
	return validateRecordEnums(data)
}

func validateRecordRequired(data map[string]any) error {
	studentID, _ := data[RecFieldStudentID].(string)
	if studentID == "" {
		return fmt.Errorf("student_id wajib diisi")
	}

	academicYearID, _ := data[RecFieldAcademicYearID].(string)
	if academicYearID == "" {
		return fmt.Errorf("academic_year_id wajib diisi")
	}

	classRoomID, _ := data[RecFieldClassRoomID].(string)
	if classRoomID == "" {
		return fmt.Errorf("class_room_id wajib diisi")
	}

	templateID, _ := data[RecFieldTemplateID].(string)
	if templateID == "" {
		return fmt.Errorf("template_id wajib diisi")
	}

	return nil
}

func validateRecordEnums(data map[string]any) error {
	semester, _ := data[RecFieldSemester].(string)
	if semester == "" {
		return fmt.Errorf("semester wajib diisi")
	}
	if !validSemesters[semester] {
		return fmt.Errorf("semester tidak valid: %q", semester)
	}

	status, _ := data[RecFieldStatus].(string)
	if status == "" {
		return fmt.Errorf("status wajib diisi")
	}
	if !validRecordStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	return nil
}
