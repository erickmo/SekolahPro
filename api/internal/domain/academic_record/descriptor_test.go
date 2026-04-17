package academic_record_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/academic_record"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &academic_record.Descriptor{}
	assert.Equal(t, "academic_records", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &academic_record.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &academic_record.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "academic_record should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &academic_record.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[academic_record.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, academic_record.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func TestDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &academic_record.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[academic_record.RelAcademicYear]
	require.True(t, ok, "should have academic_year relation")

	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, academic_record.FieldAcademicYearID, rel.FK)
	assert.True(t, rel.IsAutoload, "academic_year should be autoloaded")
	assert.Equal(t, []string{"name", "code"}, rel.Fields)
}

// ── Validation tests ────────────────────────────────────────────────────────

func validAcademicRecordData() map[string]any {
	return map[string]any{
		academic_record.FieldStudentID:      "00000000-0000-0000-0000-000000000001",
		academic_record.FieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		academic_record.FieldSemester:       "1",
		academic_record.FieldGPA:            float64(3.5),
		academic_record.FieldRank:           float64(5),
		academic_record.FieldTotalScore:     float64(87.5),
		academic_record.FieldStatus:         academic_record.StatusPass,
		academic_record.FieldNotes:          "Catatan semester 1",
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &academic_record.Descriptor{}
	err := d.Validate(validAcademicRecordData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &academic_record.Descriptor{}
	statuses := []string{
		academic_record.StatusPass, academic_record.StatusFail, academic_record.StatusRemedial,
	}
	for _, status := range statuses {
		data := validAcademicRecordData()
		data[academic_record.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &academic_record.Descriptor{}
	data := validAcademicRecordData()
	delete(data, academic_record.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &academic_record.Descriptor{}
	data := validAcademicRecordData()
	delete(data, academic_record.FieldAcademicYearID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &academic_record.Descriptor{}
	data := validAcademicRecordData()
	data[academic_record.FieldSemester] = "3"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "semester tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &academic_record.Descriptor{}
	data := validAcademicRecordData()
	data[academic_record.FieldStatus] = "pending"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_EmptySemester_OK(t *testing.T) {
	d := &academic_record.Descriptor{}
	data := validAcademicRecordData()
	data[academic_record.FieldSemester] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty semester should be valid (optional)")
}

func TestDescriptor_Validate_EmptyStatus_OK(t *testing.T) {
	d := &academic_record.Descriptor{}
	data := validAcademicRecordData()
	data[academic_record.FieldStatus] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty status should be valid (optional)")
}
