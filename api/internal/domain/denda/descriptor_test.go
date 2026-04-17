package denda_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/boilerplate/internal/domain/denda"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &denda.Descriptor{}
	assert.Equal(t, "denda", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &denda.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &denda.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "denda should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_PinjamanRelation(t *testing.T) {
	d := &denda.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[denda.RelPinjaman]
	require.True(t, ok, "should have pinjaman relation")

	assert.Equal(t, "pinjaman", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, denda.FieldPinjamanID, rel.FK)
	assert.True(t, rel.IsAutoload, "pinjaman should be autoloaded")
	assert.Equal(t, []string{"loan_number", "akad_type"}, rel.Fields)
}

func TestDescriptor_DefaultRels_RekeningRelation(t *testing.T) {
	d := &denda.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[denda.RelRekening]
	require.True(t, ok, "should have rekening relation")

	assert.Equal(t, "rekening", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, denda.FieldRekeningID, rel.FK)
	assert.True(t, rel.IsAutoload, "rekening should be autoloaded")
	assert.Equal(t, []string{"no_rekening", "category"}, rel.Fields)
}

// ── Validation tests — valid data ────────────────────────────────────────────

func validDendaData() map[string]any {
	return map[string]any{
		denda.FieldPinjamanID:       "00000000-0000-0000-0000-000000000001",
		denda.FieldRekeningID:       "00000000-0000-0000-0000-000000000002",
		denda.FieldPenaltyType:      denda.PenaltyLatePayment,
		denda.FieldCalculationBasis: float64(1000000),
		denda.FieldCalculatedAmount: float64(50000),
		denda.FieldFinalAmount:      float64(50000),
		denda.FieldOutstandingAmt:   float64(50000),
		denda.FieldPaidAmount:       float64(0),
		denda.FieldWaivedAmount:     float64(0),
		denda.FieldStatus:           denda.StatusAccruing,
		denda.FieldFundDestination:  denda.FundKoperasiIncome,
		denda.FieldPeriodStart:      "2024-01-01",
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &denda.Descriptor{}
	err := d.Validate(validDendaData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllPenaltyTypes(t *testing.T) {
	d := &denda.Descriptor{}
	types := []string{
		denda.PenaltyLatePayment, denda.PenaltyEarlySettlement,
		denda.PenaltyEarlyWithdrawal, denda.PenaltyTazir, denda.PenaltyTawidh,
	}
	for _, pt := range types {
		data := validDendaData()
		data[denda.FieldPenaltyType] = pt
		err := d.Validate(data)
		assert.NoError(t, err, "penalty_type=%q should be valid", pt)
	}
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &denda.Descriptor{}
	statuses := []string{
		denda.StatusAccruing, denda.StatusSettled,
		denda.StatusWaived, denda.StatusPartialWaived,
	}
	for _, status := range statuses {
		data := validDendaData()
		data[denda.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_AllFundDestinations(t *testing.T) {
	d := &denda.Descriptor{}
	destinations := []string{denda.FundKoperasiIncome, denda.FundSocialFund}
	for _, dest := range destinations {
		data := validDendaData()
		data[denda.FieldFundDestination] = dest
		err := d.Validate(data)
		assert.NoError(t, err, "fund_destination=%q should be valid", dest)
	}
}

func TestDescriptor_Validate_ZeroAmounts(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldCalculationBasis] = float64(0)
	data[denda.FieldCalculatedAmount] = float64(0)
	data[denda.FieldFinalAmount] = float64(0)
	data[denda.FieldOutstandingAmt] = float64(0)
	data[denda.FieldPaidAmount] = float64(0)
	data[denda.FieldWaivedAmount] = float64(0)
	err := d.Validate(data)
	assert.NoError(t, err, "zero amounts should be valid")
}

// ── Validation tests — missing required fields ───────────────────────────────

func TestDescriptor_Validate_MissingPinjamanID(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	delete(data, denda.FieldPinjamanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "pinjaman_id wajib diisi")
}

func TestDescriptor_Validate_MissingRekeningID(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	delete(data, denda.FieldRekeningID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rekening_id wajib diisi")
}

func TestDescriptor_Validate_MissingPenaltyType(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	delete(data, denda.FieldPenaltyType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "penalty_type wajib diisi")
}

func TestDescriptor_Validate_MissingPeriodStart(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	delete(data, denda.FieldPeriodStart)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period_start wajib diisi")
}

// ── Validation tests — invalid enums ─────────────────────────────────────────

func TestDescriptor_Validate_InvalidPenaltyType(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldPenaltyType] = "invalid_type"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "penalty_type tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldStatus] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_InvalidFundDestination(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldFundDestination] = "unknown"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "fund_destination tidak valid")
}

// ── Validation tests — negative amounts ───────────────────────────────────────

func TestDescriptor_Validate_NegativeCalculationBasis(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldCalculationBasis] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "calculation_basis tidak boleh negatif")
}

func TestDescriptor_Validate_NegativeCalculatedAmount(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldCalculatedAmount] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "calculated_amount tidak boleh negatif")
}

func TestDescriptor_Validate_NegativeFinalAmount(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldFinalAmount] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "final_amount tidak boleh negatif")
}

func TestDescriptor_Validate_NegativeOutstandingAmount(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldOutstandingAmt] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "outstanding_amount tidak boleh negatif")
}

func TestDescriptor_Validate_NegativePaidAmount(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldPaidAmount] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "paid_amount tidak boleh negatif")
}

func TestDescriptor_Validate_NegativeWaivedAmount(t *testing.T) {
	d := &denda.Descriptor{}
	data := validDendaData()
	data[denda.FieldWaivedAmount] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "waived_amount tidak boleh negatif")
}

// ── Validation tests — edge cases ────────────────────────────────────────────

func TestDescriptor_Validate_NilData(t *testing.T) {
	d := &denda.Descriptor{}
	err := d.Validate(nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "pinjaman_id wajib diisi")
}

func TestDescriptor_Validate_EmptyData(t *testing.T) {
	d := &denda.Descriptor{}
	err := d.Validate(map[string]any{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "pinjaman_id wajib diisi")
}

// ── Constant correctness ─────────────────────────────────────────────────────

func TestConstants_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, denda.PenaltyLatePayment)
	assert.NotEmpty(t, denda.PenaltyEarlySettlement)
	assert.NotEmpty(t, denda.PenaltyEarlyWithdrawal)
	assert.NotEmpty(t, denda.PenaltyTazir)
	assert.NotEmpty(t, denda.PenaltyTawidh)
	assert.NotEmpty(t, denda.StatusAccruing)
	assert.NotEmpty(t, denda.StatusSettled)
	assert.NotEmpty(t, denda.StatusWaived)
	assert.NotEmpty(t, denda.StatusPartialWaived)
	assert.NotEmpty(t, denda.FundKoperasiIncome)
	assert.NotEmpty(t, denda.FundSocialFund)
}
