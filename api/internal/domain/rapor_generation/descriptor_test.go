package rapor_generation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/rapor_generation"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// TemplateDescriptor tests
// ============================================================================

func TestTemplateDescriptor_TableName(t *testing.T) {
	d := &rapor_generation.TemplateDescriptor{}
	assert.Equal(t, "rapor_templates", d.TableName())
}

func TestTemplateDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &rapor_generation.TemplateDescriptor{}
}

func TestTemplateDescriptor_DefaultRels(t *testing.T) {
	d := &rapor_generation.TemplateDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 0, "rapor_templates should have no relations")
}

func validTemplateData() map[string]any {
	return map[string]any{
		rapor_generation.TplFieldName:           "Rapor Kurikulum Merdeka",
		rapor_generation.TplFieldCurriculumType: rapor_generation.CurriculumMerdeka,
		rapor_generation.TplFieldGradeLevels:    []any{"7", "8", "9"},
	}
}

func TestTemplateDescriptor_Validate_Success(t *testing.T) {
	d := &rapor_generation.TemplateDescriptor{}
	err := d.Validate(validTemplateData())
	assert.NoError(t, err)
}

func TestTemplateDescriptor_Validate_AllCurriculumTypes(t *testing.T) {
	d := &rapor_generation.TemplateDescriptor{}
	types := []string{
		rapor_generation.CurriculumMerdeka,
		rapor_generation.CurriculumK13,
		rapor_generation.CurriculumDiniyah,
		rapor_generation.CurriculumCustom,
	}
	for _, ct := range types {
		data := validTemplateData()
		data[rapor_generation.TplFieldCurriculumType] = ct
		err := d.Validate(data)
		assert.NoError(t, err, "curriculum_type=%q should be valid", ct)
	}
}

func TestTemplateDescriptor_Validate_MissingName(t *testing.T) {
	d := &rapor_generation.TemplateDescriptor{}
	data := validTemplateData()
	delete(data, rapor_generation.TplFieldName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "name wajib diisi")
}

func TestTemplateDescriptor_Validate_MissingCurriculumType(t *testing.T) {
	d := &rapor_generation.TemplateDescriptor{}
	data := validTemplateData()
	delete(data, rapor_generation.TplFieldCurriculumType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "curriculum_type wajib diisi")
}

func TestTemplateDescriptor_Validate_InvalidCurriculumType(t *testing.T) {
	d := &rapor_generation.TemplateDescriptor{}
	data := validTemplateData()
	data[rapor_generation.TplFieldCurriculumType] = "cbsa"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "curriculum_type tidak valid")
}

func TestTemplateDescriptor_Validate_MissingGradeLevels(t *testing.T) {
	d := &rapor_generation.TemplateDescriptor{}
	data := validTemplateData()
	delete(data, rapor_generation.TplFieldGradeLevels)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "grade_levels wajib diisi")
}

// ============================================================================
// RecordDescriptor tests
// ============================================================================

func TestRecordDescriptor_TableName(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	assert.Equal(t, "rapor_records", d.TableName())
}

func TestRecordDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &rapor_generation.RecordDescriptor{}
}

func TestRecordDescriptor_DefaultRels_Count(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 4, "rapor_records should have 4 BelongsTo relations")
}

func TestRecordDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[rapor_generation.RecRelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, rapor_generation.RecFieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func TestRecordDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[rapor_generation.RecRelAcademicYear]
	require.True(t, ok, "should have academic_year relation")

	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, rapor_generation.RecFieldAcademicYearID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"name", "code"}, rel.Fields)
}

func TestRecordDescriptor_DefaultRels_ClassRoomRelation(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[rapor_generation.RecRelClassRoom]
	require.True(t, ok, "should have class_room relation")

	assert.Equal(t, "class_rooms", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, rapor_generation.RecFieldClassRoomID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"name", "grade_level"}, rel.Fields)
}

func TestRecordDescriptor_DefaultRels_TemplateRelation(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[rapor_generation.RecRelTemplate]
	require.True(t, ok, "should have template relation")

	assert.Equal(t, "rapor_templates", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, rapor_generation.RecFieldTemplateID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"name", "curriculum_type"}, rel.Fields)
}

func validRecordData() map[string]any {
	return map[string]any{
		rapor_generation.RecFieldStudentID:      "00000000-0000-0000-0000-000000000001",
		rapor_generation.RecFieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		rapor_generation.RecFieldClassRoomID:    "00000000-0000-0000-0000-000000000003",
		rapor_generation.RecFieldTemplateID:     "00000000-0000-0000-0000-000000000004",
		rapor_generation.RecFieldSemester:       rapor_generation.SemesterGanjil,
		rapor_generation.RecFieldStatus:         rapor_generation.RecordStatusDraft,
	}
}

func TestRecordDescriptor_Validate_Success(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	err := d.Validate(validRecordData())
	assert.NoError(t, err)
}

func TestRecordDescriptor_Validate_AllSemesters(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	semesters := []string{
		rapor_generation.SemesterGanjil,
		rapor_generation.SemesterGenap,
	}
	for _, sem := range semesters {
		data := validRecordData()
		data[rapor_generation.RecFieldSemester] = sem
		err := d.Validate(data)
		assert.NoError(t, err, "semester=%q should be valid", sem)
	}
}

func TestRecordDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	statuses := []string{
		rapor_generation.RecordStatusDraft,
		rapor_generation.RecordStatusGenerated,
		rapor_generation.RecordStatusReviewed,
		rapor_generation.RecordStatusFinalized,
	}
	for _, status := range statuses {
		data := validRecordData()
		data[rapor_generation.RecFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestRecordDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	data := validRecordData()
	delete(data, rapor_generation.RecFieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestRecordDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	data := validRecordData()
	delete(data, rapor_generation.RecFieldAcademicYearID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestRecordDescriptor_Validate_MissingClassRoomID(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	data := validRecordData()
	delete(data, rapor_generation.RecFieldClassRoomID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "class_room_id wajib diisi")
}

func TestRecordDescriptor_Validate_MissingTemplateID(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	data := validRecordData()
	delete(data, rapor_generation.RecFieldTemplateID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "template_id wajib diisi")
}

func TestRecordDescriptor_Validate_MissingSemester(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	data := validRecordData()
	delete(data, rapor_generation.RecFieldSemester)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester wajib diisi")
}

func TestRecordDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	data := validRecordData()
	data[rapor_generation.RecFieldSemester] = "midterm"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester tidak valid")
}

func TestRecordDescriptor_Validate_MissingStatus(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	data := validRecordData()
	delete(data, rapor_generation.RecFieldStatus)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status wajib diisi")
}

func TestRecordDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &rapor_generation.RecordDescriptor{}
	data := validRecordData()
	data[rapor_generation.RecFieldStatus] = "published"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}
