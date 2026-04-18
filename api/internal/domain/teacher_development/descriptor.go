// Package teacher_development adalah domain Vernon untuk pengembangan profesional guru (PKB).
//
// Terdiri dari 3 tabel: teacher_certifications, teacher_development_activities, teacher_credit_summaries.
// Ketiganya memiliki 1 BelongsTo autoload: teacher.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teacher_development

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Certification field constants ─────────────────────────────────────────────

const (
	CertFieldTeacherID         = "teacher_id"
	CertFieldCertificationType = "certification_type"
	CertFieldCertNumber        = "certification_number"
	CertFieldIssueDate         = "issue_date"
	CertFieldExpiryDate        = "expiry_date"
	CertFieldIssuingBody       = "issuing_body"
	CertFieldStatus            = "status"
	CertFieldNotes             = "notes"
)

// ── Activity field constants ──────────────────────────────────────────────────

const (
	ActFieldTeacherID     = "teacher_id"
	ActFieldActivityType  = "activity_type"
	ActFieldTitle         = "title"
	ActFieldOrganizer     = "organizer"
	ActFieldStartDate     = "start_date"
	ActFieldEndDate       = "end_date"
	ActFieldLocation      = "location"
	ActFieldCreditPoints  = "credit_points"
	ActFieldCertNumber    = "certificate_number"
	ActFieldStatus        = "status"
	ActFieldDescription   = "description"
)

// ── Credit summary field constants ────────────────────────────────────────────

const (
	CSFieldTeacherID          = "teacher_id"
	CSFieldCurrentRank        = "current_rank"
	CSFieldTargetRank         = "target_rank"
	CSFieldTotalCredits       = "total_credits"
	CSFieldRequiredCredits    = "required_credits"
	CSFieldCreditGap          = "credit_gap"
	CSFieldLastPromotionDate  = "last_promotion_date"
	CSFieldNextEligibleDate   = "next_eligible_date"
	CSFieldSkpScore           = "skp_score"
	CSFieldNotes              = "notes"
)

// ── Certification enum constants ──────────────────────────────────────────────

const (
	CertTypeProfesi   = "profesi"
	CertTypePenilaian = "penilaian"
	CertTypePengawas  = "pengawas"
	CertTypeTeknisi   = "teknisi"
)

const (
	CertStatusActive          = "active"
	CertStatusExpired         = "expired"
	CertStatusRevoked         = "revoked"
	CertStatusPendingRenewal  = "pending_renewal"
)

// ── Activity enum constants ───────────────────────────────────────────────────

const (
	ActStatusRegistered = "registered"
	ActStatusAttended   = "attended"
	ActStatusCompleted  = "completed"
	ActStatusCancelled  = "cancelled"
)

// ── Credit summary enum constants ─────────────────────────────────────────────

const (
	RankIA  = "I/a"
	RankIB  = "I/b"
	RankIIA = "II/a"
	RankIIB = "II/b"
	RankIIIA = "III/a"
	RankIIIB = "III/b"
	RankIVA = "IV/a"
	RankIVB = "IV/b"
	RankIVC = "IV/c"
)

// ── Relation name constant ────────────────────────────────────────────────────

const (
	RelTeacher = "teacher"
)

var validCertTypes = map[string]bool{
	CertTypeProfesi: true, CertTypePenilaian: true, CertTypePengawas: true, CertTypeTeknisi: true,
}

var validCertStatuses = map[string]bool{
	CertStatusActive: true, CertStatusExpired: true, CertStatusRevoked: true, CertStatusPendingRenewal: true,
}

var validActivityStatuses = map[string]bool{
	ActStatusRegistered: true, ActStatusAttended: true, ActStatusCompleted: true, ActStatusCancelled: true,
}

var validRanks = map[string]bool{
	RankIA: true, RankIB: true, RankIIA: true, RankIIB: true,
	RankIIIA: true, RankIIIB: true, RankIVA: true, RankIVB: true, RankIVC: true,
}

// ── CertificationDescriptor ───────────────────────────────────────────────────

// CertificationDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_certifications.
type CertificationDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *CertificationDescriptor) TableName() string { return "teacher_certifications" }

// DefaultRels mendefinisikan relasi teacher_certifications: teacher.
func (d *CertificationDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         CertFieldTeacherID,
			LocalKey:   CertFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
	}
}

// Validate memvalidasi invariant teacher_certifications.
func (d *CertificationDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, CertFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	return requireEnum(data, CertFieldCertificationType, "certification_type", validCertTypes)
}

// ── ActivityDescriptor ────────────────────────────────────────────────────────

// ActivityDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_development_activities.
type ActivityDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ActivityDescriptor) TableName() string { return "teacher_development_activities" }

// DefaultRels mendefinisikan relasi teacher_development_activities: teacher.
func (d *ActivityDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         ActFieldTeacherID,
			LocalKey:   ActFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip"},
		},
	}
}

// Validate memvalidasi invariant teacher_development_activities.
func (d *ActivityDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, ActFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	if err := requireString(data, ActFieldTitle, "title"); err != nil {
		return err
	}
	return requireString(data, ActFieldStartDate, "start_date")
}

// ── CreditSummaryDescriptor ───────────────────────────────────────────────────

// CreditSummaryDescriptor mengimplementasi vernon.DomainDescriptor untuk teacher_credit_summaries.
type CreditSummaryDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *CreditSummaryDescriptor) TableName() string { return "teacher_credit_summaries" }

// DefaultRels mendefinisikan relasi teacher_credit_summaries: teacher.
func (d *CreditSummaryDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTeacher: {
			Domain:     "teachers",
			Type:       vernon.RelBelongsTo,
			FK:         CSFieldTeacherID,
			LocalKey:   CSFieldTeacherID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nip", "employee_type"},
		},
	}
}

// Validate memvalidasi invariant teacher_credit_summaries.
func (d *CreditSummaryDescriptor) Validate(data map[string]any) error {
	if err := requireString(data, CSFieldTeacherID, "teacher_id"); err != nil {
		return err
	}
	return requireEnum(data, CSFieldCurrentRank, "current_rank", validRanks)
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
