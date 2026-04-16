package deposito_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/deposito"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &deposito.Descriptor{}
	assert.Equal(t, "deposito", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &deposito.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &deposito.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "deposito should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &deposito.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[deposito.RelNasabah]
	require.True(t, ok, "should have nasabah relation")

	assert.Equal(t, "nasabah", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "nasabah should be autoloaded")
}

func TestDescriptor_DefaultRels_RekeningRelation(t *testing.T) {
	d := &deposito.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[deposito.RelRekening]
	require.True(t, ok, "should have rekening relation")

	assert.Equal(t, "rekening", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "rekening should be autoloaded")
}

// ── Validation tests ────────────────────────────────────────────────────────

func validDepositoData() map[string]any {
	return map[string]any{
		deposito.FieldNasabahID:   "00000000-0000-0000-0000-000000000001",
		deposito.FieldRekeningID:  "00000000-0000-0000-0000-000000000002",
		deposito.FieldPrincipal:   float64(10000000),
		deposito.FieldTenorMonths: float64(12),
		deposito.FieldRate:        float64(5.5),
		deposito.FieldStatus:      deposito.StatusActive,
		deposito.FieldPaymentMethod: deposito.PaymentMonthly,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &deposito.Descriptor{}
	err := d.Validate(validDepositoData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllValidTenors(t *testing.T) {
	d := &deposito.Descriptor{}
	tenors := []int{1, 3, 6, 12, 24}
	for _, tenor := range tenors {
		data := validDepositoData()
		data[deposito.FieldTenorMonths] = float64(tenor)
		err := d.Validate(data)
		assert.NoError(t, err, "tenor=%d should be valid", tenor)
	}
}

func TestDescriptor_Validate_MissingNasabahID(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	delete(data, deposito.FieldNasabahID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

func TestDescriptor_Validate_MissingRekeningID(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	delete(data, deposito.FieldRekeningID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rekening_id wajib diisi")
}

func TestDescriptor_Validate_MissingPrincipal(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	delete(data, deposito.FieldPrincipal)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "principal wajib diisi")
}

func TestDescriptor_Validate_NegativePrincipal(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	data[deposito.FieldPrincipal] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "principal harus lebih dari 0")
}

func TestDescriptor_Validate_ZeroPrincipal(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	data[deposito.FieldPrincipal] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "principal harus lebih dari 0")
}

func TestDescriptor_Validate_MissingTenor(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	delete(data, deposito.FieldTenorMonths)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tenor_months wajib diisi")
}

func TestDescriptor_Validate_InvalidTenor(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	data[deposito.FieldTenorMonths] = float64(7)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tenor_months tidak valid")
}

func TestDescriptor_Validate_MissingRate(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	delete(data, deposito.FieldRate)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rate wajib diisi")
}

func TestDescriptor_Validate_NegativeRate(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	data[deposito.FieldRate] = float64(-1)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rate tidak boleh negatif")
}

func TestDescriptor_Validate_ZeroRateAllowed(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	data[deposito.FieldRate] = float64(0)
	err := d.Validate(data)
	assert.NoError(t, err, "rate=0 should be valid (profit sharing mode)")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	data[deposito.FieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_InvalidPaymentMethod(t *testing.T) {
	d := &deposito.Descriptor{}
	data := validDepositoData()
	data[deposito.FieldPaymentMethod] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "payment_method tidak valid")
}

func TestDescriptor_Validate_AllStatusCombinations(t *testing.T) {
	d := &deposito.Descriptor{}
	statuses := []string{
		deposito.StatusActive, deposito.StatusMatured,
		deposito.StatusRolledOver, deposito.StatusEarlyWithdraw,
		deposito.StatusClosed,
	}
	for _, status := range statuses {
		data := validDepositoData()
		data[deposito.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_AllPaymentMethods(t *testing.T) {
	d := &deposito.Descriptor{}
	methods := []string{deposito.PaymentMonthly, deposito.PaymentMaturity, deposito.PaymentCompound}
	for _, method := range methods {
		data := validDepositoData()
		data[deposito.FieldPaymentMethod] = method
		err := d.Validate(data)
		assert.NoError(t, err, "payment_method=%q should be valid", method)
	}
}
