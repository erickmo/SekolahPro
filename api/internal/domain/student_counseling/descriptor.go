// Package student_counseling adalah domain Vernon untuk Bimbingan Konseling (BK).
//
// Terdiri dari 2 tabel: counseling_cases dan counseling_sessions.
// counseling_cases memiliki 2 BelongsTo autoload: student dan academic_year.
// counseling_sessions memiliki 2 BelongsTo autoload: case dan student.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package student_counseling

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Case field name constants ──────────────────────────────────────────────────

const (
	FieldStudentID      = "student_id"
	FieldAcademicYearID = "academic_year_id"
	FieldCaseNo         = "case_no"
	FieldCategory       = "category"
	FieldTitle          = "title"
	FieldSeverity       = "severity"
	FieldOpenedDate     = "opened_date"
	FieldCounselorID    = "counselor_id"
	FieldStatus         = "status"
)

// ── Session field name constants ───────────────────────────────────────────────

const (
	FieldCaseID      = "case_id"
	FieldSessionDate = "session_date"
	FieldSessionType = "session_type"
	FieldNotes       = "notes"
)

// ── Case enum constants ────────────────────────────────────────────────────────

const (
	CategoryAcademic  = "academic"
	CategorySocial    = "social"
	CategoryPersonal  = "personal"
	CategoryCareer    = "career"
	CategoryBehavioral = "behavioral"
	CategoryFamily    = "family"
	CategoryOther     = "other"
)

const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

const (
	CaseStatusOpen       = "open"
	CaseStatusInProgress = "in_progress"
	CaseStatusReferred   = "referred"
	CaseStatusResolved   = "resolved"
	CaseStatusClosed     = "closed"
)

// ── Session enum constants ─────────────────────────────────────────────────────

const (
	SessionTypeIndividual    = "individual"
	SessionTypeGroup         = "group"
	SessionTypeHomeVisit     = "home_visit"
	SessionTypeParentConf    = "parent_conference"
	SessionTypeReferral      = "referral"
)

// ── Relation name constants ────────────────────────────────────────────────────

const (
	RelStudent      = "student"
	RelAcademicYear = "academic_year"
	RelCase         = "case"
)

var validCategories = map[string]bool{
	CategoryAcademic: true, CategorySocial: true, CategoryPersonal: true,
	CategoryCareer: true, CategoryBehavioral: true, CategoryFamily: true,
	CategoryOther: true,
}

var validSeverities = map[string]bool{
	SeverityLow: true, SeverityMedium: true,
	SeverityHigh: true, SeverityCritical: true,
}

var validCaseStatuses = map[string]bool{
	CaseStatusOpen: true, CaseStatusInProgress: true,
	CaseStatusReferred: true, CaseStatusResolved: true,
	CaseStatusClosed: true,
}

var validSessionTypes = map[string]bool{
	SessionTypeIndividual: true, SessionTypeGroup: true,
	SessionTypeHomeVisit: true, SessionTypeParentConf: true,
	SessionTypeReferral: true,
}

// ── CaseDescriptor ─────────────────────────────────────────────────────────────

// CaseDescriptor mengimplementasi vernon.DomainDescriptor untuk counseling_cases.
type CaseDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL untuk kasus BK.
func (d *CaseDescriptor) TableName() string { return "counseling_cases" }

// DefaultRels mendefinisikan relasi counseling_cases: student + academic_year.
func (d *CaseDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         FieldStudentID,
			LocalKey:   FieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
		RelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         FieldAcademicYearID,
			LocalKey:   FieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant counseling_cases sebelum write.
func (d *CaseDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldStudentID, "student_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldAcademicYearID, "academic_year_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldCaseNo, "case_no"); err != nil {
		return err
	}
	if err := requireEnum(data, FieldCategory, "category", validCategories); err != nil {
		return err
	}
	if err := requireString(data, FieldTitle, "title"); err != nil {
		return err
	}
	if err := requireEnum(data, FieldSeverity, "severity", validSeverities); err != nil {
		return err
	}
	if err := requireString(data, FieldOpenedDate, "opened_date"); err != nil {
		return err
	}
	return requireString(data, FieldCounselorID, "counselor_id")
}

// ── SessionDescriptor ──────────────────────────────────────────────────────────

// SessionDescriptor mengimplementasi vernon.DomainDescriptor untuk counseling_sessions.
type SessionDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL untuk sesi konseling.
func (d *SessionDescriptor) TableName() string { return "counseling_sessions" }

// DefaultRels mendefinisikan relasi counseling_sessions: case + student.
func (d *SessionDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelCase: {
			Domain:     "counseling_cases",
			Type:       vernon.RelBelongsTo,
			FK:         FieldCaseID,
			LocalKey:   FieldCaseID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"case_no", "title", "status"},
		},
		RelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         FieldStudentID,
			LocalKey:   FieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
	}
}

// Validate memvalidasi invariant counseling_sessions sebelum write.
func (d *SessionDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, FieldCaseID, "case_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldStudentID, "student_id"); err != nil {
		return err
	}
	if err := requireString(data, FieldSessionDate, "session_date"); err != nil {
		return err
	}
	if err := requireEnum(data, FieldSessionType, "session_type", validSessionTypes); err != nil {
		return err
	}
	return requireString(data, FieldNotes, "notes")
}

// ── shared validation helpers ──────────────────────────────────────────────────

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
