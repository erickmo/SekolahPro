package lesson_plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/lesson_plan"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── LessonPlanDescriptor metadata tests ──────────────────────────────────────

func TestLessonPlanDescriptor_TableName(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	assert.Equal(t, "lesson_plans", d.TableName())
}

func TestLessonPlanDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &lesson_plan.LessonPlanDescriptor{}
}

func TestLessonPlanDescriptor_DefaultRels_Count(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 4, "lesson_plans should have 4 BelongsTo relations")
}

func TestLessonPlanDescriptor_DefaultRels_AllAutoload(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	rels := d.DefaultRels()

	expected := []string{
		lesson_plan.RelTeacherLP, lesson_plan.RelSubjectLP,
		lesson_plan.RelClassRoomLP, lesson_plan.RelAcademicYearLP,
	}
	for _, name := range expected {
		rel, ok := rels[name]
		require.True(t, ok, "should have %s relation", name)
		assert.Equal(t, vernon.RelBelongsTo, rel.Type, "%s should be belongs_to", name)
		assert.True(t, rel.IsAutoload, "%s should be autoloaded", name)
	}
}

// ── LessonPlanDescriptor validation tests ────────────────────────────────────

func validLessonPlanData() map[string]any {
	return map[string]any{
		lesson_plan.FieldTitle:              "RPP Matematika Bab 1",
		lesson_plan.FieldPlanType:           lesson_plan.PlanTypeRPPK13,
		lesson_plan.FieldSemesterLP:         lesson_plan.SemesterGanjilLP,
		lesson_plan.FieldTopic:              "Bilangan Bulat",
		lesson_plan.FieldLearningObjectives: "Siswa mampu memahami bilangan bulat",
		lesson_plan.FieldLearningActivities: "Ceramah, diskusi, latihan soal",
		lesson_plan.FieldDurationMinutes:    float64(90),
		lesson_plan.FieldStatusLP:           lesson_plan.StatusDraft,
		lesson_plan.FieldTeacherIDLP:        "00000000-0000-0000-0000-000000000001",
		lesson_plan.FieldSubjectIDLP:        "00000000-0000-0000-0000-000000000002",
		lesson_plan.FieldClassRoomIDLP:      "00000000-0000-0000-0000-000000000003",
		lesson_plan.FieldAcademicYearIDLP:   "00000000-0000-0000-0000-000000000004",
	}
}

func TestLessonPlanDescriptor_Validate_Success(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	err := d.Validate(validLessonPlanData())
	assert.NoError(t, err)
}

func TestLessonPlanDescriptor_Validate_AllPlanTypes(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	types := []string{
		lesson_plan.PlanTypeModulAjar, lesson_plan.PlanTypeRPPK13,
		lesson_plan.PlanTypeRPPKTSP, lesson_plan.PlanTypeRPPDiniyah,
	}
	for _, pt := range types {
		data := validLessonPlanData()
		data[lesson_plan.FieldPlanType] = pt
		err := d.Validate(data)
		assert.NoError(t, err, "plan_type=%q should be valid", pt)
	}
}

func TestLessonPlanDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	statuses := []string{
		lesson_plan.StatusDraft, lesson_plan.StatusSubmitted,
		lesson_plan.StatusInReview, lesson_plan.StatusRevisionNeeded,
		lesson_plan.StatusApproved, lesson_plan.StatusArchived,
	}
	for _, status := range statuses {
		data := validLessonPlanData()
		data[lesson_plan.FieldStatusLP] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestLessonPlanDescriptor_Validate_AllTeachingMethods(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	methods := []string{
		lesson_plan.MethodBandongan, lesson_plan.MethodSorogan,
		lesson_plan.MethodHalaqah, lesson_plan.MethodMuhafadzah,
		lesson_plan.MethodMudzakarah, lesson_plan.MethodCeramah,
		lesson_plan.MethodDemonstrasi, lesson_plan.MethodTanyaJawab,
	}
	for _, method := range methods {
		data := validLessonPlanData()
		data[lesson_plan.FieldTeachingMethodPesantrenLP] = method
		err := d.Validate(data)
		assert.NoError(t, err, "teaching_method_pesantren=%q should be valid", method)
	}
}

func TestLessonPlanDescriptor_Validate_MissingTitle(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	delete(data, lesson_plan.FieldTitle)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "title wajib diisi")
}

func TestLessonPlanDescriptor_Validate_MissingTopic(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	delete(data, lesson_plan.FieldTopic)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "topic wajib diisi")
}

func TestLessonPlanDescriptor_Validate_MissingLearningObjectives(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	delete(data, lesson_plan.FieldLearningObjectives)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "learning_objectives wajib diisi")
}

func TestLessonPlanDescriptor_Validate_MissingLearningActivities(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	delete(data, lesson_plan.FieldLearningActivities)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "learning_activities wajib diisi")
}

func TestLessonPlanDescriptor_Validate_MissingSemester(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	delete(data, lesson_plan.FieldSemesterLP)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester wajib diisi")
}

func TestLessonPlanDescriptor_Validate_InvalidPlanType(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	data[lesson_plan.FieldPlanType] = "invalid_type"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "plan_type tidak valid")
}

func TestLessonPlanDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	data[lesson_plan.FieldSemesterLP] = "summer"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester tidak valid")
}

func TestLessonPlanDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	data[lesson_plan.FieldStatusLP] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestLessonPlanDescriptor_Validate_InvalidTeachingMethod(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	data[lesson_plan.FieldTeachingMethodPesantrenLP] = "invalid_method"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teaching_method_pesantren tidak valid")
}

func TestLessonPlanDescriptor_Validate_DurationTooHigh(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	data[lesson_plan.FieldDurationMinutes] = float64(500)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "duration_minutes harus antara")
}

func TestLessonPlanDescriptor_Validate_DurationZero(t *testing.T) {
	d := &lesson_plan.LessonPlanDescriptor{}
	data := validLessonPlanData()
	data[lesson_plan.FieldDurationMinutes] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "duration_minutes harus antara")
}

// ── LessonPlanAttachmentDescriptor metadata tests ────────────────────────────

func TestLessonPlanAttachmentDescriptor_TableName(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	assert.Equal(t, "lesson_plan_attachments", d.TableName())
}

func TestLessonPlanAttachmentDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &lesson_plan.LessonPlanAttachmentDescriptor{}
}

func TestLessonPlanAttachmentDescriptor_DefaultRels_Count(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "lesson_plan_attachments should have 1 BelongsTo relation")
}

func TestLessonPlanAttachmentDescriptor_DefaultRels_LessonPlan(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[lesson_plan.RelLessonPlan]
	require.True(t, ok, "should have lesson_plan relation")

	assert.Equal(t, "lesson_plans", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "lesson_plan should be autoloaded")
}

// ── LessonPlanAttachmentDescriptor validation tests ──────────────────────────

func validLessonPlanAttachmentData() map[string]any {
	return map[string]any{
		lesson_plan.FieldLessonPlanID:  "00000000-0000-0000-0000-000000000001",
		lesson_plan.FieldFileName:      "rpp_matematika.pdf",
		lesson_plan.FieldFilePath:      "/uploads/rpp/rpp_matematika.pdf",
		lesson_plan.FieldFileSizeBytes: float64(1048576),
		lesson_plan.FieldMimeType:      "application/pdf",
		lesson_plan.FieldFileType:      lesson_plan.FileTypeDocument,
	}
}

func TestLessonPlanAttachmentDescriptor_Validate_Success(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	err := d.Validate(validLessonPlanAttachmentData())
	assert.NoError(t, err)
}

func TestLessonPlanAttachmentDescriptor_Validate_AllMimeTypes(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	mimeTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"image/jpeg",
		"image/png",
		"image/webp",
	}
	for _, mt := range mimeTypes {
		data := validLessonPlanAttachmentData()
		data[lesson_plan.FieldMimeType] = mt
		err := d.Validate(data)
		assert.NoError(t, err, "mime_type=%q should be valid", mt)
	}
}

func TestLessonPlanAttachmentDescriptor_Validate_AllFileTypes(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	fileTypes := []string{
		lesson_plan.FileTypeDocument, lesson_plan.FileTypeImage,
		lesson_plan.FileTypePresentation, lesson_plan.FileTypeSpreadsheet,
		lesson_plan.FileTypeOther,
	}
	for _, ft := range fileTypes {
		data := validLessonPlanAttachmentData()
		data[lesson_plan.FieldFileType] = ft
		err := d.Validate(data)
		assert.NoError(t, err, "file_type=%q should be valid", ft)
	}
}

func TestLessonPlanAttachmentDescriptor_Validate_MissingLessonPlanID(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	data := validLessonPlanAttachmentData()
	delete(data, lesson_plan.FieldLessonPlanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "lesson_plan_id wajib diisi")
}

func TestLessonPlanAttachmentDescriptor_Validate_MissingFileName(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	data := validLessonPlanAttachmentData()
	delete(data, lesson_plan.FieldFileName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_name wajib diisi")
}

func TestLessonPlanAttachmentDescriptor_Validate_MissingFilePath(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	data := validLessonPlanAttachmentData()
	delete(data, lesson_plan.FieldFilePath)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_path wajib diisi")
}

func TestLessonPlanAttachmentDescriptor_Validate_MissingMimeType(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	data := validLessonPlanAttachmentData()
	delete(data, lesson_plan.FieldMimeType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "mime_type wajib diisi")
}

func TestLessonPlanAttachmentDescriptor_Validate_InvalidMimeType(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	data := validLessonPlanAttachmentData()
	data[lesson_plan.FieldMimeType] = "application/exe"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "mime_type tidak valid")
}

func TestLessonPlanAttachmentDescriptor_Validate_InvalidFileType(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	data := validLessonPlanAttachmentData()
	data[lesson_plan.FieldFileType] = "video"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_type tidak valid")
}

func TestLessonPlanAttachmentDescriptor_Validate_FileSizeTooLarge(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	data := validLessonPlanAttachmentData()
	data[lesson_plan.FieldFileSizeBytes] = float64(60000000)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_size_bytes harus antara")
}

func TestLessonPlanAttachmentDescriptor_Validate_FileSizeZero(t *testing.T) {
	d := &lesson_plan.LessonPlanAttachmentDescriptor{}
	data := validLessonPlanAttachmentData()
	data[lesson_plan.FieldFileSizeBytes] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "file_size_bytes harus antara")
}
