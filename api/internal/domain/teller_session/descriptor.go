// Package teller_session adalah domain Vernon untuk sesi teller koperasi.
//
// TellerSession adalah entity root — tidak memiliki relasi belongs_to.
// Digunakan untuk tracking pembukaan dan penutupan sesi kasir/teller.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package teller_session

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldUserID      = "user_id"
	FieldSessionDate = "session_date"
	FieldStartTime   = "start_time"
	FieldEndTime     = "end_time"
	FieldOpeningBalance = "opening_balance"
	FieldClosingBalance = "closing_balance"
	FieldCashTotal   = "cash_total"
	FieldTransactionCount = "transaction_count"
	FieldStatus      = "status"
	FieldNotes       = "notes"
)

// Status constants.
const (
	StatusOpen    = "open"
	StatusClosed  = "closed"
	StatusReconciled = "reconciled"
)

var validStatuses = map[string]bool{
	StatusOpen: true, StatusClosed: true, StatusReconciled: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk teller_sessions.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "teller_sessions" }

// DefaultRels mendefinisikan relasi domain ini.
// TellerSession adalah entity root — tidak punya belongs_to.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	userID, _ := data[FieldUserID].(string)
	if userID == "" {
		return fmt.Errorf("user_id wajib diisi")
	}

	sessionDate, _ := data[FieldSessionDate].(string)
	if sessionDate == "" {
		return fmt.Errorf("session_date wajib diisi")
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
