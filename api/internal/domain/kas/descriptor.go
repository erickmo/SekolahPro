// Package kas adalah domain Vernon untuk kas koperasi sekolah.
//
// Kas memiliki 1 BelongsTo autoload: teller_session.
// Digunakan untuk tracking mutasi kas per sesi teller.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package kas

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldTellerSessionID = "teller_session_id"
	FieldKasType         = "kas_type"
	FieldAmount          = "amount"
	FieldDescription     = "description"
	FieldReferenceNo     = "reference_no"
	FieldPostedDate      = "posted_date"
	FieldStatus          = "status"
)

// Kas type constants.
const (
	KasTypeMasuk = "masuk"
	KasTypeKeluar = "keluar"
)

// Status constants.
const (
	StatusPending  = "pending"
	StatusPosted   = "posted"
	StatusReversed = "reversed"
)

// Relation name constants.
const (
	RelTellerSession = "teller_session"
)

var validKasTypes = map[string]bool{
	KasTypeMasuk: true, KasTypeKeluar: true,
}

var validStatuses = map[string]bool{
	StatusPending: true, StatusPosted: true, StatusReversed: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk kas.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "kas" }

// DefaultRels mendefinisikan relasi domain ini.
// 1 BelongsTo autoload: teller_session.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelTellerSession: {
			Domain:     "teller_sessions",
			Type:       vernon.RelBelongsTo,
			FK:         FieldTellerSessionID,
			LocalKey:   FieldTellerSessionID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"session_date", "status"},
		},
	}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	tellerSessionID, _ := data[FieldTellerSessionID].(string)
	if tellerSessionID == "" {
		return fmt.Errorf("teller_session_id wajib diisi")
	}

	amount, ok := data[FieldAmount].(float64)
	if !ok {
		return fmt.Errorf("amount wajib diisi dan harus berupa angka")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih besar dari 0")
	}

	kasType, _ := data[FieldKasType].(string)
	if kasType != "" && !validKasTypes[kasType] {
		return fmt.Errorf("kas_type tidak valid: %q (harus masuk/keluar)", kasType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
