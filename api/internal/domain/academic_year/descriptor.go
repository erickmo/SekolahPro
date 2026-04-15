// Package academic_year adalah domain Vernon untuk tahun ajaran (ADR-010).
//
// Tahun ajaran adalah unit waktu fundamental — referenced by 10+ student domains.
// Ini adalah entity root (tidak punya belongs_to ke domain lain).
// Lifecycle: planning → active → closed.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package academic_year

import (
	"errors"
	"fmt"
	"time"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldName           = "name"
	FieldCode           = "code"
	FieldStartDate      = "start_date"
	FieldEndDate        = "end_date"
	FieldSemester1Start = "semester1_start"
	FieldSemester1End   = "semester1_end"
	FieldSemester2Start = "semester2_start"
	FieldSemester2End   = "semester2_end"
	FieldIsActive       = "is_active"
	FieldStatus         = "status"
)

// Status constants.
const (
	StatusPlanning = "planning"
	StatusActive   = "active"
	StatusClosed   = "closed"
)

// validStatuses berisi semua status yang valid.
var validStatuses = map[string]bool{
	StatusPlanning: true,
	StatusActive:   true,
	StatusClosed:   true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk academic_years.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "academic_years" }

// DefaultRels mendefinisikan relasi domain ini.
// academic_years adalah entity root — tidak punya belongs_to.
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant domain sebelum write.
func (d *Descriptor) Validate(data map[string]any) error {
	if err := validateRequired(data); err != nil {
		return err
	}
	if err := validateStatus(data); err != nil {
		return err
	}
	return validateDates(data)
}

// validateRequired memeriksa field wajib.
func validateRequired(data map[string]any) error {
	name, _ := data[FieldName].(string)
	if name == "" {
		return errors.New("name wajib diisi")
	}
	code, _ := data[FieldCode].(string)
	if code == "" {
		return errors.New("code wajib diisi")
	}
	return nil
}

// validateStatus memeriksa status valid.
func validateStatus(data map[string]any) error {
	status, _ := data[FieldStatus].(string)
	if status == "" {
		return nil // default akan di-set oleh service layer
	}
	if !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus planning/active/closed)", status)
	}
	return nil
}

// validateDates memeriksa konsistensi tanggal.
func validateDates(data map[string]any) error {
	startDate, err1 := parseDate(data, FieldStartDate)
	endDate, err2 := parseDate(data, FieldEndDate)
	if err1 != nil || err2 != nil {
		return nil // date fields optional saat validasi (bisa diisi nanti)
	}
	if !startDate.Before(endDate) {
		return errors.New("start_date harus sebelum end_date")
	}

	s1Start, err1 := parseDate(data, FieldSemester1Start)
	s1End, err2 := parseDate(data, FieldSemester1End)
	if err1 == nil && err2 == nil && !s1Start.Before(s1End) {
		return errors.New("semester1_start harus sebelum semester1_end")
	}

	s2Start, err1 := parseDate(data, FieldSemester2Start)
	s2End, err2 := parseDate(data, FieldSemester2End)
	if err1 == nil && err2 == nil && !s2Start.Before(s2End) {
		return errors.New("semester2_start harus sebelum semester2_end")
	}

	if s1End.IsZero() || s2Start.IsZero() {
		return nil
	}
	if s1End.After(s2Start) {
		return errors.New("semester1_end harus sebelum atau sama dengan semester2_start")
	}
	return nil
}

// parseDate memparse tanggal dari _data field (format: "2006-01-02").
func parseDate(data map[string]any, field string) (time.Time, error) {
	raw, _ := data[field].(string)
	if raw == "" {
		return time.Time{}, fmt.Errorf("%s kosong", field)
	}
	return time.Parse("2006-01-02", raw)
}
