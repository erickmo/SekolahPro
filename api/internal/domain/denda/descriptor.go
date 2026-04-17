// Package denda adalah domain Vernon untuk denda & penalti koperasi.
//
// Denda mencatat penalti yang dikenakan pada nasabah: keterlambatan,
// pelunasan awal, penarikan awal, tazir (denda syariah), tawidh (kompensasi).
// Lifecycle: accruing → settled / waived / partial_waived.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package denda

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// Field name constants.
const (
	FieldPinjamanID       = "pinjaman_id"
	FieldRekeningID       = "rekening_id"
	FieldPenaltyType      = "penalty_type"
	FieldCalculationBasis = "calculation_basis"
	FieldCalculatedAmount = "calculated_amount"
	FieldFinalAmount      = "final_amount"
	FieldOutstandingAmt   = "outstanding_amount"
	FieldPaidAmount       = "paid_amount"
	FieldWaivedAmount     = "waived_amount"
	FieldStatus           = "status"
	FieldFundDestination  = "fund_destination"
	FieldPeriodStart      = "period_start"
)

// Penalty type constants.
const (
	PenaltyLatePayment    = "late_payment"
	PenaltyEarlySettlement = "early_settlement"
	PenaltyEarlyWithdrawal = "early_withdrawal"
	PenaltyTazir          = "tazir"
	PenaltyTawidh         = "tawidh"
)

// Status constants.
const (
	StatusAccruing     = "accruing"
	StatusSettled      = "settled"
	StatusWaived       = "waived"
	StatusPartialWaived = "partial_waived"
)

// Fund destination constants.
const (
	FundKoperasiIncome = "koperasi_income"
	FundSocialFund     = "social_fund"
)

// Relation name constants.
const (
	RelPinjaman = "pinjaman"
	RelRekening = "rekening"
)

// validPenaltyTypes berisi semua penalty_type yang valid.
var validPenaltyTypes = map[string]bool{
	PenaltyLatePayment: true, PenaltyEarlySettlement: true,
	PenaltyEarlyWithdrawal: true, PenaltyTazir: true, PenaltyTawidh: true,
}

// validStatuses berisi semua status yang valid.
var validStatuses = map[string]bool{
	StatusAccruing: true, StatusSettled: true,
	StatusWaived: true, StatusPartialWaived: true,
}

// validFundDestinations berisi semua fund_destination yang valid.
var validFundDestinations = map[string]bool{
	FundKoperasiIncome: true, FundSocialFund: true,
}

// Descriptor mengimplementasi vernon.DomainDescriptor untuk denda.
type Descriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *Descriptor) TableName() string { return "denda" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo pinjaman (autoload) dan rekening (autoload).
func (d *Descriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		RelPinjaman: {
			Domain:     "pinjaman",
			Type:       vernon.RelBelongsTo,
			FK:         FieldPinjamanID,
			LocalKey:   FieldPinjamanID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"loan_number", "akad_type"},
		},
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
	if err := validateRequired(data); err != nil {
		return err
	}
	if err := validateEnums(data); err != nil {
		return err
	}
	return validateAmounts(data)
}

// validateRequired memeriksa field wajib.
func validateRequired(data map[string]any) error {
	pinjamanID, _ := data[FieldPinjamanID].(string)
	if pinjamanID == "" {
		return errors.New("pinjaman_id wajib diisi")
	}

	rekeningID, _ := data[FieldRekeningID].(string)
	if rekeningID == "" {
		return errors.New("rekening_id wajib diisi")
	}

	penaltyType, _ := data[FieldPenaltyType].(string)
	if penaltyType == "" {
		return errors.New("penalty_type wajib diisi")
	}

	periodStart, _ := data[FieldPeriodStart].(string)
	if periodStart == "" {
		return errors.New("period_start wajib diisi")
	}
	return nil
}

// validateEnums memeriksa enum fields.
func validateEnums(data map[string]any) error {
	penaltyType, _ := data[FieldPenaltyType].(string)
	if penaltyType != "" && !validPenaltyTypes[penaltyType] {
		return fmt.Errorf("penalty_type tidak valid: %q", penaltyType)
	}

	status, _ := data[FieldStatus].(string)
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %q (harus accruing/settled/waived/partial_waived)", status)
	}

	fundDest, _ := data[FieldFundDestination].(string)
	if fundDest != "" && !validFundDestinations[fundDest] {
		return fmt.Errorf("fund_destination tidak valid: %q (harus koperasi_income/social_fund)", fundDest)
	}
	return nil
}

// validateAmounts memeriksa field amount >= 0.
func validateAmounts(data map[string]any) error {
	amountFields := []struct {
		field string
		label string
	}{
		{FieldCalculationBasis, "calculation_basis"},
		{FieldCalculatedAmount, "calculated_amount"},
		{FieldFinalAmount, "final_amount"},
		{FieldOutstandingAmt, "outstanding_amount"},
		{FieldPaidAmount, "paid_amount"},
		{FieldWaivedAmount, "waived_amount"},
	}

	for _, af := range amountFields {
		val, ok := data[af.field].(float64)
		if ok && val < 0 {
			return fmt.Errorf("%s tidak boleh negatif", af.label)
		}
	}
	return nil
}
