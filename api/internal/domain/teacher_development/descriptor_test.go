package teacher_development_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/teacher_development"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── CertificationDescriptor ───────────────────────────────────────────────────

func TestCertificationDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "teacher_certifications", (&teacher_development.CertificationDescriptor{}).TableName())
}

func TestCertificationDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_development.CertificationDescriptor{}
}

func TestCertificationDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&teacher_development.CertificationDescriptor{}).DefaultRels(), 1)
}

func TestCertificationDescriptor_DefaultRels_Teacher(t *testing.T) {
	rel, ok := (&teacher_development.CertificationDescriptor{}).DefaultRels()[teacher_development.RelTeacher]
	require.True(t, ok)
	assert.Equal(t, "teachers", rel.Domain)
	assert.True(t, rel.IsAutoload)
}

func validCertificationData() map[string]any {
	return map[string]any{
		teacher_development.CertFieldTeacherID:         "00000000-0000-0000-0000-000000000001",
		teacher_development.CertFieldCertificationType: teacher_development.CertTypeProfesi,
	}
}

func TestCertificationDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&teacher_development.CertificationDescriptor{}).Validate(validCertificationData()))
}

func TestCertificationDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := validCertificationData()
	delete(d, teacher_development.CertFieldTeacherID)
	assert.ErrorContains(t, (&teacher_development.CertificationDescriptor{}).Validate(d), "teacher_id wajib diisi")
}

func TestCertificationDescriptor_Validate_InvalidType(t *testing.T) {
	d := validCertificationData()
	d[teacher_development.CertFieldCertificationType] = "magang"
	assert.ErrorContains(t, (&teacher_development.CertificationDescriptor{}).Validate(d), "certification_type tidak valid")
}

// ── ActivityDescriptor ────────────────────────────────────────────────────────

func TestActivityDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "teacher_development_activities", (&teacher_development.ActivityDescriptor{}).TableName())
}

func TestActivityDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_development.ActivityDescriptor{}
}

func TestActivityDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&teacher_development.ActivityDescriptor{}).DefaultRels(), 1)
}

func validActivityData() map[string]any {
	return map[string]any{
		teacher_development.ActFieldTeacherID:   "00000000-0000-0000-0000-000000000001",
		teacher_development.ActFieldTitle:       "Workshop Kurikulum Merdeka",
		teacher_development.ActFieldStartDate:   "2026-04-10",
	}
}

func TestActivityDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&teacher_development.ActivityDescriptor{}).Validate(validActivityData()))
}

func TestActivityDescriptor_Validate_MissingTitle(t *testing.T) {
	d := validActivityData()
	delete(d, teacher_development.ActFieldTitle)
	assert.ErrorContains(t, (&teacher_development.ActivityDescriptor{}).Validate(d), "title wajib diisi")
}

func TestActivityDescriptor_Validate_MissingStartDate(t *testing.T) {
	d := validActivityData()
	delete(d, teacher_development.ActFieldStartDate)
	assert.ErrorContains(t, (&teacher_development.ActivityDescriptor{}).Validate(d), "start_date wajib diisi")
}

// ── CreditSummaryDescriptor ───────────────────────────────────────────────────

func TestCreditSummaryDescriptor_TableName(t *testing.T) {
	assert.Equal(t, "teacher_credit_summaries", (&teacher_development.CreditSummaryDescriptor{}).TableName())
}

func TestCreditSummaryDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &teacher_development.CreditSummaryDescriptor{}
}

func TestCreditSummaryDescriptor_DefaultRels_Count(t *testing.T) {
	assert.Len(t, (&teacher_development.CreditSummaryDescriptor{}).DefaultRels(), 1)
}

func validCreditSummaryData() map[string]any {
	return map[string]any{
		teacher_development.CSFieldTeacherID:   "00000000-0000-0000-0000-000000000001",
		teacher_development.CSFieldCurrentRank: teacher_development.RankIIIA,
	}
}

func TestCreditSummaryDescriptor_Validate_Success(t *testing.T) {
	assert.NoError(t, (&teacher_development.CreditSummaryDescriptor{}).Validate(validCreditSummaryData()))
}

func TestCreditSummaryDescriptor_Validate_MissingTeacherID(t *testing.T) {
	d := validCreditSummaryData()
	delete(d, teacher_development.CSFieldTeacherID)
	assert.ErrorContains(t, (&teacher_development.CreditSummaryDescriptor{}).Validate(d), "teacher_id wajib diisi")
}

func TestCreditSummaryDescriptor_Validate_InvalidRank(t *testing.T) {
	d := validCreditSummaryData()
	d[teacher_development.CSFieldCurrentRank] = "V/a"
	assert.ErrorContains(t, (&teacher_development.CreditSummaryDescriptor{}).Validate(d), "current_rank tidak valid")
}
