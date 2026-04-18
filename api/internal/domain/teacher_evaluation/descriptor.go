// Package teacher_evaluation adalah domain Vernon untuk evaluasi kinerja guru (PKG).
//
// Terdiri dari 3 tabel: teacher_evaluation_competencies, teacher_evaluations, teacher_evaluation_scores.
// teacher_evaluation_competencies tidak memiliki BelongsTo (root master).
// teacher_evaluations memiliki 2 BelongsTo autoload: teacher, academic_year.
// teacher_evaluation_scores memiliki 2 BelongsTo autoload: evaluation, competency.
// assessor adalah polymorphic (3 tipe) dan TIDAK di-autoload.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teacher_evaluation

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Competency field constants ────────────────────────────────────────────────

const (
	CompFieldName            = "name"
	CompFieldArea            = "area"
	CompFieldIndicatorCount  = "indicator_count"
	CompFieldWeight          = "weight"
	CompFieldDescription     = "description"
	CompFieldIsActive        = "is_active"
)

// ── Evaluation field constants ────────────────────────────────────────────────

const (
	EvFieldTeacherID      = "teacher_id"
	EvFieldAcademicYearID = "academic_year_id"
	EvFieldSemester       = "semester"
	EvFieldEvaluatorID    = "evaluator_id"
	EvFieldTotalScore     = "total_score"
	EvFieldGrade          = "grade"
	EvFieldWorkflowStatus = "workflow_status"
	EvFieldEvaluationDate = "evaluation_date"
	EvFieldNotes          = "notes"
)

// ── Score field constants ─────────────────────────────────────────────────────

const (
	ScFieldEvaluationID  = "evaluation_id"
	ScFieldCompetencyID  = "competency_id"
	ScFieldAssessorType  = "assessor_type"
	ScFieldAssessorID    = "assessor_id"
	ScFieldScore         = "score"
	ScFieldEvidence      = "evidence"
	ScFieldNotes         = "notes"
)

// ── Competency enum constants ─────────────────────────────────────────────────

const (
	AreaPedagogic   = "pedagogic"
	AreaPersonality = "personality"
	AreaSocial      = "social"
	AreaProfessional = "professional"
)

// ── Evaluation enum constants ─────────────────────────────────────────────────

const (
	SemGanjil = "ganjil"
	SemGenap  = "genap"
)

const (
	GradeA = "A"
	GradeB = "B"
	GradeC = "C"
	GradeD = "D"
	GradeE = "E"
)

const (
	WSDraft            = "draft"
	WSSelfAssessment   = "self_assessment"
	WSPeerReview       = "peer_review"
	WSSupervisorReview = "supervisor_review"
	WSFinal            = "final"
	WSApproved         = "approved"
)

// ── Score enum constants ──────────────────────────────────────────────────────

const (
	AssessorSelf       = "self"
	AssessorPeer       = "peer"
	AssessorSupervisor = "supervisor"
)

// ── Relation name constants ───────────────────────────────────────────────────

const (
	RelTeacher       = "teacher"
	RelAcademicYear  = "academic_year"
	RelEvaluation    = "evaluation"
	RelCompetency    = "competency"
)

var validAreas = map[string]bool{
	AreaPedagogic: true, AreaPersonality: true, AreaSocial: true, AreaProfessional: true,
}

var validSemesters = map[string]bool{
	SemGanjil: true, SemGenap: true,
}

var validGrades = map[string]bool{
	GradeA: true, GradeB: true, GradeC: true, GradeD: true, GradeE: true,
}

var validWorkflowStatuses = map[string]bool{
	WSDraft: true, WSSelfAssessment: true, WSPeerReview: true,
	WSSupervisorReview: true, WSFinal: true, WSApproved: true,
}

var validAssessorTypes = map[string]bool{
	AssessorSelf: true, AssessorPeer: true, AssessorSupervisor: true,
}

// ── CompetencyDescriptor ──────────────────────────────────────────────────────

// CompetencyDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_evaluation_competencies.
type CompetencyDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *CompetencyDescriptor) TableName() string { return "teacher_evaluation_competencies" }

// DefaultRels — competencies tidak memiliki relasi (root master).
func (d *CompetencyDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant teacher_evaluation_competencies.
func (d *CompetencyDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, CompFieldName, "name"); err != nil {
		return err
	}
	return requireEnum(data, CompFieldArea, "area", validAreas)
}

// ── EvaluationDescriptor ──────────────────────────────────────────────────────

// EvaluationDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_evaluations.
type EvaluationDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *EvaluationDescriptor) TableName() string { return "teacher_evaluations" }

// DefaultRels mendefinisikan relasi teacher_evaluations: teacher + academic_year.
func (d *EvaluationDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         EvFieldTeacherID,
			LocalKey:   EvFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "employee_type"},
		},
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         EvFieldAcademicYearID,
			LocalKey:   EvFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name"},
		},
	}
}

// Validate memvalidasi invariant teacher_evaluations.
func (d *EvaluationDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, EvFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	if err := requireString(data, EvFieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}
	if err := requireEnum(data, EvFieldSemester, "semester", validSemesters); err != nil {
		return err
	}
	return requireString(data, EvFieldEvaluatorID, "evaluator_id")
}

// ── ScoreDescriptor ───────────────────────────────────────────────────────────

// ScoreDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_evaluation_scores.
type ScoreDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ScoreDescriptor) TableName() string { return "teacher_evaluation_scores" }

// DefaultRels mendefinisikan relasi teacher_evaluation_scores: evaluation + competency.
// assessor adalah polymorphic dan TIDAK di-autoload.
func (d *ScoreDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelEvaluation: {
			Domain:     "teacher_evaluations",
			Type:       vernon.RelBelongsTo,
			FK:         ScFieldEvaluationID,
			LocalKey:   ScFieldEvaluationID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"teacher_id", "semester", "workflow_status"},
		},
		RelCompetency: {
			Domain:     "teacher_evaluation_competencies",
			Type:       vernon.RelBelongsTo,
			FK:         ScFieldCompetencyID,
			LocalKey:   ScFieldCompetencyID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "area", "weight"},
		},
	}
}

// Validate memvalidasi invariant teacher_evaluation_scores.
func (d *ScoreDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, ScFieldEvaluationID, "evaluation_id"); err != nil {
		return err
	}
	if err := requireString(data, ScFieldCompetencyID, "competency_id"); err != nil {
		return err
	}
	if err := requireEnum(data, ScFieldAssessorType, "assessor_type", validAssessorTypes); err != nil {
		return err
	}
	return requireIntRange(data, ScFieldScore, "score", 1, 100)
}

// ── shared validation helpers ─────────────────────────────────────────────────

func requireString(data map[string]any, field, label string) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	return nil
}

func requireEnum(data map[string]any, field, label string, valid map[string]bool) error {
	val, _ := data[field].(string)
	if val == "" {
		return fmt.Errorf("%s wajib diisi", label)
	}
	if !valid[val] {
		return fmt.Errorf("%s tidak valid: %q", label, val)
	}
	return nil
}

func requireIntRange(data map[string]any, field, label string, min, max int) error {
	val, ok := data[field]
	if !ok {
		return fmt.Errorf("%s wajib diisi", label)
	}
	var intVal int
	switch v := val.(type) {
	case int:
		intVal = v
	case float64:
		intVal = int(v)
	default:
		return fmt.Errorf("%s harus berupa angka", label)
	}
	if intVal < min || intVal > max {
		return fmt.Errorf("%s harus antara %d dan %d", label, min, max)
	}
	return nil
}
