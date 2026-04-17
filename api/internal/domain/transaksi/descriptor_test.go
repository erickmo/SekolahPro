package transaksi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/transaksi"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ── Descriptor metadata tests ───────────────────────────────────────────────

func TestDescriptor_TableName(t *testing.T) {
	d := &transaksi.Descriptor{}
	assert.Equal(t, "transaksi", d.TableName())
}

func TestDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &transaksi.Descriptor{}
}

// ── DefaultRels tests ────────────────────────────────────────────────────────

func TestDescriptor_DefaultRels_Count(t *testing.T) {
	d := &transaksi.Descriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 1, "transaksi should have 1 BelongsTo relation")
}

func TestDescriptor_DefaultRels_RekeningRelation(t *testing.T) {
	d := &transaksi.Descriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[transaksi.RelRekening]
	require.True(t, ok, "should have rekening relation")

	assert.Equal(t, "rekening", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, transaksi.FieldRekeningID, rel.FK)
	assert.True(t, rel.IsAutoload, "rekening should be autoloaded")
	assert.Equal(t, []string{"no_rekening", "category"}, rel.Fields)
}

// ── Validation tests ────────────────────────────────────────────────────────

func validTransaksiData() map[string]any {
	return map[string]any{
		transaksi.FieldRekeningID:     "00000000-0000-0000-0000-000000000001",
		transaksi.FieldTransactionType: transaksi.TypeCredit,
		transaksi.FieldAmount:         float64(50000),
		transaksi.FieldStatus:         transaksi.StatusPending,
	}
}

func TestDescriptor_Validate_Success(t *testing.T) {
	d := &transaksi.Descriptor{}
	err := d.Validate(validTransaksiData())
	assert.NoError(t, err)
}

func TestDescriptor_Validate_MissingRekeningID(t *testing.T) {
	d := &transaksi.Descriptor{}
	data := validTransaksiData()
	delete(data, transaksi.FieldRekeningID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "rekening_id wajib diisi")
}

func TestDescriptor_Validate_MissingTransactionType(t *testing.T) {
	d := &transaksi.Descriptor{}
	data := validTransaksiData()
	delete(data, transaksi.FieldTransactionType)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "transaction_type wajib diisi")
}

func TestDescriptor_Validate_MissingAmount(t *testing.T) {
	d := &transaksi.Descriptor{}
	data := validTransaksiData()
	delete(data, transaksi.FieldAmount)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount wajib diisi dan harus berupa angka")
}

func TestDescriptor_Validate_InvalidTransactionType(t *testing.T) {
	d := &transaksi.Descriptor{}
	data := validTransaksiData()
	data[transaksi.FieldTransactionType] = "refund"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "transaction_type tidak valid")
}

func TestDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &transaksi.Descriptor{}
	data := validTransaksiData()
	data[transaksi.FieldStatus] = "cancelled"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

func TestDescriptor_Validate_ZeroAmount(t *testing.T) {
	d := &transaksi.Descriptor{}
	data := validTransaksiData()
	data[transaksi.FieldAmount] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih besar dari 0")
}

func TestDescriptor_Validate_NegativeAmount(t *testing.T) {
	d := &transaksi.Descriptor{}
	data := validTransaksiData()
	data[transaksi.FieldAmount] = float64(-100)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "amount harus lebih besar dari 0")
}

func TestDescriptor_Validate_AllTransactionTypes(t *testing.T) {
	d := &transaksi.Descriptor{}
	types := []string{transaksi.TypeCredit, transaksi.TypeDebit, transaksi.TypeTransfer}
	for _, typ := range types {
		data := validTransaksiData()
		data[transaksi.FieldTransactionType] = typ
		err := d.Validate(data)
		assert.NoError(t, err, "transaction_type=%q should be valid", typ)
	}
}

func TestDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &transaksi.Descriptor{}
	statuses := []string{transaksi.StatusPending, transaksi.StatusPosted, transaksi.StatusReversed}
	for _, status := range statuses {
		data := validTransaksiData()
		data[transaksi.FieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}
