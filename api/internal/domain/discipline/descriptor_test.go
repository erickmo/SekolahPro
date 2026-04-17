package discipline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/discipline"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// -- Descriptor metadata tests ---------------------------------------------------

func TestDescriptor_TableName(t *testing.T) {
	d := &discipline.Descriptor{}
	assert.Equal(t, "disciplines", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &discipline.Descriptor{}
}

// -- DefaultRels tests -----------------------------------------------------------

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &discipline.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "discipline should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &discipline.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[discipline.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, discipline.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func TestDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &discipline.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[discipline.RelTeacher]
	require.True(t, ok, "should have teacher relation")

	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, discipline.FieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload, "teacher should be autoloaded")
	assert.Equal(t, []string{"full_name", "nip"}, rel.Fields)
}

// -- Validation tests ------------------------------------------------------------

func validDisciplineData() map[string]any {
	return map[string]any{
		discipline.FieldStudentID:     "00000000-0000-0000-0000-000000000001",
		discipline.FieldTeacherID:     "00000000-0000-0000-0000-000000000002",
		discipline.FieldViolationType: discipline.ViolationMinor,
		discipline.FieldStatus:        discipline.StatusReported,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &discipline.Descriptor{}
	err := d.Validate(validDisciplineData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllViolationTypes(t *testing.T) {
	d := &discipline.Descriptor{}
	types := []string{
		discipline.ViolationMinor, discipline.ViolationModerate, discipline.ViolationMajor,
	}
	for _, vt := range types {
		data := validDisciplineData()
		data[discipline.FieldViolationType] = vt
		err := d.Validate(data)
		assert.NoError(t, err, "violation_type=%q should be valid", vt)
	}
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &discipline.Descriptor{}
	statuses := []string{
		discipline.StatusReported, discipline.StatusReviewed,
		discipline.StatusSanctioned, discipline.StatusResolved,
	}
	for _, status := range statuses {
		data := validDisciplineData()
		data[discipline.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &discipline.Descriptor{}
	data := validDisciplineData()
	delete(data, discipline.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := &discipline.Descriptor{}
	data := validDisciplineData()
	delete(data, discipline.FieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestDescriptor_Validate_InvalidViolationType(t *testing.T) {
	d := &discipline.Descriptor{}
	data := validDisciplineData()
	data[discipline.FieldViolationType] = "criminal"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "violation_type tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &discipline.Descriptor{}
	data := validDisciplineData()
	data[discipline.FieldStatus] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}
