package subject_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/subject"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── SubjectsDescriptor tests ─────────────────────────────────────────────────

func TestSubjectsDescriptor_TableName(t *testing.T) {
	d := &subject.SubjectsDescriptor{}
	if got := d.TableName(); got != "subjects" {
		t.Errorf("TableName() = %q, want %q", got, "subjects")
	}
}

func TestSubjectsDescriptor_DefaultRels_Empty(t *testing.T) {
	d := &subject.SubjectsDescriptor{}
	rels := d.DefaultRels()
	if len(rels) != 0 {
		t.Errorf("DefaultRels() length = %d, want 0 (root entity)", len(rels))
	}
}

func TestSubjectsDescriptor_Validate_Valid(t *testing.T) {
	d := &subject.SubjectsDescriptor{}
	data := map[string]any{
		"name":          "Matematika",
		"code":          "MAT-01",
		"subject_group": "matematika",
		"category":      "wajib_nasional",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestSubjectsDescriptor_Validate_MissingName(t *testing.T) {
	d := &subject.SubjectsDescriptor{}
	data := map[string]any{
		"code": "MAT-01",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing name")
	}
}

func TestSubjectsDescriptor_Validate_MissingCode(t *testing.T) {
	d := &subject.SubjectsDescriptor{}
	data := map[string]any{
		"name": "Matematika",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing code")
	}
}

func TestSubjectsDescriptor_Validate_InvalidSubjectGroup(t *testing.T) {
	d := &subject.SubjectsDescriptor{}
	data := map[string]any{
		"name":          "Test",
		"code":          "TST-01",
		"subject_group": "invalid_group",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid subject_group")
	}
}

func TestSubjectsDescriptor_Validate_InvalidCategory(t *testing.T) {
	d := &subject.SubjectsDescriptor{}
	data := map[string]any{
		"name":     "Test",
		"code":     "TST-01",
		"category": "invalid_category",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid category")
	}
}

// ── SubjectConfigurationsDescriptor tests ────────────────────────────────────

func TestSubjectConfigurationsDescriptor_TableName(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	if got := d.TableName(); got != "subject_configurations" {
		t.Errorf("TableName() = %q, want %q", got, "subject_configurations")
	}
}

func TestSubjectConfigurationsDescriptor_DefaultRels(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	rels := d.DefaultRels()

	if len(rels) != 2 {
		t.Fatalf("DefaultRels() length = %d, want 2", len(rels))
	}

	for _, name := range []string{"subject", "academic_year"} {
		rel, ok := rels[name]
		if !ok {
			t.Fatalf("DefaultRels() missing %q relation", name)
		}
		if rel.Type != vernon.RelBelongsTo {
			t.Errorf("%s type = %q, want %q", name, rel.Type, vernon.RelBelongsTo)
		}
		if !rel.IsAutoload {
			t.Errorf("%s IsAutoload should be true", name)
		}
	}
}

func TestSubjectConfigurationsDescriptor_Validate_Valid(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	data := map[string]any{
		"subject_id":          "uuid-subject-001",
		"academic_year_id":    "uuid-year-001",
		"grade_level":         7,
		"credit_hours_per_week": 4,
		"weight_knowledge":    60,
		"weight_skill":        40,
		"passing_grade":       75,
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestSubjectConfigurationsDescriptor_Validate_MissingSubjectID(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	data := map[string]any{
		"academic_year_id": "uuid-year-001",
		"grade_level":      7,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing subject_id")
	}
}

func TestSubjectConfigurationsDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	data := map[string]any{
		"subject_id":  "uuid-subject-001",
		"grade_level": 7,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing academic_year_id")
	}
}

func TestSubjectConfigurationsDescriptor_Validate_InvalidGradeLevel(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	data := map[string]any{
		"subject_id":       "uuid-subject-001",
		"academic_year_id": "uuid-year-001",
		"grade_level":      13,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject grade_level > 12")
	}
}

func TestSubjectConfigurationsDescriptor_Validate_InvalidCreditHours(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	data := map[string]any{
		"subject_id":            "uuid-subject-001",
		"academic_year_id":      "uuid-year-001",
		"grade_level":           7,
		"credit_hours_per_week": 15,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject credit_hours_per_week > 12")
	}
}

func TestSubjectConfigurationsDescriptor_Validate_WeightSumNot100(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	data := map[string]any{
		"subject_id":        "uuid-subject-001",
		"academic_year_id":  "uuid-year-001",
		"grade_level":       7,
		"weight_knowledge":  50,
		"weight_skill":      30,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject weight_knowledge + weight_skill != 100")
	}
}

func TestSubjectConfigurationsDescriptor_Validate_InvalidPassingGrade(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	data := map[string]any{
		"subject_id":       "uuid-subject-001",
		"academic_year_id": "uuid-year-001",
		"grade_level":      7,
		"passing_grade":    150,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject passing_grade > 100")
	}
}

func TestSubjectConfigurationsDescriptor_Validate_GradeLevelAsFloat(t *testing.T) {
	d := &subject.SubjectConfigurationsDescriptor{}
	data := map[string]any{
		"subject_id":       "uuid-subject-001",
		"academic_year_id": "uuid-year-001",
		"grade_level":      float64(7),
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept grade_level as float64, got: %v", err)
	}
}
