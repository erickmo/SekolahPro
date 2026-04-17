// Package subject adalah domain Vernon untuk manajemen mata pelajaran.
//
// Dua tabel: subjects (root entity), subject_configurations.
// Subject tidak memiliki belongs_to — ia adalah entity root yang direferensikan
// oleh banyak domain lain (learning_outcomes, nilai, jadwal, dll).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package subject

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── subjects ─────────────────────────────────────────────────────────────────

// Field name constants for subjects.
const (
	SubjectsFieldName         = "name"
	SubjectsFieldCode        = "code"
	SubjectsFieldSubjectGroup = "subject_group"
	SubjectsFieldCategory    = "category"
	SubjectsFieldDescription = "description"
	SubjectsFieldIsActive    = "is_active"
)

// Subject group constants (16 kelompok mata pelajaran).
const (
	SubjectGroupAgama         = "agama"
	SubjectGroupPKN           = "pkn"
	SubjectGroupBahasa        = "bahasa"
	SubjectGroupMatematika    = "matematika"
	SubjectGroupIPA           = "ipa"
	SubjectGroupIPS           = "ips"
	SubjectGroupSeniBudaya    = "seni_budaya"
	SubjectGroupPJOK          = "pjok"
	SubjectGroupPrakarya      = "prakarya"
	SubjectGroupInformatika   = "informatika"
	SubjectGroupMuatanLokal   = "muatan_lokal"
	SubjectGroupKeagamaan     = "keagamaan"
	SubjectGroupTahfidz       = "tahfidz"
	SubjectGroupLintasMinat   = "lintas_minat"
	SubjectGroupPeminatan     = "peminatan"
	SubjectGroupOther         = "other"
)

// Category constants (5 kategori).
const (
	CategoryWajibNasional  = "wajib_nasional"
	CategoryWajibLokal     = "wajib_lokal"
	CategoryPilihan        = "pilihan"
	CategoryDiniyah        = "diniyah"
	CategoryEkstraKurikuler = "ekstra_kurikuler"
)

var validSubjectGroups = map[string]bool{
	SubjectGroupAgama: true, SubjectGroupPKN: true,
	SubjectGroupBahasa: true, SubjectGroupMatematika: true,
	SubjectGroupIPA: true, SubjectGroupIPS: true,
	SubjectGroupSeniBudaya: true, SubjectGroupPJOK: true,
	SubjectGroupPrakarya: true, SubjectGroupInformatika: true,
	SubjectGroupMuatanLokal: true, SubjectGroupKeagamaan: true,
	SubjectGroupTahfidz: true, SubjectGroupLintasMinat: true,
	SubjectGroupPeminatan: true, SubjectGroupOther: true,
}

var validCategories = map[string]bool{
	CategoryWajibNasional: true, CategoryWajibLokal: true,
	CategoryPilihan: true, CategoryDiniyah: true,
	CategoryEkstraKurikuler: true,
}

// SubjectsDescriptor mengimplementasi vernon.DomainDescriptor untuk subjects.
type SubjectsDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *SubjectsDescriptor) TableName() string { return "subjects" }

// DefaultRels mendefinisikan relasi domain subjects.
// subjects adalah entity root — tidak punya belongs_to.
func (d *SubjectsDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain subjects sebelum write.
func (d *SubjectsDescriptor) Validate(data map[string]any) error {
	if err := subjectsValidateRequired(data); err != nil {
		return err
	}
	return subjectsValidateEnums(data)
}

// subjectsValidateRequired memeriksa field wajib subjects.
func subjectsValidateRequired(data map[string]any) error {
	name, _ := data[SubjectsFieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}

	code, _ := data[SubjectsFieldCode].(string)
	if code == "" {
		return errors.New("code wajib diisi")
	}
	return nil
}

// subjectsValidateEnums memeriksa enum fields subjects.
func subjectsValidateEnums(data map[string]any) error {
	subjectGroup, _ := data[SubjectsFieldSubjectGroup].(string)
	if subjectGroup != "" && !validSubjectGroups[subjectGroup] {
		return fmt.Errorf("subject_group tidak valid: %q", subjectGroup)
	}

	category, _ := data[SubjectsFieldCategory].(string)
	if category != "" && !validCategories[category] {
		return fmt.Errorf("category tidak valid: %q", category)
	}
	return nil
}

// ── subject_configurations ───────────────────────────────────────────────────

// Field name constants for subject_configurations.
const (
	SubConfFieldSubjectID         = "subject_id"
	SubConfFieldAcademicYearID    = "academic_year_id"
	SubConfFieldGradeLevel        = "grade_level"
	SubConfFieldCreditHoursPerWeek = "credit_hours_per_week"
	SubConfFieldWeightKnowledge   = "weight_knowledge"
	SubConfFieldWeightSkill       = "weight_skill"
	SubConfFieldPassingGrade      = "passing_grade"
	SubConfFieldSemester          = "semester"
)

// Relation name constants.
const (
	SubConfRelSubject      = "subject"
	SubConfRelAcademicYear = "academic_year"
)

const (
	subConfGradeMin         = 1
	subConfGradeMax         = 12
	subConfCreditMin        = 1
	subConfCreditMax        = 12
	subConfWeightTotal      = 100
	subConfPassingGradeMin  = 0
	subConfPassingGradeMax  = 100
)

// SubjectConfigurationsDescriptor mengimplementasi vernon.DomainDescriptor untuk subject_configurations.
type SubjectConfigurationsDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *SubjectConfigurationsDescriptor) TableName() string { return "subject_configurations" }

// DefaultRels mendefinisikan relasi domain subject_configurations.
// belongs_to subject (autoload) dan academic_year (autoload).
func (d *SubjectConfigurationsDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		SubConfRelSubject: {
			Domain:     "subjects",
			Type:       vernon.RelBelongsTo,
			FK:         SubConfFieldSubjectID,
			LocalKey:   SubConfFieldSubjectID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code", "subject_group", "category"},
		},
		SubConfRelAcademicYear: {
			Domain:     "academic_year",
			Type:       vernon.RelBelongsTo,
			FK:         SubConfFieldAcademicYearID,
			LocalKey:   SubConfFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "start_date", "end_date"},
		},
	}
}

// Validate memvalidasi invariant domain subject_configurations sebelum write.
func (d *SubjectConfigurationsDescriptor) Validate(data map[string]any) error {
	if err := subConfValidateRequired(data); err != nil {
		return err
	}
	return subConfValidateRanges(data)
}

// subConfValidateRequired memeriksa field wajib subject_configurations.
func subConfValidateRequired(data map[string]any) error {
	subjectID, _ := data[SubConfFieldSubjectID].(string)
	if subjectID == "" {
		return errors.New("subject_id wajib diisi")
	}

	academicYearID, _ := data[SubConfFieldAcademicYearID].(string)
	if academicYearID == "" {
		return errors.New("academic_year_id wajib diisi")
	}
	return nil
}

// subConfValidateRanges memeriksa range constraints subject_configurations.
func subConfValidateRanges(data map[string]any) error {
	if err := subConfValidateGradeLevel(data); err != nil {
		return err
	}

	if err := subConfValidateCreditHours(data); err != nil {
		return err
	}

	if err := subConfValidateWeights(data); err != nil {
		return err
	}

	return subConfValidatePassingGrade(data)
}

// subConfValidateGradeLevel memeriksa grade_level 1-12.
func subConfValidateGradeLevel(data map[string]any) error {
	gradeLevel, ok := toInt(data[SubConfFieldGradeLevel])
	if !ok {
		return errors.New("grade_level wajib diisi")
	}
	if gradeLevel < subConfGradeMin || gradeLevel > subConfGradeMax {
		return fmt.Errorf("grade_level tidak valid: %d (harus 1-12)", gradeLevel)
	}
	return nil
}

// subConfValidateCreditHours memeriksa credit_hours_per_week 1-12.
func subConfValidateCreditHours(data map[string]any) error {
	credit, ok := toInt(data[SubConfFieldCreditHoursPerWeek])
	if !ok {
		return nil // opsional
	}
	if credit < subConfCreditMin || credit > subConfCreditMax {
		return fmt.Errorf("credit_hours_per_week tidak valid: %d (harus 1-12)", credit)
	}
	return nil
}

// subConfValidateWeights memeriksa weight_knowledge + weight_skill = 100.
func subConfValidateWeights(data map[string]any) error {
	wk, hasWK := toInt(data[SubConfFieldWeightKnowledge])
	ws, hasWS := toInt(data[SubConfFieldWeightSkill])

	if !hasWK || !hasWS {
		return nil // keduanya harus ada untuk divalidasi
	}

	total := wk + ws
	if total != subConfWeightTotal {
		return fmt.Errorf("weight_knowledge + weight_skill harus = 100, got %d", total)
	}
	return nil
}

// subConfValidatePassingGrade memeriksa passing_grade 0-100.
func subConfValidatePassingGrade(data map[string]any) error {
	pg, ok := toFloat(data[SubConfFieldPassingGrade])
	if !ok {
		return nil // opsional
	}
	if pg < subConfPassingGradeMin || pg > subConfPassingGradeMax {
		return fmt.Errorf("passing_grade tidak valid: %.0f (harus 0-100)", pg)
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

// toFloat mengkonversi value dari map[string]any ke float64.
// Mendukung float64, int, dan int64.
func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
