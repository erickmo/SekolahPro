// Package curriculum adalah domain Vernon untuk manajemen kurikulum.
//
// Tiga tabel: curricula (root), learning_outcomes, p5_projects.
// Curricula mereferensikan academic_year; learning_outcomes dan p5_projects
// merupakan sub-entity yang belongs_to curricula.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package curriculum

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── curricula ────────────────────────────────────────────────────────────────

// Field name constants for curricula.
const (
	CurriculaFieldName          = "name"
	CurriculaFieldCurriculumType = "curriculum_type"
	CurriculaFieldGradeLevel    = "grade_level"
	CurriculaFieldPhase         = "phase"
	CurriculaFieldAcademicYearID = "academic_year_id"
	CurriculaFieldDescription   = "description"
	CurriculaFieldStatus        = "status"
)

// Curriculum type constants.
const (
	CurriculumTypeMerdeka = "merdeka"
	CurriculumTypeK13     = "k13"
	CurriculumTypeKTSP    = "ktsp"
	CurriculumTypeDiniyah = "diniyah"
	CurriculumTypeCustom  = "custom"
)

// Phase constants (Kurikulum Merdeka phases A-F).
const (
	PhaseA = "A"
	PhaseB = "B"
	PhaseC = "C"
	PhaseD = "D"
	PhaseE = "E"
	PhaseF = "F"
)

// Relation name constants.
const (
	CurriculaRelAcademicYear = "academic_year"
)

const (
	curriculaGradeMin = 1
	curriculaGradeMax = 12
)

var validCurriculumTypes = map[string]bool{
	CurriculumTypeMerdeka: true,
	CurriculumTypeK13:     true,
	CurriculumTypeKTSP:    true,
	CurriculumTypeDiniyah: true,
	CurriculumTypeCustom:  true,
}

var validPhases = map[string]bool{
	PhaseA: true, PhaseB: true, PhaseC: true,
	PhaseD: true, PhaseE: true, PhaseF: true,
}

// CurriculaDescriptor mengimplementasi vernon.DomainDescriptor untuk curricula.
type CurriculaDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *CurriculaDescriptor) TableName() string { return "curricula" }

// DefaultRels mendefinisikan relasi domain curricula.
// belongs_to academic_year (autoload).
func (d *CurriculaDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		CurriculaRelAcademicYear: {
			Domain:     "academic_year",
			Type:       vernon.RelBelongsTo,
			FK:         CurriculaFieldAcademicYearID,
			LocalKey:   CurriculaFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "start_date", "end_date"},
		},
	}
}

// Validate memvalidasi invariant domain curricula sebelum write.
func (d *CurriculaDescriptor) Validate(data map[string]any) error {
	if err := curriculaValidateRequired(data); err != nil {
		return err
	}
	return curriculaValidateEnums(data)
}

// curriculaValidateRequired memeriksa field wajib curricula.
func curriculaValidateRequired(data map[string]any) error {
	name, _ := data[CurriculaFieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}

	curriculumType, _ := data[CurriculaFieldCurriculumType].(string)
	if curriculumType == "" {
		return errors.New("curriculum_type wajib diisi")
	}
	if !validCurriculumTypes[curriculumType] {
		return fmt.Errorf("curriculum_type tidak valid: %q (harus merdeka/k13/ktsp/diniyah/custom)", curriculumType)
	}

	gradeLevel, ok := toInt(data[CurriculaFieldGradeLevel])
	if !ok {
		return errors.New("grade_level wajib diisi")
	}
	if gradeLevel < curriculaGradeMin || gradeLevel > curriculaGradeMax {
		return fmt.Errorf("grade_level tidak valid: %d (harus 1-12)", gradeLevel)
	}
	return nil
}

// curriculaValidateEnums memeriksa enum fields curricula.
func curriculaValidateEnums(data map[string]any) error {
	phase, _ := data[CurriculaFieldPhase].(string)
	if phase != "" && !validPhases[phase] {
		return fmt.Errorf("phase tidak valid: %q (harus A-F)", phase)
	}
	return nil
}

// ── learning_outcomes ────────────────────────────────────────────────────────

// Field name constants for learning_outcomes.
const (
	LOFieldOutcomeType  = "outcome_type"
	LOFieldCode         = "code"
	LOFieldTitle        = "title"
	LOFieldDescription  = "description"
	LOFieldCurriculumID = "curriculum_id"
	LOFieldSubjectID    = "subject_id"
	LOFieldGradeLevel   = "grade_level"
	LOFieldSemester     = "semester"
)

// Outcome type constants.
const (
	OutcomeTypeCP  = "cp"
	OutcomeTypeTP  = "tp"
	OutcomeTypeATP = "atp"
)

// Relation name constants.
const (
	LORelCurriculum = "curriculum"
	LORelSubject    = "subject"
)

var validOutcomeTypes = map[string]bool{
	OutcomeTypeCP: true, OutcomeTypeTP: true, OutcomeTypeATP: true,
}

// LearningOutcomesDescriptor mengimplementasi vernon.DomainDescriptor untuk learning_outcomes.
type LearningOutcomesDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *LearningOutcomesDescriptor) TableName() string { return "learning_outcomes" }

// DefaultRels mendefinisikan relasi domain learning_outcomes.
// belongs_to curriculum (autoload) dan subject (autoload).
func (d *LearningOutcomesDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		LORelCurriculum: {
			Domain:     "curricula",
			Type:       vernon.RelBelongsTo,
			FK:         LOFieldCurriculumID,
			LocalKey:   LOFieldCurriculumID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "curriculum_type", "grade_level"},
		},
		LORelSubject: {
			Domain:     "subjects",
			Type:       vernon.RelBelongsTo,
			FK:         LOFieldSubjectID,
			LocalKey:   LOFieldSubjectID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code", "subject_group"},
		},
	}
}

// Validate memvalidasi invariant domain learning_outcomes sebelum write.
func (d *LearningOutcomesDescriptor) Validate(data map[string]any) error {
	return loValidateRequired(data)
}

// loValidateRequired memeriksa field wajib learning_outcomes.
func loValidateRequired(data map[string]any) error {
	outcomeType, _ := data[LOFieldOutcomeType].(string)
	if outcomeType == "" {
		return errors.New("outcome_type wajib diisi")
	}
	if !validOutcomeTypes[outcomeType] {
		return fmt.Errorf("outcome_type tidak valid: %q (harus cp/tp/atp)", outcomeType)
	}

	code, _ := data[LOFieldCode].(string)
	if code == "" {
		return errors.New("code wajib diisi")
	}

	title, _ := data[LOFieldTitle].(string)
	if title == "" {
		return errors.New("title wajib diisi")
	}
	return nil
}

// ── p5_projects ──────────────────────────────────────────────────────────────

// Field name constants for p5_projects.
const (
	P5FieldName          = "name"
	P5FieldTheme         = "theme"
	P5FieldDescription   = "description"
	P5FieldGradeLevel    = "grade_level"
	P5FieldSemester      = "semester"
	P5FieldCurriculumID  = "curriculum_id"
	P5FieldAcademicYearID = "academic_year_id"
	P5FieldStatus        = "status"
)

// P5 theme constants (7 tema P5 Kurikulum Merdeka).
const (
	P5ThemeGayaHidupBerkelanjutan  = "gaya_hidup_berkelanjutan"
	P5ThemeKearifanLokal           = "kearifan_lokal"
	P5ThemeBhinnekaTunggalIka      = "bhinneka_tunggal_ika"
	P5ThemeKewirausahaan           = "kewirausahaan"
	P5ThemeRekayasaDanTeknologi    = "rekayasa_dan_teknologi"
	P5ThemePembangunanBerkelanjutan = "pembangunan_berkelanjutan"
	P5ThemeBudayaKerja             = "budaya_kerja"
)

// Semester constants.
const (
	SemesterGanjil = "ganjil"
	SemesterGenap  = "genap"
)

// Status constants for p5_projects.
const (
	P5StatusDraft     = "draft"
	P5StatusActive    = "active"
	P5StatusCompleted = "completed"
	P5StatusArchived  = "archived"
)

// Relation name constants.
const (
	P5RelCurriculum    = "curriculum"
	P5RelAcademicYear  = "academic_year"
)

var validP5Themes = map[string]bool{
	P5ThemeGayaHidupBerkelanjutan: true,
	P5ThemeKearifanLokal:          true,
	P5ThemeBhinnekaTunggalIka:     true,
	P5ThemeKewirausahaan:          true,
	P5ThemeRekayasaDanTeknologi:   true,
	P5ThemePembangunanBerkelanjutan: true,
	P5ThemeBudayaKerja:            true,
}

var validP5Statuses = map[string]bool{
	P5StatusDraft: true, P5StatusActive: true,
	P5StatusCompleted: true, P5StatusArchived: true,
}

var validP5Semesters = map[string]bool{
	SemesterGanjil: true, SemesterGenap: true,
}

// P5ProjectsDescriptor mengimplementasi vernon.DomainDescriptor untuk p5_projects.
type P5ProjectsDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *P5ProjectsDescriptor) TableName() string { return "p5_projects" }

// DefaultRels mendefinisikan relasi domain p5_projects.
// belongs_to curriculum (autoload) dan academic_year (autoload).
func (d *P5ProjectsDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		P5RelCurriculum: {
			Domain:     "curricula",
			Type:       vernon.RelBelongsTo,
			FK:         P5FieldCurriculumID,
			LocalKey:   P5FieldCurriculumID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "curriculum_type"},
		},
		P5RelAcademicYear: {
			Domain:     "academic_year",
			Type:       vernon.RelBelongsTo,
			FK:         P5FieldAcademicYearID,
			LocalKey:   P5FieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "start_date", "end_date"},
		},
	}
}

// Validate memvalidasi invariant domain p5_projects sebelum write.
func (d *P5ProjectsDescriptor) Validate(data map[string]any) error {
	if err := p5ValidateRequired(data); err != nil {
		return err
	}
	return p5ValidateEnums(data)
}

// p5ValidateRequired memeriksa field wajib p5_projects.
func p5ValidateRequired(data map[string]any) error {
	name, _ := data[P5FieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}

	theme, _ := data[P5FieldTheme].(string)
	if theme == "" {
		return errors.New("theme wajib diisi")
	}
	if !validP5Themes[theme] {
		return fmt.Errorf("theme tidak valid: %q", theme)
	}

	gradeLevel, ok := toInt(data[P5FieldGradeLevel])
	if !ok {
		return errors.New("grade_level wajib diisi")
	}
	if gradeLevel < curriculaGradeMin || gradeLevel > curriculaGradeMax {
		return fmt.Errorf("grade_level tidak valid: %d (harus 1-12)", gradeLevel)
	}

	semester, _ := data[P5FieldSemester].(string)
	if semester == "" {
		return errors.New("semester wajib diisi")
	}
	if !validP5Semesters[semester] {
		return fmt.Errorf("semester tidak valid: %q (harus ganjil/genap)", semester)
	}
	return nil
}

// p5ValidateEnums memeriksa enum fields p5_projects.
func p5ValidateEnums(data map[string]any) error {
	status, _ := data[P5FieldStatus].(string)
	if status != "" && !validP5Statuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus draft/active/completed/archived)", status)
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// toInt mengkonversi value dari map[string]any ke int.
// Mendukung int, int64, dan float64 (JSON unmarshaling).
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
