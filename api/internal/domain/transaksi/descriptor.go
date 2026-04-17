// Package transaksi adalah domain Vernon untuk transaksi koperasi sekolah.
//
// Transaksi memiliki 1 BelongsTo autoload: rekening.
// Digunakan untuk tracking semua transaksi kredit dan debet rekening koperasi.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package transaksi

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldRekeningID     = "rekening_id"
	FieldTransactionType = "transaction_type"
	FieldAmount         = "amount"
	FieldDescription    = "description"
	FieldReferenceNo    = "reference_no"
	FieldBalanceBefore  = "balance_before"
	FieldBalanceAfter   = "balance_after"
	FieldPostedDate     = "posted_date"
	FieldStatus         = "status"
)

// Transaction type constants.
const (
	TypeCredit  = "credit"
	TypeDebit   = "debit"
	TypeTransfer = "transfer"
)

// Status constants.
const (
	StatusPending  = "pending"
	StatusPosted   = "posted"
	StatusReversed = "reversed"
)

// Relation name constants.
const (
	RelRekening = "rekening"
)

var validTransactionTypes = map[string]bool{
	TypeCredit: true, TypeDebit: true, TypeTransfer: true,
}

var validStatuses = map[string]bool{
	StatusPending: true, StatusPosted: true, StatusReversed: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk transaksi.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "transaksi" }

// DefaultRels mendefinisikan relasi domain ini.
// 1 BelongsTo autoload: rekening.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelRekening: {
			Domain:     "rekening",
			Type:       vernon.RelBelongsTo,
			FK:         FieldRekeningID,
			LocalKey:   FieldRekeningID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"no_rekening", "category"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	rekeningID, _ := data[FieldRekeningID].(string)
	if rekeningID == "" {
		return fmt.Errorf("rekening_id wajib diisi")
	}

	transactionType, _ := data[FieldTransactionType].(string)
	if transactionType == "" {
		return fmt.Errorf("transaction_type wajib diisi")
	}
	if !validTransactionTypes[transactionType] {
		return fmt.Errorf("transaction_type tidak valid: %q", transactionType)
	}

	amount, ok := data[FieldAmount].(float64)
	if !ok {
		return fmt.Errorf("amount wajib diisi dan harus berupa angka")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih besar dari 0")
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
