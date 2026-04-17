package curriculum_test

import (
	"testing"

	"github.com/yourorg/boilerplate/internal/domain/curriculum"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── CurriculaDescriptor tests ────────────────────────────────────────────────

func TestCurriculaDescriptor_TableName(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	if got := d.TableName(); got != "curricula" {
		t.Errorf("TableName() = %q, want %q", got, "curricula")
	}
}

func TestCurriculaDescriptor_DefaultRels(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	rels := d.DefaultRels()

	if len(rels) != 1 {
		t.Fatalf("DefaultRels() length = %d, want 1", len(rels))
	}

	rel, ok := rels["academic_year"]
	if !ok {
		t.Fatal(`DefaultRels() missing "academic_year" relation`)
	}

	if rel.Type != vernon.RelBelongsTo {
		t.Errorf("academic_year type = %q, want %q", rel.Type, vernon.RelBelongsTo)
	}
	if !rel.IsAutoload {
		t.Error("academic_year IsAutoload should be true")
	}
	if rel.FK != "academic_year_id" {
		t.Errorf("academic_year FK = %q, want %q", rel.FK, "academic_year_id")
	}
}

func TestCurriculaDescriptor_Validate_Valid(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	data := map[string]any{
		"name":            "Kurikulum Merdeka SD",
		"curriculum_type": "merdeka",
		"grade_level":     4,
		"phase":           "B",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestCurriculaDescriptor_Validate_MissingName(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	data := map[string]any{
		"curriculum_type": "merdeka",
		"grade_level":     4,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing name")
	}
}

func TestCurriculaDescriptor_Validate_MissingCurriculumType(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	data := map[string]any{
		"name":        "Kurikulum Test",
		"grade_level": 4,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing curriculum_type")
	}
}

func TestCurriculaDescriptor_Validate_MissingGradeLevel(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	data := map[string]any{
		"name":            "Kurikulum Test",
		"curriculum_type": "k13",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing grade_level")
	}
}

func TestCurriculaDescriptor_Validate_InvalidCurriculumType(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	data := map[string]any{
		"name":            "Test",
		"curriculum_type": "invalid_type",
		"grade_level":     5,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid curriculum_type")
	}
}

func TestCurriculaDescriptor_Validate_InvalidGradeLevel(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	data := map[string]any{
		"name":            "Test",
		"curriculum_type": "merdeka",
		"grade_level":     13,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject grade_level > 12")
	}
}

func TestCurriculaDescriptor_Validate_InvalidPhase(t *testing.T) {
	d := &curriculum.CurriculaDescriptor{}
	data := map[string]any{
		"name":            "Test",
		"curriculum_type": "merdeka",
		"grade_level":     5,
		"phase":           "Z",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid phase")
	}
}

// ── LearningOutcomesDescriptor tests ─────────────────────────────────────────

func TestLearningOutcomesDescriptor_TableName(t *testing.T) {
	d := &curriculum.LearningOutcomesDescriptor{}
	if got := d.TableName(); got != "learning_outcomes" {
		t.Errorf("TableName() = %q, want %q", got, "learning_outcomes")
	}
}

func TestLearningOutcomesDescriptor_DefaultRels(t *testing.T) {
	d := &curriculum.LearningOutcomesDescriptor{}
	rels := d.DefaultRels()

	if len(rels) != 2 {
		t.Fatalf("DefaultRels() length = %d, want 2", len(rels))
	}

	for _, name := range []string{"curriculum", "subject"} {
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

func TestLearningOutcomesDescriptor_Validate_Valid(t *testing.T) {
	d := &curriculum.LearningOutcomesDescriptor{}
	data := map[string]any{
		"outcome_type": "cp",
		"code":         "CP-MAT-04-01",
		"title":        "Capaian Pembelajaran Matematika Fase B",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestLearningOutcomesDescriptor_Validate_MissingOutcomeType(t *testing.T) {
	d := &curriculum.LearningOutcomesDescriptor{}
	data := map[string]any{
		"code":  "CP-MAT-04-01",
		"title": "Test Title",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing outcome_type")
	}
}

func TestLearningOutcomesDescriptor_Validate_InvalidOutcomeType(t *testing.T) {
	d := &curriculum.LearningOutcomesDescriptor{}
	data := map[string]any{
		"outcome_type": "invalid",
		"code":         "CP-MAT-04-01",
		"title":        "Test Title",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid outcome_type")
	}
}

func TestLearningOutcomesDescriptor_Validate_MissingCode(t *testing.T) {
	d := &curriculum.LearningOutcomesDescriptor{}
	data := map[string]any{
		"outcome_type": "tp",
		"title":        "Test Title",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing code")
	}
}

func TestLearningOutcomesDescriptor_Validate_MissingTitle(t *testing.T) {
	d := &curriculum.LearningOutcomesDescriptor{}
	data := map[string]any{
		"outcome_type": "atp",
		"code":         "ATP-MAT-04-01",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing title")
	}
}

// ── P5ProjectsDescriptor tests ───────────────────────────────────────────────

func TestP5ProjectsDescriptor_TableName(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	if got := d.TableName(); got != "p5_projects" {
		t.Errorf("TableName() = %q, want %q", got, "p5_projects")
	}
}

func TestP5ProjectsDescriptor_DefaultRels(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	rels := d.DefaultRels()

	if len(rels) != 2 {
		t.Fatalf("DefaultRels() length = %d, want 2", len(rels))
	}

	for _, name := range []string{"curriculum", "academic_year"} {
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

func TestP5ProjectsDescriptor_Validate_Valid(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	data := map[string]any{
		"name":        "Projek Gaya Hidup Berkelanjutan",
		"theme":       "gaya_hidup_berkelanjutan",
		"grade_level": 7,
		"semester":    "ganjil",
		"status":      "draft",
	}
	if err := d.Validate(data); err != nil {
		t.Errorf("Validate() should accept valid data, got: %v", err)
	}
}

func TestP5ProjectsDescriptor_Validate_MissingName(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	data := map[string]any{
		"theme":       "gaya_hidup_berkelanjutan",
		"grade_level": 7,
		"semester":    "ganjil",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing name")
	}
}

func TestP5ProjectsDescriptor_Validate_InvalidTheme(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	data := map[string]any{
		"name":        "Test Project",
		"theme":       "invalid_theme",
		"grade_level": 7,
		"semester":    "ganjil",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid theme")
	}
}

func TestP5ProjectsDescriptor_Validate_InvalidGradeLevel(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	data := map[string]any{
		"name":        "Test Project",
		"theme":       "kearifan_lokal",
		"grade_level": 0,
		"semester":    "genap",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject grade_level < 1")
	}
}

func TestP5ProjectsDescriptor_Validate_MissingSemester(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	data := map[string]any{
		"name":        "Test Project",
		"theme":       "budaya_kerja",
		"grade_level": 10,
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject missing semester")
	}
}

func TestP5ProjectsDescriptor_Validate_InvalidSemester(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	data := map[string]any{
		"name":        "Test Project",
		"theme":       "kewirausahaan",
		"grade_level": 10,
		"semester":    "semester1",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid semester")
	}
}

func TestP5ProjectsDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &curriculum.P5ProjectsDescriptor{}
	data := map[string]any{
		"name":        "Test Project",
		"theme":       "kewirausahaan",
		"grade_level": 10,
		"semester":    "ganjil",
		"status":      "unknown",
	}
	if err := d.Validate(data); err == nil {
		t.Error("Validate() should reject invalid status")
	}
}
