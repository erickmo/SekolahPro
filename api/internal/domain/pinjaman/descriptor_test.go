package pinjaman_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/pinjaman"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &pinjaman.Descriptor{}
	assert.Equal(t, "pinjaman", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &pinjaman.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &pinjaman.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 3, "pinjaman should have 3 BelongsTo relations")
}

func TestDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &pinjaman.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[pinjaman.RelNasabah]
	require.True(t, ok, "should have nasabah relation")

	assert.Equal(t, "nasabah", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, pinjaman.FieldNasabahID, rel.FK)
	assert.True(t, rel.IsAutoload, "nasabah should be autoloaded")
	assert.Equal(t, []string{"nama_lengkap", "no_nasabah", "type"}, rel.Fields)
}

func TestDescriptor_DefaultRels_RekeningRelation(t *testing.T) {
	d := &pinjaman.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[pinjaman.RelRekening]
	require.True(t, ok, "should have rekening relation")

	assert.Equal(t, "rekening", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, pinjaman.FieldRekeningID, rel.FK)
	assert.True(t, rel.IsAutoload, "rekening should be autoloaded")
	assert.Equal(t, []string{"no_rekening", "category"}, rel.Fields)
}

func TestDescriptor_DefaultRels_ProdukAkadRelation(t *testing.T) {
	d := &pinjaman.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[pinjaman.RelProdukAkad]
	require.True(t, ok, "should have produk_akad relation")

	assert.Equal(t, "produk_akad", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, pinjaman.FieldProdukAkadID, rel.FK)
	assert.True(t, rel.IsAutoload, "produk_akad should be autoloaded")
	assert.Equal(t, []string{"name", "akad_type"}, rel.Fields)
}

// ── Validation tests ────────────────────────────────────────────────────────

func validPinjamanData() map[string]any {
	return map[string]any{
		pinjaman.FieldNasabahID:       "00000000-0000-0000-0000-000000000001",
		pinjaman.FieldRekeningID:      "00000000-0000-0000-0000-000000000002",
		pinjaman.FieldProdukAkadID:    "00000000-0000-0000-0000-000000000003",
		pinjaman.FieldPrincipalAmount: float64(10000000),
		pinjaman.FieldTenorMonths:     float64(12),
		pinjaman.FieldInterestRate:    float64(12.5),
		pinjaman.FieldStatus:          pinjaman.StatusPending,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &pinjaman.Descriptor{}
	err := d.Validate(validPinjamanData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &pinjaman.Descriptor{}
	statuses := []string{
		pinjaman.StatusPending, pinjaman.StatusApproved, pinjaman.StatusActive,
		pinjaman.StatusPaidOff, pinjaman.StatusRejected, pinjaman.StatusWrittenOff,
		pinjaman.StatusOverdue,
	}
	for _, status := range statuses {
		data := validPinjamanData()
		data[pinjaman.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_MissingNasabahID(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	delete(data, pinjaman.FieldNasabahID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

func TestDescriptor_Validate_MissingRekeningID(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	delete(data, pinjaman.FieldRekeningID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rekening_id wajib diisi")
}

func TestDescriptor_Validate_MissingProdukAkadID(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	delete(data, pinjaman.FieldProdukAkadID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "produk_akad_id wajib diisi")
}

func TestDescriptor_Validate_MissingPrincipalAmount(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	delete(data, pinjaman.FieldPrincipalAmount)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "principal_amount wajib diisi")
}

func TestDescriptor_Validate_MissingTenorMonths(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	delete(data, pinjaman.FieldTenorMonths)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tenor_months wajib diisi")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	data[pinjaman.FieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_InvalidDisbursementMethod(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	data[pinjaman.FieldDisbursementMethod] = "cheque"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "disbursement_method tidak valid")
}

func TestDescriptor_Validate_NegativePrincipalAmount(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	data[pinjaman.FieldPrincipalAmount] = float64(-5000000)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "principal_amount harus lebih dari 0")
}

func TestDescriptor_Validate_ZeroPrincipalAmount(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	data[pinjaman.FieldPrincipalAmount] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "principal_amount harus lebih dari 0")
}

func TestDescriptor_Validate_TenorMonthsBelow1(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	data[pinjaman.FieldTenorMonths] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "tenor_months minimal 1")
}

func TestDescriptor_Validate_NegativeInterestRate(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	data[pinjaman.FieldInterestRate] = float64(-1.5)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "interest_rate tidak boleh negatif")
}

func TestDescriptor_Validate_ZeroInterestRate_OK(t *testing.T) {
	d := &pinjaman.Descriptor{}
	data := validPinjamanData()
	data[pinjaman.FieldInterestRate] = float64(0)
	err := d.Validate(data)
	assert.NoError(t, err, "interest_rate=0 should be valid")
}
