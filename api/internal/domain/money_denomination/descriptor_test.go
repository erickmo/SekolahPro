package money_denomination_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/money_denomination"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &money_denomination.Descriptor{}
	assert.Equal(t, "money_denominations", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &money_denomination.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &money_denomination.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "money_denomination should have 1 BelongsTo relation")
}

func TestDescriptor_DefaultRels_TellerSessionRelation(t *testing.T) {
	d := &money_denomination.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[money_denomination.RelTellerSession]
	require.True(t, ok, "should have teller_session relation")

	assert.Equal(t, "teller_sessions", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, money_denomination.FieldTellerSessionID, rel.FK)
	assert.True(t, rel.IsAutoload, "teller_session should be autoloaded")
	assert.Equal(t, []string{"session_date", "status"}, rel.Fields)
}

// ── Validation tests ────────────────────────────────────────────────────────

func validMoneyDenominationData() map[string]any {
	return map[string]any{
		money_denomination.FieldTellerSessionID: "00000000-0000-0000-0000-000000000001",
		money_denomination.FieldDenomination:    float64(100000),
		money_denomination.FieldCount:           float64(5),
		money_denomination.FieldTotal:           float64(500000),
		money_denomination.FieldType:            money_denomination.TypeOpening,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &money_denomination.Descriptor{}
	err := d.Validate(validMoneyDenominationData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_MissingTellerSessionID(t *testing.T) {
	d := &money_denomination.Descriptor{}
	data := validMoneyDenominationData()
	delete(data, money_denomination.FieldTellerSessionID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "teller_session_id wajib diisi")
}

func TestDescriptor_Validate_MissingDenomination(t *testing.T) {
	d := &money_denomination.Descriptor{}
	data := validMoneyDenominationData()
	delete(data, money_denomination.FieldDenomination)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "denomination wajib diisi dan harus berupa angka")
}

func TestDescriptor_Validate_InvalidType(t *testing.T) {
	d := &money_denomination.Descriptor{}
	data := validMoneyDenominationData()
	data[money_denomination.FieldType] = "banknote"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "type tidak valid")
}

func TestDescriptor_Validate_ZeroDenomination(t *testing.T) {
	d := &money_denomination.Descriptor{}
	data := validMoneyDenominationData()
	data[money_denomination.FieldDenomination] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "denomination harus lebih besar dari 0")
}

func TestDescriptor_Validate_NegativeDenomination(t *testing.T) {
	d := &money_denomination.Descriptor{}
	data := validMoneyDenominationData()
	data[money_denomination.FieldDenomination] = float64(-5000)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "denomination harus lebih besar dari 0")
}

func TestDescriptor_Validate_AllTypes(t *testing.T) {
	d := &money_denomination.Descriptor{}
	types := []string{money_denomination.TypeOpening, money_denomination.TypeClosing}
	for _, typ := range types {
		data := validMoneyDenominationData()
		data[money_denomination.FieldType] = typ
		err := d.Validate(data)
		assert.NoError(t, err, "type=%q should be valid", typ)
	}
}

func TestDescriptor_Validate_EmptyType_OK(t *testing.T) {
	d := &money_denomination.Descriptor{}
	data := validMoneyDenominationData()
	data[money_denomination.FieldType] = ""
	err := d.Validate(data)
	assert.NoError(t, err, "empty type should be valid (optional)")
}
