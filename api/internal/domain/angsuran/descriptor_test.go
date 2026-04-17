package angsuran_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/angsuran"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &angsuran.Descriptor{}
	assert.Equal(t, "angsuran", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &angsuran.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &angsuran.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "angsuran should have 1 BelongsTo relation")
}

func TestDescriptor_DefaultRels_PinjamanRelation(t *testing.T) {
	d := &angsuran.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[angsuran.RelPinjaman]
	require.True(t, ok, "should have pinjaman relation")

	assert.Equal(t, "pinjaman", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, angsuran.FieldPinjamanID, rel.FK)
	assert.True(t, rel.IsAutoload, "pinjaman should be autoloaded")
	assert.Equal(t, []string{"loan_number", "akad_type", "status"}, rel.Fields)
}

// ── Validation tests ────────────────────────────────────────────────────────

func validAngsuranData() map[string]any {
	return map[string]any{
		angsuran.FieldPinjamanID:        "00000000-0000-0000-0000-000000000001",
		angsuran.FieldInstallmentNumber: float64(1),
		angsuran.FieldDueDate:           "2026-05-17",
		angsuran.FieldPrincipalAmount:   float64(833333),
		angsuran.FieldInterestAmount:    float64(104167),
		angsuran.FieldTotalAmount:       float64(937500),
		angsuran.FieldStatus:            angsuran.StatusScheduled,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &angsuran.Descriptor{}
	err := d.Validate(validAngsuranData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &angsuran.Descriptor{}
	statuses := []string{
		angsuran.StatusScheduled, angsuran.StatusPaid, angsuran.StatusPartial,
		angsuran.StatusOverdue, angsuran.StatusWaived,
	}
	for _, status := range statuses {
		data := validAngsuranData()
		data[angsuran.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_MissingPinjamanID(t *testing.T) {
	d := &angsuran.Descriptor{}
	data := validAngsuranData()
	delete(data, angsuran.FieldPinjamanID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "pinjaman_id wajib diisi")
}

func TestDescriptor_Validate_MissingInstallmentNumber(t *testing.T) {
	d := &angsuran.Descriptor{}
	data := validAngsuranData()
	delete(data, angsuran.FieldInstallmentNumber)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "installment_number wajib diisi")
}

func TestDescriptor_Validate_InstallmentNumberBelow1(t *testing.T) {
	d := &angsuran.Descriptor{}
	data := validAngsuranData()
	data[angsuran.FieldInstallmentNumber] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "installment_number minimal 1")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &angsuran.Descriptor{}
	data := validAngsuranData()
	data[angsuran.FieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_NegativePrincipalAmount(t *testing.T) {
	d := &angsuran.Descriptor{}
	data := validAngsuranData()
	data[angsuran.FieldPrincipalAmount] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "principal_amount tidak boleh negatif")
}

func TestDescriptor_Validate_NegativeInterestAmount(t *testing.T) {
	d := &angsuran.Descriptor{}
	data := validAngsuranData()
	data[angsuran.FieldInterestAmount] = float64(-50)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "interest_amount tidak boleh negatif")
}

func TestDescriptor_Validate_NegativeTotalAmount(t *testing.T) {
	d := &angsuran.Descriptor{}
	data := validAngsuranData()
	data[angsuran.FieldTotalAmount] = float64(-150)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "total_amount tidak boleh negatif")
}

func TestDescriptor_Validate_ZeroAmounts_OK(t *testing.T) {
	d := &angsuran.Descriptor{}
	data := validAngsuranData()
	data[angsuran.FieldPrincipalAmount] = float64(0)
	data[angsuran.FieldInterestAmount] = float64(0)
	data[angsuran.FieldTotalAmount] = float64(0)
	err := d.Validate(data)
	assert.NoError(t, err, "zero amounts should be valid")
}
