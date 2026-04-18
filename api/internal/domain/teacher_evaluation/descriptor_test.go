package teacher_evaluation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/teacher_evaluation"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

func TestCompetencyDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "teacher_evaluation_competencies", (&teacher_evaluation.CompetencyDescriptor{}).TableName())
}
func TestCompetencyDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_evaluation.CompetencyDescriptor{}
}
func TestCompetencyDescriptor_DefaultRels(t *testing.T) {
	assert.Len(t, (&teacher_evaluation.CompetencyDescriptor{}).DefaultRels(), 0)
}

func validCompetencyData() map[string]any {
	return map[string]any{
		teacher_evaluation.CompFieldName: "Pedagogik", teacher_evaluation.CompFieldArea: teacher_evaluation.AreaPedagogic,
	}
}
func TestCompetencyDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&teacher_evaluation.CompetencyDescriptor{}).Validate(validCompetencyData()))
}
func TestCompetencyDescriptor_Validate_InvalidArea(t *testing.T) {
	d := validCompetencyData(); d[teacher_evaluation.CompFieldArea] = "tech"
	assert.ErrorContains(t, (&teacher_evaluation.CompetencyDescriptor{}).Validate(d), "area tidak valid")
}

func TestEvaluationDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "teacher_evaluations", (&teacher_evaluation.EvaluationDescriptor{}).TableName())
}
func TestEvaluationDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_evaluation.EvaluationDescriptor{}
}
func TestEvaluationDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&teacher_evaluation.EvaluationDescriptor{}).DefaultRels(), 2)
}
func TestEvaluationDescriptor_DefaultRels_Teacher(t *testing.T) {
	rel, ok := (&teacher_evaluation.EvaluationDescriptor{}).DefaultRels()[teacher_evaluation.RelTeacher]
	require.True(t, ok); assert.Equal(t, "teachers", rel.Domain); assert.True(t, rel.IsAutoload)
}

func validEvaluationData() map[string]any {
	return map[string]any{
		teacher_evaluation.EvFieldTeacherID: "00000000-0000-0000-0000-000000000001",
		teacher_evaluation.EvFieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		teacher_evaluation.EvFieldSemester: teacher_evaluation.SemGanjil,
		teacher_evaluation.EvFieldEvaluatorID: "00000000-0000-0000-0000-000000000003",
	}
}
func TestEvaluationDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&teacher_evaluation.EvaluationDescriptor{}).Validate(validEvaluationData()))
}
func TestEvaluationDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := validEvaluationData(); d[teacher_evaluation.EvFieldSemester] = "sem3"
	assert.ErrorContains(t, (&teacher_evaluation.EvaluationDescriptor{}).Validate(d), "semester tidak valid")
}

func TestScoreDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "teacher_evaluation_scores", (&teacher_evaluation.ScoreDescriptor{}).TableName())
}
func TestScoreDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_evaluation.ScoreDescriptor{}
}
func TestScoreDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&teacher_evaluation.ScoreDescriptor{}).DefaultRels(), 2)
}

func validScoreData() map[string]any {
	return map[string]any{
		teacher_evaluation.ScFieldEvaluationID: "00000000-0000-0000-0000-000000000001",
		teacher_evaluation.ScFieldCompetencyID: "00000000-0000-0000-0000-000000000002",
		teacher_evaluation.ScFieldAssessorType: teacher_evaluation.AssessorSelf,
		teacher_evaluation.ScFieldScore: float64(85),
	}
}
func TestScoreDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&teacher_evaluation.ScoreDescriptor{}).Validate(validScoreData()))
}
func TestScoreDescriptor_Validate_InvalidAssessorType(t *testing.T) {
	d := validScoreData(); d[teacher_evaluation.ScFieldAssessorType] = "ext"
	assert.ErrorContains(t, (&teacher_evaluation.ScoreDescriptor{}).Validate(d), "assessor_type tidak valid")
}
func TestScoreDescriptor_Validate_ScoreOutOfRange(t *testing.T) {
	d := validScoreData(); d[teacher_evaluation.ScFieldScore] = float64(150)
	assert.ErrorContains(t, (&teacher_evaluation.ScoreDescriptor{}).Validate(d), "score harus antara")
}
