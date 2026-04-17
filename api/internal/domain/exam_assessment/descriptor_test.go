package exam_assessment_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/exam_assessment"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// -- Descriptor metadata tests ---------------------------------------------------

func TestDescriptor_TableName(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	assert.Equal(t, "exam_assessments", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &exam_assessment.Descriptor{}
}

// -- DefaultRels tests -----------------------------------------------------------

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 3, "exam_assessment should have 3 BelongsTo relations")
}

func TestDescriptor_DefaultRels_SubjectRelation(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[exam_assessment.RelSubject]
	require.True(t, ok, "should have subject relation")

	assert.Equal(t, "subjects", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, exam_assessment.FieldSubjectID, rel.FK)
	assert.True(t, rel.IsAutoload, "subject should be autoloaded")
	assert.Equal(t, []string{"name", "code"}, rel.Fields)
}

func TestDescriptor_DefaultRels_TeacherRelation(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[exam_assessment.RelTeacher]
	require.True(t, ok, "should have teacher relation")

	assert.Equal(t, "teachers", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, exam_assessment.FieldTeacherID, rel.FK)
	assert.True(t, rel.IsAutoload, "teacher should be autoloaded")
	assert.Equal(t, []string{"full_name", "nip"}, rel.Fields)
}

func TestDescriptor_DefaultRels_ClassRoomRelation(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[exam_assessment.RelClassRoom]
	require.True(t, ok, "should have class_room relation")

	assert.Equal(t, "class_rooms", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, exam_assessment.FieldClassRoomID, rel.FK)
	assert.True(t, rel.IsAutoload, "class_room should be autoloaded")
	assert.Equal(t, []string{"name", "grade_level"}, rel.Fields)
}

// -- Validation tests ------------------------------------------------------------

func validExamAssessmentData() map[string]any {
	return map[string]any{
		exam_assessment.FieldSubjectID:   "00000000-0000-0000-0000-000000000001",
		exam_assessment.FieldTeacherID:   "00000000-0000-0000-0000-000000000002",
		exam_assessment.FieldClassRoomID: "00000000-0000-0000-0000-000000000003",
		exam_assessment.FieldTitle:       "UTS Matematika Semester 1",
		exam_assessment.FieldExamType:    exam_assessment.ExamTypeUTS,
		exam_assessment.FieldDate:        "2026-05-17",
		exam_assessment.FieldMaxScore:    float64(100),
		exam_assessment.FieldStatus:      exam_assessment.StatusDraft,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	err := d.Validate(validExamAssessmentData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_MissingSubjectID_ReturnsError(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	data := validExamAssessmentData()
	delete(data, exam_assessment.FieldSubjectID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "subject_id wajib diisi")
}

func TestDescriptor_Validate_MissingTeacherID_ReturnsError(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	data := validExamAssessmentData()
	delete(data, exam_assessment.FieldTeacherID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teacher_id wajib diisi")
}

func TestDescriptor_Validate_MissingClassRoomID_ReturnsError(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	data := validExamAssessmentData()
	delete(data, exam_assessment.FieldClassRoomID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "class_room_id wajib diisi")
}

func TestDescriptor_Validate_MissingTitle_ReturnsError(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	data := validExamAssessmentData()
	delete(data, exam_assessment.FieldTitle)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "title wajib diisi")
}

func TestDescriptor_Validate_InvalidExamType_ReturnsError(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	data := validExamAssessmentData()
	data[exam_assessment.FieldExamType] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "exam_type tidak valid")
}

func TestDescriptor_Validate_InvalidStatus_ReturnsError(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	data := validExamAssessmentData()
	data[exam_assessment.FieldStatus] = "cancelled"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_AllExamTypes(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	examTypes := []string{
		exam_assessment.ExamTypeUTS, exam_assessment.ExamTypeUAS,
		exam_assessment.ExamTypeQuiz, exam_assessment.ExamTypeAssignment,
		exam_assessment.ExamTypeProject, exam_assessment.ExamTypePractice,
	}
	for _, et := range examTypes {
		data := validExamAssessmentData()
		data[exam_assessment.FieldExamType] = et
		err := d.Validate(data)
		assert.NoError(t, err, "exam_type=%q should be valid", et)
	}
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	statuses := []string{
		exam_assessment.StatusDraft, exam_assessment.StatusPublished,
		exam_assessment.StatusCompleted,
	}
	for _, s := range statuses {
		data := validExamAssessmentData()
		data[exam_assessment.FieldStatus] = s
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", s)
	}
}

func TestDescriptor_Validate_EmptyExamType_OK(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	data := validExamAssessmentData()
	data[exam_assessment.FieldExamType] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty exam_type should be valid (optional)")
}

func TestDescriptor_Validate_EmptyStatus_OK(t *testing.T) {
	d := &exam_assessment.Descriptor{}
	data := validExamAssessmentData()
	data[exam_assessment.FieldStatus] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty status should be valid (optional)")
}
