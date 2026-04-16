package tabungan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/tabungan"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &tabungan.Descriptor{}
	assert.Equal(t, "tabungan", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &tabungan.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &tabungan.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "tabungan should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &tabungan.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[tabungan.RelNasabah]
	require.True(t, ok, "should have nasabah relation")

	assert.Equal(t, "nasabah", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "nasabah should be autoloaded")
}

func TestDescriptor_DefaultRels_RekeningRelation(t *testing.T) {
	d := &tabungan.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[tabungan.RelRekening]
	require.True(t, ok, "should have rekening relation")

	assert.Equal(t, "rekening", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.True(t, rel.IsAutoload, "rekening should be autoloaded")
}

// ── Validation tests ────────────────────────────────────────────────────────

func validTabunganData() map[string]any {
	return map[string]any{
		tabungan.FieldNasabahID:   "00000000-0000-0000-0000-000000000001",
		tabungan.FieldRekeningID:  "00000000-0000-0000-0000-000000000002",
		tabungan.FieldProductType: tabungan.ProductRegular,
		tabungan.FieldBalance:     float64(1000000),
		tabungan.FieldHoldBalance: float64(0),
		tabungan.FieldStatus:      tabungan.StatusActive,
	}
}

func TestDescriptor_Validate_Success_Regular(t *testing.T) {
	d := &tabungan.Descriptor{}
	err := d.Validate(validTabunganData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_Success_AllProductTypes(t *testing.T) {
	d := &tabungan.Descriptor{}
	types := []string{
		tabungan.ProductRegular, tabungan.ProductEducation,
		tabungan.ProductHoliday, tabungan.ProductQurban, tabungan.ProductGoal,
	}
	for _, pt := range types {
		data := validTabunganData()
		data[tabungan.FieldProductType] = pt
		if pt == tabungan.ProductGoal {
			data[tabungan.FieldGoalAmount] = float64(5000000)
		}
		err := d.Validate(data)
		assert.NoError(t, err, "product_type=%q should be valid", pt)
	}
}

func TestDescriptor_Validate_MissingNasabahID(t *testing.T) {
	d := &tabungan.Descriptor{}
	data := validTabunganData()
	delete(data, tabungan.FieldNasabahID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

func TestDescriptor_Validate_MissingRekeningID(t *testing.T) {
	d := &tabungan.Descriptor{}
	data := validTabunganData()
	delete(data, tabungan.FieldRekeningID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rekening_id wajib diisi")
}

func TestDescriptor_Validate_MissingProductType(t *testing.T) {
	d := &tabungan.Descriptor{}
	data := validTabunganData()
	delete(data, tabungan.FieldProductType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "product_type wajib diisi")
}

func TestDescriptor_Validate_InvalidProductType(t *testing.T) {
	d := &tabungan.Descriptor{}
	data := validTabunganData()
	data[tabungan.FieldProductType] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "product_type tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &tabungan.Descriptor{}
	data := validTabunganData()
	data[tabungan.FieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_NegativeBalance(t *testing.T) {
	d := &tabungan.Descriptor{}
	data := validTabunganData()
	data[tabungan.FieldBalance] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "balance tidak boleh negatif")
}

func TestDescriptor_Validate_HoldExceedsBalance(t *testing.T) {
	d := &tabungan.Descriptor{}
	data := validTabunganData()
	data[tabungan.FieldBalance] = float64(100)
	data[tabungan.FieldHoldBalance] = float64(200)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "hold_balance tidak boleh lebih besar dari balance")
}

func TestDescriptor_Validate_GoalAmountOnNonGoalProduct(t *testing.T) {
	d := &tabungan.Descriptor{}
	data := validTabunganData()
	data[tabungan.FieldProductType] = tabungan.ProductRegular
	data[tabungan.FieldGoalAmount] = float64(5000000)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "goal_amount hanya untuk product_type goal")
}

func TestDescriptor_Validate_AllStatusCombinations(t *testing.T) {
	d := &tabungan.Descriptor{}
	statuses := []string{
		tabungan.StatusActive, tabungan.StatusDormant,
		tabungan.StatusFrozen, tabungan.StatusClosed,
	}
	for _, status := range statuses {
		data := validTabunganData()
		data[tabungan.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}
