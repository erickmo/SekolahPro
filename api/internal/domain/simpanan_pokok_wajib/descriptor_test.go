package simpanan_pokok_wajib_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/simpanan_pokok_wajib"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	assert.Equal(t, "simpanan_pokok_wajib", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &simpanan_pokok_wajib.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "simpanan_pokok_wajib should have 2 BelongsTo relations")
}

func TestDescriptor_DefaultRels_NasabahRelation(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[simpanan_pokok_wajib.RelNasabah]
	require.True(t, ok, "should have nasabah relation")

	assert.Equal(t, "nasabah", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, simpanan_pokok_wajib.FieldNasabahID, rel.FK)
	assert.True(t, rel.IsAutoload, "nasabah should be autoloaded")
}

func TestDescriptor_DefaultRels_RekeningRelation(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[simpanan_pokok_wajib.RelRekening]
	require.True(t, ok, "should have rekening relation")

	assert.Equal(t, "rekening", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, simpanan_pokok_wajib.FieldRekeningID, rel.FK)
	assert.True(t, rel.IsAutoload, "rekening should be autoloaded")
}

// ── Validation tests ────────────────────────────────────────────────────────

func validSimpananData() map[string]any {
	return map[string]any{
		simpanan_pokok_wajib.FieldNasabahID:  "00000000-0000-0000-0000-000000000001",
		simpanan_pokok_wajib.FieldRekeningID: "00000000-0000-0000-0000-000000000002",
		simpanan_pokok_wajib.FieldType:       simpanan_pokok_wajib.TypePokok,
		simpanan_pokok_wajib.FieldAmount:     float64(500000),
		simpanan_pokok_wajib.FieldStatus:     simpanan_pokok_wajib.StatusPending,
	}
}

func TestDescriptor_Validate_Success_Pokok(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	err := d.Validate(validSimpananData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_Success_Wajib(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	data[simpanan_pokok_wajib.FieldType] = simpanan_pokok_wajib.TypeWajib
	data[simpanan_pokok_wajib.FieldPeriod] = "2026-04"
	err := d.Validate(data)
	assert.NoError(t, err)
}

func TestDescriptor_Validate_MissingNasabahID(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	delete(data, simpanan_pokok_wajib.FieldNasabahID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "nasabah_id wajib diisi")
}

func TestDescriptor_Validate_MissingRekeningID(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	delete(data, simpanan_pokok_wajib.FieldRekeningID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rekening_id wajib diisi")
}

func TestDescriptor_Validate_MissingType(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	delete(data, simpanan_pokok_wajib.FieldType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "type wajib diisi")
}

func TestDescriptor_Validate_MissingAmount(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	delete(data, simpanan_pokok_wajib.FieldAmount)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount wajib diisi")
}

func TestDescriptor_Validate_NegativeAmount(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	data[simpanan_pokok_wajib.FieldAmount] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih dari 0")
}

func TestDescriptor_Validate_ZeroAmount(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	data[simpanan_pokok_wajib.FieldAmount] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih dari 0")
}

func TestDescriptor_Validate_InvalidType(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	data[simpanan_pokok_wajib.FieldType] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "type tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	data[simpanan_pokok_wajib.FieldStatus] = "invalid"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_WajibMissingPeriod(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	data := validSimpananData()
	data[simpanan_pokok_wajib.FieldType] = simpanan_pokok_wajib.TypeWajib
	// period not set
	err := d.Validate(data)
	assert.ErrorContains(t, err, "period wajib diisi untuk simpanan wajib")
}

func TestDescriptor_Validate_AllStatusCombinations(t *testing.T) {
	d := &simpanan_pokok_wajib.Descriptor{}
	statuses := []string{
		simpanan_pokok_wajib.StatusPending,
		simpanan_pokok_wajib.StatusPaid,
		simpanan_pokok_wajib.StatusOverdue,
		simpanan_pokok_wajib.StatusRefunded,
	}
	for _, status := range statuses {
		data := validSimpananData()
		data[simpanan_pokok_wajib.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}
