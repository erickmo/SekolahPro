package extracurricular_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/extracurricular"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// -- Descriptor metadata tests ---------------------------------------------------

func TestDescriptor_TableName(t *testing.T) {
	d := &extracurricular.Descriptor{}
	assert.Equal(t, "extracurriculars", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &extracurricular.Descriptor{}
}

// -- DefaultRels tests -----------------------------------------------------------

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &extracurricular.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "extracurricular should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &extracurricular.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[extracurricular.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, extracurricular.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func TestDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &extracurricular.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[extracurricular.RelTeacher]
	require.True(t, ok, "should have teacher relation")

	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, extracurricular.FieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload, "teacher should be autoloaded")
	assert.Equal(t, []string{"full_name", "nip"}, rel.Fields)
}

// -- Validation tests ------------------------------------------------------------

func validExtracurricularData() map[string]any {
	return map[string]any{
		extracurricular.FieldStudentID:    "00000000-0000-0000-0000-000000000001",
		extracurricular.FieldTeacherID:    "00000000-0000-0000-0000-000000000002",
		extracurricular.FieldActivityType: extracurricular.ActivityTypeSport,
		extracurricular.FieldStatus:       extracurricular.StatusActive,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &extracurricular.Descriptor{}
	err := d.Validate(validExtracurricularData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllActivityTypes(t *testing.T) {
	d := &extracurricular.Descriptor{}
	types := []string{
		extracurricular.ActivityTypeSport, extracurricular.ActivityTypeArt,
		extracurricular.ActivityTypeAcademic, extracurricular.ActivityTypeSocial,
		extracurricular.ActivityTypeReligion, extracurricular.ActivityTypeTechnology,
	}
	for _, at := range types {
		data := validExtracurricularData()
		data[extracurricular.FieldActivityType] = at
		err := d.Validate(data)
		assert.NoError(t, err, "activity_type=%q should be valid", at)
	}
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &extracurricular.Descriptor{}
	statuses := []string{
		extracurricular.StatusActive, extracurricular.StatusInactive,
		extracurricular.StatusCompleted,
	}
	for _, status := range statuses {
		data := validExtracurricularData()
		data[extracurricular.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &extracurricular.Descriptor{}
	data := validExtracurricularData()
	delete(data, extracurricular.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &extracurricular.Descriptor{}
	data := validExtracurricularData()
	delete(data, extracurricular.FieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestDescriptor_Validate_InvalidActivityType(t *testing.T) {
	d := &extracurricular.Descriptor{}
	data := validExtracurricularData()
	data[extracurricular.FieldActivityType] = "gaming"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "activity_type tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &extracurricular.Descriptor{}
	data := validExtracurricularData()
	data[extracurricular.FieldStatus] = "pending"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}
