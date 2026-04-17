package kas_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/kas"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &kas.Descriptor{}
	assert.Equal(t, "kas", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &kas.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &kas.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "kas should have 1 BelongsTo relation")
}

func TestDescriptor_DefaultRels_TellerSessionRelation(t *testing.T) {
	d := &kas.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[kas.RelTellerSession]
	require.True(t, ok, "should have teller_session relation")

	assert.Equal(t, "teller_sessions", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, kas.FieldTellerSessionID, rel.FK)
	assert.True(t, rel.IsAutoload, "teller_session should be autoloaded")
	assert.Equal(t, []string{"session_date", "status"}, rel.Fields)
}

// ── Validation tests ────────────────────────────────────────────────────────

func validKasData() map[string]any {
	return map[string]any{
		kas.FieldTellerSessionID: "00000000-0000-0000-0000-000000000001",
		kas.FieldAmount:          float64(250000),
		kas.FieldKasType:         kas.KasTypeMasuk,
		kas.FieldStatus:          kas.StatusPending,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &kas.Descriptor{}
	err := d.Validate(validKasData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_MissingTellerSessionID(t *testing.T) {
	d := &kas.Descriptor{}
	data := validKasData()
	delete(data, kas.FieldTellerSessionID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teller_session_id wajib diisi")
}

func TestDescriptor_Validate_MissingAmount(t *testing.T) {
	d := &kas.Descriptor{}
	data := validKasData()
	delete(data, kas.FieldAmount)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount wajib diisi dan harus berupa angka")
}

func TestDescriptor_Validate_InvalidKasType(t *testing.T) {
	d := &kas.Descriptor{}
	data := validKasData()
	data[kas.FieldKasType] = "transfer"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "kas_type tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &kas.Descriptor{}
	data := validKasData()
	data[kas.FieldStatus] = "cancelled"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_ZeroAmount(t *testing.T) {
	d := &kas.Descriptor{}
	data := validKasData()
	data[kas.FieldAmount] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih besar dari 0")
}

func TestDescriptor_Validate_NegativeAmount(t *testing.T) {
	d := &kas.Descriptor{}
	data := validKasData()
	data[kas.FieldAmount] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih besar dari 0")
}

func TestDescriptor_Validate_AllKasTypes(t *testing.T) {
	d := &kas.Descriptor{}
	types := []string{kas.KasTypeMasuk, kas.KasTypeKeluar}
	for _, typ := range types {
		data := validKasData()
		data[kas.FieldKasType] = typ
		err := d.Validate(data)
		assert.NoError(t, err, "kas_type=%q should be valid", typ)
	}
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &kas.Descriptor{}
	statuses := []string{kas.StatusPending, kas.StatusPosted, kas.StatusReversed}
	for _, status := range statuses {
		data := validKasData()
		data[kas.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestDescriptor_Validate_EmptyKasType_OK(t *testing.T) {
	d := &kas.Descriptor{}
	data := validKasData()
	data[kas.FieldKasType] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty kas_type should be valid (optional)")
}

func TestDescriptor_Validate_EmptyStatus_OK(t *testing.T) {
	d := &kas.Descriptor{}
	data := validKasData()
	data[kas.FieldStatus] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty status should be valid (optional)")
}
