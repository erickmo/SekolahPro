package zakat_infaq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/zakat_infaq"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── ZakatCollectionDescriptor metadata tests ────────────────────────────────

func TestZakatCollectionDescriptor_TableName(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	assert.Equal(t, "zakat_collection", d.TableName())
}

func TestZakatCollectionDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &zakat_infaq.ZakatCollectionDescriptor{}
}

func TestZakatCollectionDescriptor_DefaultRels_Count(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "zakat_collection should have 2 BelongsTo relations")
}

func TestZakatCollectionDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[zakat_infaq.ZCRelNasabah]
	require.True(t, ok, "should have nasabah relation")
	assert.Equal(t, "nasabah", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, zakat_infaq.ZCFieldNasabahID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

func TestZakatCollectionDescriptor_DefaultRels_BranchRelation(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[zakat_infaq.ZCRelBranch]
	require.True(t, ok, "should have branch relation")
	assert.Equal(t, "branches", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload)
}

// ── ZakatCollectionDescriptor validation tests ──────────────────────────────

func validZakatCollectionData() map[string]any {
	return map[string]any{
		zakat_infaq.ZCFieldMuzakkiName:   "Ahmad Fauzi",
		zakat_infaq.ZCFieldZakatType:     zakat_infaq.ZakatTypeMal,
		zakat_infaq.ZCFieldAmount:        float64(3187500),
		zakat_infaq.ZCFieldPaymentMethod: zakat_infaq.PaymentMethodCash,
		zakat_infaq.ZCFieldStatus:        zakat_infaq.ZCStatusPending,
	}
}

func TestZakatCollectionDescriptor_Validate_Success(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	err := d.Validate(validZakatCollectionData())
	assert.NoError(t, err)
}

func TestZakatCollectionDescriptor_Validate_MissingMuzakkiName(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	data := validZakatCollectionData()
	delete(data, zakat_infaq.ZCFieldMuzakkiName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "muzakki_name wajib diisi")
}

func TestZakatCollectionDescriptor_Validate_MissingZakatType(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	data := validZakatCollectionData()
	delete(data, zakat_infaq.ZCFieldZakatType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "zakat_type wajib diisi")
}

func TestZakatCollectionDescriptor_Validate_MissingAmount(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	data := validZakatCollectionData()
	delete(data, zakat_infaq.ZCFieldAmount)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount wajib diisi")
}

func TestZakatCollectionDescriptor_Validate_NegativeAmount(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	data := validZakatCollectionData()
	data[zakat_infaq.ZCFieldAmount] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih dari 0")
}

func TestZakatCollectionDescriptor_Validate_InvalidZakatType(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	data := validZakatCollectionData()
	data[zakat_infaq.ZCFieldZakatType] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "zakat_type tidak valid")
}

func TestZakatCollectionDescriptor_Validate_InvalidPaymentMethod(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	data := validZakatCollectionData()
	data[zakat_infaq.ZCFieldPaymentMethod] = "cheque"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "payment_method tidak valid")
}

func TestZakatCollectionDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	data := validZakatCollectionData()
	data[zakat_infaq.ZCFieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestZakatCollectionDescriptor_Validate_AllZakatTypes(t *testing.T) {
	d := &zakat_infaq.ZakatCollectionDescriptor{}
	types := []string{
		zakat_infaq.ZakatTypeMal, zakat_infaq.ZakatTypeFitrah,
		zakat_infaq.ZakatTypeInstitusi,
	}
	for _, zt := range types {
		data := validZakatCollectionData()
		data[zakat_infaq.ZCFieldZakatType] = zt
		err := d.Validate(data)
		assert.NoError(t, err, "zakat_type=%q should be valid", zt)
	}
}

// ── ZakatDistributionDescriptor metadata tests ──────────────────────────────

func TestZakatDistributionDescriptor_TableName(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	assert.Equal(t, "zakat_distribution", d.TableName())
}

func TestZakatDistributionDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &zakat_infaq.ZakatDistributionDescriptor{}
}

func TestZakatDistributionDescriptor_DefaultRels_Count(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "zakat_distribution should have 2 BelongsTo relations")
}

func TestZakatDistributionDescriptor_DefaultRels_MustahikRelation(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[zakat_infaq.ZDRelMustahik]
	require.True(t, ok, "should have mustahik relation")
	assert.Equal(t, "mustahik", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, zakat_infaq.ZDFieldMustahikID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

// ── ZakatDistributionDescriptor validation tests ────────────────────────────

func validZakatDistributionData() map[string]any {
	return map[string]any{
		zakat_infaq.ZDFieldMustahikID:         "00000000-0000-0000-0000-000000000001",
		zakat_infaq.ZDFieldSourceFund:         zakat_infaq.SourceFundZakat,
		zakat_infaq.ZDFieldAmount:             float64(1000000),
		zakat_infaq.ZDFieldPurpose:            "Bantuan sembako",
		zakat_infaq.ZDFieldAsnafCategory:      zakat_infaq.AsnafFakir,
		zakat_infaq.ZDFieldStatus:             zakat_infaq.ZDStatusDraft,
		zakat_infaq.ZDFieldDistributionMethod: zakat_infaq.DistributionMethodCash,
	}
}

func TestZakatDistributionDescriptor_Validate_Success(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	err := d.Validate(validZakatDistributionData())
	assert.NoError(t, err)
}

func TestZakatDistributionDescriptor_Validate_MissingMustahikID(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	data := validZakatDistributionData()
	delete(data, zakat_infaq.ZDFieldMustahikID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "mustahik_id wajib diisi")
}

func TestZakatDistributionDescriptor_Validate_MissingSourceFund(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	data := validZakatDistributionData()
	delete(data, zakat_infaq.ZDFieldSourceFund)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "source_fund wajib diisi")
}

func TestZakatDistributionDescriptor_Validate_MissingPurpose(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	data := validZakatDistributionData()
	delete(data, zakat_infaq.ZDFieldPurpose)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "purpose wajib diisi")
}

func TestZakatDistributionDescriptor_Validate_InvalidSourceFund(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	data := validZakatDistributionData()
	data[zakat_infaq.ZDFieldSourceFund] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "source_fund tidak valid")
}

func TestZakatDistributionDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	data := validZakatDistributionData()
	data[zakat_infaq.ZDFieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestZakatDistributionDescriptor_Validate_InvalidDistributionMethod(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	data := validZakatDistributionData()
	data[zakat_infaq.ZDFieldDistributionMethod] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "distribution_method tidak valid")
}

func TestZakatDistributionDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &zakat_infaq.ZakatDistributionDescriptor{}
	statuses := []string{
		zakat_infaq.ZDStatusDraft, zakat_infaq.ZDStatusPending,
		zakat_infaq.ZDStatusApproved, zakat_infaq.ZDStatusRejected,
		zakat_infaq.ZDStatusCompleted,
	}
	for _, status := range statuses {
		data := validZakatDistributionData()
		data[zakat_infaq.ZDFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

// ── InfaqDescriptor metadata tests ──────────────────────────────────────────

func TestInfaqDescriptor_TableName(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	assert.Equal(t, "infaq", d.TableName())
}

func TestInfaqDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &zakat_infaq.InfaqDescriptor{}
}

func TestInfaqDescriptor_DefaultRels_Count(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "infaq should have 2 BelongsTo relations")
}

// ── InfaqDescriptor validation tests ────────────────────────────────────────

func validInfaqData() map[string]any {
	return map[string]any{
		zakat_infaq.IFieldDonaturName:   "Siti Aminah",
		zakat_infaq.IFieldInfaqType:     zakat_infaq.InfaqTypeOneTime,
		zakat_infaq.IFieldAmount:        float64(50000),
		zakat_infaq.IFieldDesignation:   "pendidikan",
		zakat_infaq.IFieldPaymentMethod: zakat_infaq.PaymentMethodCash,
	}
}

func TestInfaqDescriptor_Validate_Success(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	err := d.Validate(validInfaqData())
	assert.NoError(t, err)
}

func TestInfaqDescriptor_Validate_MissingDonaturName(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	data := validInfaqData()
	delete(data, zakat_infaq.IFieldDonaturName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "donatur_name wajib diisi")
}

func TestInfaqDescriptor_Validate_MissingInfaqType(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	data := validInfaqData()
	delete(data, zakat_infaq.IFieldInfaqType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "infaq_type wajib diisi")
}

func TestInfaqDescriptor_Validate_MissingDesignation(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	data := validInfaqData()
	delete(data, zakat_infaq.IFieldDesignation)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "designation wajib diisi")
}

func TestInfaqDescriptor_Validate_InvalidInfaqType(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	data := validInfaqData()
	data[zakat_infaq.IFieldInfaqType] = "monthly"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "infaq_type tidak valid")
}

func TestInfaqDescriptor_Validate_ZeroAmount(t *testing.T) {
	d := &zakat_infaq.InfaqDescriptor{}
	data := validInfaqData()
	data[zakat_infaq.IFieldAmount] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih dari 0")
}

// ── MustahikDescriptor metadata tests ───────────────────────────────────────

func TestMustahikDescriptor_TableName(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	assert.Equal(t, "mustahik", d.TableName())
}

func TestMustahikDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &zakat_infaq.MustahikDescriptor{}
}

func TestMustahikDescriptor_DefaultRels_Count(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "mustahik should have 1 optional BelongsTo (branch)")
}

// ── MustahikDescriptor validation tests ─────────────────────────────────────

func validMustahikData() map[string]any {
	return map[string]any{
		zakat_infaq.MFieldFullName:        "Siti Aminah",
		zakat_infaq.MFieldAsnafCategories: []any{"fakir"},
		zakat_infaq.MFieldPrimaryCategory: zakat_infaq.AsnafFakir,
		zakat_infaq.MFieldNeedsAssessment: "Tidak memiliki penghasilan tetap",
		zakat_infaq.MFieldAssessmentDate:  "2026-01-15",
		zakat_infaq.MFieldAssessedBy:      "00000000-0000-0000-0000-000000000099",
		zakat_infaq.MFieldStatus:          zakat_infaq.MStatusActive,
	}
}

func TestMustahikDescriptor_Validate_Success(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	err := d.Validate(validMustahikData())
	assert.NoError(t, err)
}

func TestMustahikDescriptor_Validate_MissingFullName(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	data := validMustahikData()
	delete(data, zakat_infaq.MFieldFullName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "full_name wajib diisi")
}

func TestMustahikDescriptor_Validate_MissingAsnafCategories(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	data := validMustahikData()
	delete(data, zakat_infaq.MFieldAsnafCategories)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "asnaf_categories wajib diisi")
}

func TestMustahikDescriptor_Validate_InvalidPrimaryCategory(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	data := validMustahikData()
	data[zakat_infaq.MFieldPrimaryCategory] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "primary_category tidak valid")
}

func TestMustahikDescriptor_Validate_MissingNeedsAssessment(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	data := validMustahikData()
	delete(data, zakat_infaq.MFieldNeedsAssessment)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "needs_assessment wajib diisi")
}

func TestMustahikDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	data := validMustahikData()
	data[zakat_infaq.MFieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestMustahikDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &zakat_infaq.MustahikDescriptor{}
	statuses := []string{
		zakat_infaq.MStatusActive, zakat_infaq.MStatusInactive,
		zakat_infaq.MStatusGraduated,
	}
	for _, status := range statuses {
		data := validMustahikData()
		data[zakat_infaq.MFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

// ── TazirFundDescriptor metadata tests ──────────────────────────────────────

func TestTazirFundDescriptor_TableName(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	assert.Equal(t, "tazir_fund", d.TableName())
}

func TestTazirFundDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &zakat_infaq.TazirFundDescriptor{}
}

func TestTazirFundDescriptor_DefaultRels_Count(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 3, "tazir_fund should have 3 BelongsTo relations")
}

func TestTazirFundDescriptor_DefaultRels_DendaRelation(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[zakat_infaq.TFRelDenda]
	require.True(t, ok, "should have denda relation")
	assert.Equal(t, "denda", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, zakat_infaq.TFFieldDendaID, rel.FK)
	assert.True(t, rel.IsAutoload)
}

func TestTazirFundDescriptor_DefaultRels_PinjamanRelation(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[zakat_infaq.TFRelPinjaman]
	require.True(t, ok, "should have pinjaman relation")
	assert.Equal(t, "pinjaman", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload)
}

func TestTazirFundDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[zakat_infaq.TFRelNasabah]
	require.True(t, ok, "should have nasabah relation")
	assert.Equal(t, "nasabah", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload)
}

// ── TazirFundDescriptor validation tests ────────────────────────────────────

func validTazirFundData() map[string]any {
	return map[string]any{
		zakat_infaq.TFFieldDendaID:    "00000000-0000-0000-0000-000000000001",
		zakat_infaq.TFFieldPinjamanID: "00000000-0000-0000-0000-000000000002",
		zakat_infaq.TFFieldNasabahID:  "00000000-0000-0000-0000-000000000003",
		zakat_infaq.TFFieldAmount:     float64(5000),
		zakat_infaq.TFFieldStatus:     zakat_infaq.TFStatusCollected,
	}
}

func TestTazirFundDescriptor_Validate_Success(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	err := d.Validate(validTazirFundData())
	assert.NoError(t, err)
}

func TestTazirFundDescriptor_Validate_MissingDendaID(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	data := validTazirFundData()
	delete(data, zakat_infaq.TFFieldDendaID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "denda_id wajib diisi")
}

func TestTazirFundDescriptor_Validate_MissingPinjamanID(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	data := validTazirFundData()
	delete(data, zakat_infaq.TFFieldPinjamanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "pinjaman_id wajib diisi")
}

func TestTazirFundDescriptor_Validate_MissingNasabahID(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	data := validTazirFundData()
	delete(data, zakat_infaq.TFFieldNasabahID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

func TestTazirFundDescriptor_Validate_MissingAmount(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	data := validTazirFundData()
	delete(data, zakat_infaq.TFFieldAmount)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount wajib diisi")
}

func TestTazirFundDescriptor_Validate_ZeroAmount(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	data := validTazirFundData()
	data[zakat_infaq.TFFieldAmount] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih dari 0")
}

func TestTazirFundDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &zakat_infaq.TazirFundDescriptor{}
	data := validTazirFundData()
	data[zakat_infaq.TFFieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}
