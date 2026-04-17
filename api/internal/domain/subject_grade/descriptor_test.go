package subject_grade_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/subject_grade"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// -- Descriptor metadata tests ---------------------------------------------------

func TestDescriptor_TableName(t *testing.T) {
	d := &subject_grade.Descriptor{}
	assert.Equal(t, "subject_grades", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &subject_grade.Descriptor{}
}

// -- DefaultRels tests -----------------------------------------------------------

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &subject_grade.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 3, "subject_grade should have 3 BelongsTo relations")
}

func TestDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &subject_grade.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[subject_grade.RelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, subject_grade.FieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload, "student should be autoloaded")
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func TestDescriptor_DefaultRels_SubjectRelation(t *testing.T) {
	d := &subject_grade.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[subject_grade.RelSubject]
	require.True(t, ok, "should have subject relation")

	assert.Equal(t, "subjects", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, subject_grade.FieldSubjectID, rel.FK)
	assert.True(t, rel.IsAutoload, "subject should be autoloaded")
	assert.Equal(t, []string{"name", "code"}, rel.Fields)
}

func TestDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &subject_grade.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[subject_grade.RelTeacher]
	require.True(t, ok, "should have teacher relation")

	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, subject_grade.FieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload, "teacher should be autoloaded")
	assert.Equal(t, []string{"full_name", "nip"}, rel.Fields)
}

// -- Validation tests ------------------------------------------------------------

func validSubjectGradeData() map[string]any {
	return map[string]any{
		subject_grade.FieldStudentID: "00000000-0000-0000-0000-000000000001",
		subject_grade.FieldSubjectID: "00000000-0000-0000-0000-000000000002",
		subject_grade.FieldTeacherID: "00000000-0000-0000-0000-000000000003",
		subject_grade.FieldGrade:     float64(85),
		subject_grade.FieldGradeType: subject_grade.GradeTypeAssignment,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &subject_grade.Descriptor{}
	err := d.Validate(validSubjectGradeData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_MissingStudentID_ReturnsError(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	delete(data, subject_grade.FieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestDescriptor_Validate_MissingSubjectID_ReturnsError(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	delete(data, subject_grade.FieldSubjectID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "subject_id wajib diisi")
}

func TestDescriptor_Validate_MissingTeacherID_ReturnsError(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	delete(data, subject_grade.FieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestDescriptor_Validate_MissingGrade_ReturnsError(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	delete(data, subject_grade.FieldGrade)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "grade wajib diisi dan harus berupa angka")
}

func TestDescriptor_Validate_GradeBelowZero_ReturnsError(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	data[subject_grade.FieldGrade] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "grade harus antara 0 dan 100")
}

func TestDescriptor_Validate_GradeAbove100_ReturnsError(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	data[subject_grade.FieldGrade] = float64(101)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "grade harus antara 0 dan 100")
}

func TestDescriptor_Validate_GradeBoundaryZero_OK(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	data[subject_grade.FieldGrade] = float64(0)
	err := d.Validate(data)
	assert.NoError(t, err, "grade 0 should be valid")
}

func TestDescriptor_Validate_GradeBoundary100_OK(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	data[subject_grade.FieldGrade] = float64(100)
	err := d.Validate(data)
	assert.NoError(t, err, "grade 100 should be valid")
}

func TestDescriptor_Validate_InvalidGradeType_ReturnsError(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	data[subject_grade.FieldGradeType] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "grade_type tidak valid")
}

func TestDescriptor_Validate_AllGradeTypes(t *testing.T) {
	d := &subject_grade.Descriptor{}
	gradeTypes := []string{
		subject_grade.GradeTypeAssignment, subject_grade.GradeTypeMidterm,
		subject_grade.GradeTypeFinal, subject_grade.GradeTypeQuiz,
		subject_grade.GradeTypeProject,
	}
	for _, gt := range gradeTypes {
		data := validSubjectGradeData()
		data[subject_grade.FieldGradeType] = gt
		err := d.Validate(data)
		assert.NoError(t, err, "grade_type=%q should be valid", gt)
	}
}

func TestDescriptor_Validate_EmptyGradeType_OK(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	data[subject_grade.FieldGradeType] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty grade_type should be valid (optional)")
}

func TestDescriptor_Validate_GradeNotNumber_ReturnsError(t *testing.T) {
	d := &subject_grade.Descriptor{}
	data := validSubjectGradeData()
	data[subject_grade.FieldGrade] = "not-a-number"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "grade wajib diisi dan harus berupa angka")
}
