// Package money_denomination adalah domain Vernon untuk pecahan uang di sesi teller.
//
// MoneyDenomination memiliki 1 BelongsTo autoload: teller_session.
// Digunakan untuk tracking detail pecahan uang saat buka/tutup sesi kasir.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package money_denomination

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldTellerSessionID = "teller_session_id"
	FieldDenomination    = "denomination"
	FieldCount           = "count"
	FieldTotal           = "total"
	FieldType            = "type"
)

// Denomination type constants.
const (
	TypeOpening = "opening"
	TypeClosing = "closing"
)

// Relation name constants.
const (
	RelTellerSession = "teller_session"
)

var validTypes = map[string]bool{
	TypeOpening: true, TypeClosing: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk money_denominations.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "money_denominations" }

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

	denomination, ok := data[FieldDenomination].(float64)
	if !ok {
		return fmt.Errorf("denomination wajib diisi dan harus berupa angka")
	}
	if denomination <= 0 {
		return fmt.Errorf("denomination harus lebih besar dari 0")
	}

	typ, _ := data[FieldType].(string)
	if typ != "" && !validTypes[typ] {
		return fmt.Errorf("type tidak valid: %q (harus opening/closing)", typ)
	}
	return nil
}
