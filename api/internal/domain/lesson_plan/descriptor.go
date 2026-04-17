// Package lesson_plan adalah domain Vernon untuk Rencana Pelaksanaan Pembelajaran (RPP).
//
// Lesson plan terdiri dari dua tabel:
//   - lesson_plans: data utama RPP/modul ajar
//   - lesson_plan_attachments: lampiran file RPP
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package lesson_plan

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Field name constants ─────────────────────────────────────────────────────

// LessonPlan field names.
const (
	FieldTeacherIDLP       = "teacher_id"
	FieldSubjectIDLP       = "subject_id"
	FieldClassRoomIDLP     = "class_room_id"
	FieldAcademicYearIDLP  = "academic_year_id"
	FieldTitle             = "title"
	FieldPlanType          = "plan_type"
	FieldSemesterLP        = "semester"
	FieldTopic             = "topic"
	FieldLearningObjectives = "learning_objectives"
	FieldLearningActivities = "learning_activities"
	FieldDurationMinutes   = "duration_minutes"
	FieldStatusLP          = "status"
	FieldTeachingMethodPesantrenLP = "teaching_method_pesantren"
)

// LessonPlanAttachment field names.
const (
	FieldLessonPlanID  = "lesson_plan_id"
	FieldFileName      = "file_name"
	FieldFilePath      = "file_path"
	FieldFileSizeBytes = "file_size_bytes"
	FieldMimeType      = "mime_type"
	FieldFileType      = "file_type"
)

// ── Enum constants ───────────────────────────────────────────────────────────

// Plan type constants.
const (
	PlanTypeModulAjar = "modul_ajar"
	PlanTypeRPPK13    = "rpp_k13"
	PlanTypeRPPKTSP   = "rpp_ktsp"
	PlanTypeRPPDiniyah = "rpp_diniyah"
)

// Status constants.
const (
	StatusDraft          = "draft"
	StatusSubmitted      = "submitted"
	StatusInReview       = "in_review"
	StatusRevisionNeeded = "revision_needed"
	StatusApproved       = "approved"
	StatusArchived       = "archived"
)

// Teaching method pesantren constants.
const (
	MethodBandongan    = "bandongan"
	MethodSorogan      = "sorogan"
	MethodHalaqah      = "halaqah"
	MethodMuhafadzah   = "muhafadzah"
	MethodMudzakarah   = "mudzakarah"
	MethodCeramah      = "ceramah"
	MethodDemonstrasi  = "demonstrasi"
	MethodTanyaJawab   = "tanya_jawab"
)

// File type constants.
const (
	FileTypeDocument      = "document"
	FileTypeImage         = "image"
	FileTypePresentation  = "presentation"
	FileTypeSpreadsheet   = "spreadsheet"
	FileTypeOther         = "other"
)

// Semester constants.
const (
	SemesterGanjilLP = "ganjil"
	SemesterGenapLP  = "genap"
)

// ── Range constants ──────────────────────────────────────────────────────────

const (
	DurationMinutesMin = 1
	DurationMinutesMax = 480
	FileSizeBytesMin   = 1
	FileSizeBytesMax   = 52428800 // 50 MB
)

// ── Relation name constants ──────────────────────────────────────────────────

const (
	RelTeacherLP       = "teacher"
	RelSubjectLP       = "subject"
	RelClassRoomLP     = "class_room"
	RelAcademicYearLP  = "academic_year"
	RelLessonPlan      = "lesson_plan"
)

// ── Validation maps ──────────────────────────────────────────────────────────

var validPlanTypes = map[string]bool{
	PlanTypeModulAjar: true, PlanTypeRPPK13: true,
	PlanTypeRPPKTSP: true, PlanTypeRPPDiniyah: true,
}

var validStatusesLP = map[string]bool{
	StatusDraft: true, StatusSubmitted: true, StatusInReview: true,
	StatusRevisionNeeded: true, StatusApproved: true, StatusArchived: true,
}

var validTeachingMethods = map[string]bool{
	MethodBandongan: true, MethodSorogan: true, MethodHalaqah: true,
	MethodMuhafadzah: true, MethodMudzakarah: true, MethodCeramah: true,
	MethodDemonstrasi: true, MethodTanyaJawab: true,
}

var validFileTypes = map[string]bool{
	FileTypeDocument: true, FileTypeImage: true, FileTypePresentation: true,
	FileTypeSpreadsheet: true, FileTypeOther: true,
}

var validMimeTypes = map[string]bool{
	"application/pdf": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"image/jpeg": true,
	"image/png": true,
	"image/webp": true,
}

var validSemestersLP = map[string]bool{
	SemesterGanjilLP: true, SemesterGenapLP: true,
}

// ── LessonPlanDescriptor ─────────────────────────────────────────────────────

// LessonPlanDescriptor mengimplementasi vernon.DomainDescriptor untuk lesson_plans.
type LessonPlanDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LessonPlanDescriptor) TableName() string { return "lesson_plans" }

// DefaultRels mendefinisikan relasi domain lesson_plans.
// 4 BelongsTo autoload: teacher, subject, class_room, academic_year.
func (d *LessonPlanDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacherLP: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTeacherIDLP,
			LocalKey:   FieldTeacherIDLP,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
		RelSubjectLP: {
			Domain:     "subjects",
			Type:       vernon.RelBelongsTo,
			FK:         FieldSubjectIDLP,
			LocalKey:   FieldSubjectIDLP,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
		RelClassRoomLP: {
			Domain:     "class_rooms",
			Type:       vernon.RelBelongsTo,
			FK:         FieldClassRoomIDLP,
			LocalKey:   FieldClassRoomIDLP,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "grade_level"},
		},
		RelAcademicYearLP: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearIDLP,
			LocalKey:   FieldAcademicYearIDLP,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant domain lesson_plans sebelum write.
func (d *LessonPlanDescriptor) Validate(data map[string]any) error {
	if err := validateLPRequired(data); err != nil {
		return err
	}
	return validateLPEnums(data)
}

// validateLPRequired memeriksa field wajib lesson_plans.
func validateLPRequired(data map[string]any) error {
	title, _ := data[FieldTitle].(string)
	if title == "" {
		return errors.New("title wajib diisi")
	}

	topic, _ := data[FieldTopic].(string)
	if topic == "" {
		return errors.New("topic wajib diisi")
	}

	objectives, _ := data[FieldLearningObjectives].(string)
	if objectives == "" {
		return errors.New("learning_objectives wajib diisi")
	}

	activities, _ := data[FieldLearningActivities].(string)
	if activities == "" {
		return errors.New("learning_activities wajib diisi")
	}

	semester, _ := data[FieldSemesterLP].(string)
	if semester == "" {
		return errors.New("semester wajib diisi")
	}
	return nil
}

// validateLPEnums memeriksa enum dan range constraints lesson_plans.
func validateLPEnums(data map[string]any) error {
	planType, _ := data[FieldPlanType].(string)
	if planType != "" && !validPlanTypes[planType] {
		return fmt.Errorf("plan_type tidak valid: %q", planType)
	}

	semester, _ := data[FieldSemesterLP].(string)
	if semester != "" && !validSemestersLP[semester] {
		return fmt.Errorf("semester tidak valid: %q (harus ganjil/genap)", semester)
	}

	status, _ := data[FieldStatusLP].(string)
	if status != "" && !validStatusesLP[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	method, _ := data[FieldTeachingMethodPesantrenLP].(string)
	if method != "" && !validTeachingMethods[method] {
		return fmt.Errorf("teaching_method_pesantren tidak valid: %q", method)
	}

	return validateLPDuration(data)
}

// validateLPDuration memeriksa duration_minutes 1-480.
func validateLPDuration(data map[string]any) error {
	duration, ok := data[FieldDurationMinutes].(float64)
	if !ok {
		return nil
	}
	if duration < DurationMinutesMin || duration > DurationMinutesMax {
		return fmt.Errorf("duration_minutes harus antara %d dan %d", DurationMinutesMin, DurationMinutesMax)
	}
	return nil
}

// ── LessonPlanAttachmentDescriptor ───────────────────────────────────────────

// LessonPlanAttachmentDescriptor mengimplementasi vernon.DomainDescriptor untuk lesson_plan_attachments.
type LessonPlanAttachmentDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LessonPlanAttachmentDescriptor) TableName() string { return "lesson_plan_attachments" }

// DefaultRels mendefinisikan relasi domain lesson_plan_attachments.
// belongs_to lesson_plan dengan autoload.
func (d *LessonPlanAttachmentDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelLessonPlan: {
			Domain:     "lesson_plans",
			Type:       vernon.RelBelongsTo,
			FK:         FieldLessonPlanID,
			LocalKey:   FieldLessonPlanID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"title", "plan_type", "status"},
		},
	}
}

// Validate memvalidasi invariant domain lesson_plan_attachments sebelum write.
func (d *LessonPlanAttachmentDescriptor) Validate(data map[string]any) error {
	if err := validateLPARequired(data); err != nil {
		return err
	}
	return validateLPAEnums(data)
}

// validateLPARequired memeriksa field wajib lesson_plan_attachments.
func validateLPARequired(data map[string]any) error {
	lessonPlanID, _ := data[FieldLessonPlanID].(string)
	if lessonPlanID == "" {
		return errors.New("lesson_plan_id wajib diisi")
	}

	fileName, _ := data[FieldFileName].(string)
	if fileName == "" {
		return errors.New("file_name wajib diisi")
	}

	filePath, _ := data[FieldFilePath].(string)
	if filePath == "" {
		return errors.New("file_path wajib diisi")
	}

	mimeType, _ := data[FieldMimeType].(string)
	if mimeType == "" {
		return errors.New("mime_type wajib diisi")
	}
	return nil
}

// validateLPAEnums memeriksa enum dan range constraints lesson_plan_attachments.
func validateLPAEnums(data map[string]any) error {
	mimeType, _ := data[FieldMimeType].(string)
	if mimeType != "" && !validMimeTypes[mimeType] {
		return fmt.Errorf("mime_type tidak valid: %q", mimeType)
	}

	fileType, _ := data[FieldFileType].(string)
	if fileType != "" && !validFileTypes[fileType] {
		return fmt.Errorf("file_type tidak valid: %q", fileType)
	}

	return validateLPAFileSize(data)
}

// validateLPAFileSize memeriksa file_size_bytes 1-52428800.
func validateLPAFileSize(data map[string]any) error {
	fileSize, ok := data[FieldFileSizeBytes].(float64)
	if !ok {
		return nil
	}
	if fileSize < FileSizeBytesMin || fileSize > FileSizeBytesMax {
		return fmt.Errorf("file_size_bytes harus antara %d dan %d", FileSizeBytesMin, FileSizeBytesMax)
	}
	return nil
}
