// Package zakat_infaq adalah domain Vernon untuk pengelolaan zakat dan infaq.
//
// Modul ini hanya aktif untuk koperasi dengan coop_type = "islamic" (BMT).
// Meliputi: zakat_collection, zakat_distribution, infaq, mustahik, tazir_fund.
// Dua tabel non-Vernon (infaq_recurring_config, zakat_collection_batch)
// hanya ada di migration tanpa descriptor.
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
//   - coop_type check dilakukan di handler level, BUKAN di descriptor.
package zakat_infaq

import (
	"errors"
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// zakat_collection — Field name constants
// ============================================================================

const (
	ZCFieldNasabahID      = "nasabah_id"
	ZCFieldBranchID       = "branch_id"
	ZCFieldBatchID        = "batch_id"
	ZCFieldMuzakkiName    = "muzakki_name"
	ZCFieldZakatType      = "zakat_type"
	ZCFieldAmount         = "amount"
	ZCFieldPaymentMethod  = "payment_method"
	ZCFieldFitrahHeadCount = "fitrah_head_count"
	ZCFieldCollectionDate = "collection_date"
	ZCFieldConfirmedAt    = "confirmed_at"
	ZCFieldConfirmedBy    = "confirmed_by"
	ZCFieldNotes          = "notes"
	ZCFieldStatus         = "status"
)

// zakat_collection — Enum constants.
const (
	ZakatTypeMal        = "zakat_mal"
	ZakatTypeFitrah     = "zakat_fitrah"
	ZakatTypeInstitusi  = "zakat_institusi"

	PaymentMethodCash     = "cash"
	PaymentMethodAutoDebit = "auto_debit"
	PaymentMethodTransfer  = "transfer"

	ZCStatusPending   = "pending"
	ZCStatusConfirmed = "confirmed"
	ZCStatusCancelled = "cancelled"
)

// zakat_collection — Relation name constants.
const (
	ZCRelNasabah = "nasabah"
	ZCRelBranch  = "branch"
)

var validZakatTypes = map[string]bool{
	ZakatTypeMal: true, ZakatTypeFitrah: true, ZakatTypeInstitusi: true,
}

var validPaymentMethods = map[string]bool{
	PaymentMethodCash: true, PaymentMethodAutoDebit: true, PaymentMethodTransfer: true,
}

var validZCStatuses = map[string]bool{
	ZCStatusPending: true, ZCStatusConfirmed: true, ZCStatusCancelled: true,
}

// ZakatCollectionDescriptor mengimplementasi vernon.DomainDescriptor untuk zakat_collection.
type ZakatCollectionDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ZakatCollectionDescriptor) TableName() string { return "zakat_collection" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo nasabah dan branch (autoload).
func (d *ZakatCollectionDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		ZCRelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         ZCFieldNasabahID,
			LocalKey:   ZCFieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap", "no_nasabah"},
		},
		ZCRelBranch: {
			Domain:     "branches",
			Type:       vernon.RelBelongsTo,
			FK:         ZCFieldBranchID,
			LocalKey:   ZCFieldBranchID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant zakat_collection sebelum write.
func (d *ZakatCollectionDescriptor) Validate(data map[string]any) error {
	if err := validateZCRequired(data); err != nil {
		return err
	}
	return validateZCEnums(data)
}

// validateZCRequired memeriksa field wajib zakat_collection.
func validateZCRequired(data map[string]any) error {
	muzakkiName, _ := data[ZCFieldMuzakkiName].(string)
	if muzakkiName == "" {
		return errors.New("muzakki_name wajib diisi")
	}

	zakatType, _ := data[ZCFieldZakatType].(string)
	if zakatType == "" {
		return errors.New("zakat_type wajib diisi")
	}

	amount, ok := data[ZCFieldAmount].(float64)
	if !ok {
		return errors.New("amount wajib diisi")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih dari 0, got %.2f", amount)
	}

	paymentMethod, _ := data[ZCFieldPaymentMethod].(string)
	if paymentMethod == "" {
		return errors.New("payment_method wajib diisi")
	}
	return nil
}

// validateZCEnums memeriksa enum fields zakat_collection.
func validateZCEnums(data map[string]any) error {
	zakatType, _ := data[ZCFieldZakatType].(string)
	if zakatType != "" && !validZakatTypes[zakatType] {
		return fmt.Errorf("zakat_type tidak valid: %q", zakatType)
	}

	paymentMethod, _ := data[ZCFieldPaymentMethod].(string)
	if paymentMethod != "" && !validPaymentMethods[paymentMethod] {
		return fmt.Errorf("payment_method tidak valid: %q", paymentMethod)
	}

	status, _ := data[ZCFieldStatus].(string)
	if status != "" && !validZCStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}

// ============================================================================
// zakat_distribution — Field name constants
// ============================================================================

const (
	ZDFieldMustahikID         = "mustahik_id"
	ZDFieldBranchID           = "branch_id"
	ZDFieldSourceFund         = "source_fund"
	ZDFieldAmount             = "amount"
	ZDFieldPurpose            = "purpose"
	ZDFieldAsnafCategory      = "asnaf_category"
	ZDFieldDistributionDate   = "distribution_date"
	ZDFieldApprovedAt         = "approved_at"
	ZDFieldApprovedBy         = "approved_by"
	ZDFieldProofURL           = "proof_url"
	ZDFieldNotes              = "notes"
	ZDFieldStatus             = "status"
	ZDFieldDistributionMethod = "distribution_method"
)

// zakat_distribution — Enum constants.
const (
	SourceFundZakat  = "zakat"
	SourceFundInfaq  = "infaq"
	SourceFundTazir  = "tazir"

	ZDStatusDraft     = "draft"
	ZDStatusPending   = "pending"
	ZDStatusApproved  = "approved"
	ZDStatusRejected  = "rejected"
	ZDStatusCompleted = "completed"

	DistributionMethodCash           = "cash"
	DistributionMethodGoods          = "goods"
	DistributionMethodAccountTransfer = "account_transfer"
)

// zakat_distribution — Relation name constants.
const (
	ZDRelMustahik = "mustahik"
	ZDRelBranch   = "branch"
)

// Asnaf category constants.
const (
	AsnafFakir        = "fakir"
	AsnafMiskin       = "miskin"
	AsnafAmil         = "amil"
	AsnafMuallaf      = "muallaf"
	AsnafRiqab        = "riqab"
	AsnafGharimin     = "gharimin"
	AsnafFisabilillah = "fisabilillah"
	AsnafIbnuSabil    = "ibnu_sabil"
)

var validSourceFunds = map[string]bool{
	SourceFundZakat: true, SourceFundInfaq: true, SourceFundTazir: true,
}

var validZDStatuses = map[string]bool{
	ZDStatusDraft: true, ZDStatusPending: true,
	ZDStatusApproved: true, ZDStatusRejected: true, ZDStatusCompleted: true,
}

var validDistributionMethods = map[string]bool{
	DistributionMethodCash: true, DistributionMethodGoods: true,
	DistributionMethodAccountTransfer: true,
}

var validAsnafCategories = map[string]bool{
	AsnafFakir: true, AsnafMiskin: true, AsnafAmil: true,
	AsnafMuallaf: true, AsnafRiqab: true, AsnafGharimin: true,
	AsnafFisabilillah: true, AsnafIbnuSabil: true,
}

// ZakatDistributionDescriptor mengimplementasi vernon.DomainDescriptor untuk zakat_distribution.
type ZakatDistributionDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *ZakatDistributionDescriptor) TableName() string { return "zakat_distribution" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo mustahik dan branch (autoload).
func (d *ZakatDistributionDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		ZDRelMustahik: {
			Domain:     "mustahik",
			Type:       vernon.RelBelongsTo,
			FK:         ZDFieldMustahikID,
			LocalKey:   ZDFieldMustahikID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "primary_category"},
		},
		ZDRelBranch: {
			Domain:     "branches",
			Type:       vernon.RelBelongsTo,
			FK:         ZDFieldBranchID,
			LocalKey:   ZDFieldBranchID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant zakat_distribution sebelum write.
func (d *ZakatDistributionDescriptor) Validate(data map[string]any) error {
	if err := validateZDRequired(data); err != nil {
		return err
	}
	return validateZDEnums(data)
}

// validateZDRequired memeriksa field wajib zakat_distribution.
func validateZDRequired(data map[string]any) error {
	mustahikID, _ := data[ZDFieldMustahikID].(string)
	if mustahikID == "" {
		return errors.New("mustahik_id wajib diisi")
	}

	sourceFund, _ := data[ZDFieldSourceFund].(string)
	if sourceFund == "" {
		return errors.New("source_fund wajib diisi")
	}

	amount, ok := data[ZDFieldAmount].(float64)
	if !ok {
		return errors.New("amount wajib diisi")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih dari 0, got %.2f", amount)
	}

	purpose, _ := data[ZDFieldPurpose].(string)
	if purpose == "" {
		return errors.New("purpose wajib diisi")
	}
	return nil
}

// validateZDEnums memeriksa enum fields zakat_distribution.
func validateZDEnums(data map[string]any) error {
	sourceFund, _ := data[ZDFieldSourceFund].(string)
	if sourceFund != "" && !validSourceFunds[sourceFund] {
		return fmt.Errorf("source_fund tidak valid: %q", sourceFund)
	}

	status, _ := data[ZDFieldStatus].(string)
	if status != "" && !validZDStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}

	method, _ := data[ZDFieldDistributionMethod].(string)
	if method != "" && !validDistributionMethods[method] {
		return fmt.Errorf("distribution_method tidak valid: %q", method)
	}

	asnaf, _ := data[ZDFieldAsnafCategory].(string)
	if asnaf != "" && !validAsnafCategories[asnaf] {
		return fmt.Errorf("asnaf_category tidak valid: %q", asnaf)
	}
	return nil
}

// ============================================================================
// infaq — Field name constants
// ============================================================================

const (
	IFieldNasabahID     = "nasabah_id"
	IFieldBranchID      = "branch_id"
	IFieldDonaturName   = "donatur_name"
	IFieldInfaqType     = "infaq_type"
	IFieldAmount        = "amount"
	IFieldDesignation   = "designation"
	IFieldPaymentMethod = "payment_method"
	IFieldInfaqDate     = "infaq_date"
	IFieldRecurringCfgID = "recurring_config_id"
	IFieldNotes         = "notes"
)

// infaq — Enum constants.
const (
	InfaqTypeOneTime  = "one_time"
	InfaqTypeRecurring = "recurring"
)

// infaq — Relation name constants.
const (
	IRelNasabah = "nasabah"
	IRelBranch  = "branch"
)

var validInfaqTypes = map[string]bool{
	InfaqTypeOneTime: true, InfaqTypeRecurring: true,
}

// InfaqDescriptor mengimplementasi vernon.DomainDescriptor untuk infaq.
type InfaqDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *InfaqDescriptor) TableName() string { return "infaq" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo nasabah dan branch (autoload).
func (d *InfaqDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		IRelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         IFieldNasabahID,
			LocalKey:   IFieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap", "no_nasabah"},
		},
		IRelBranch: {
			Domain:     "branches",
			Type:       vernon.RelBelongsTo,
			FK:         IFieldBranchID,
			LocalKey:   IFieldBranchID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant infaq sebelum write.
func (d *InfaqDescriptor) Validate(data map[string]any) error {
	donaturName, _ := data[IFieldDonaturName].(string)
	if donaturName == "" {
		return errors.New("donatur_name wajib diisi")
	}

	infaqType, _ := data[IFieldInfaqType].(string)
	if infaqType == "" {
		return errors.New("infaq_type wajib diisi")
	}
	if !validInfaqTypes[infaqType] {
		return fmt.Errorf("infaq_type tidak valid: %q", infaqType)
	}

	amount, ok := data[IFieldAmount].(float64)
	if !ok {
		return errors.New("amount wajib diisi")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih dari 0, got %.2f", amount)
	}

	designation, _ := data[IFieldDesignation].(string)
	if designation == "" {
		return errors.New("designation wajib diisi")
	}

	paymentMethod, _ := data[IFieldPaymentMethod].(string)
	if paymentMethod == "" {
		return errors.New("payment_method wajib diisi")
	}
	if !validPaymentMethods[paymentMethod] {
		return fmt.Errorf("payment_method tidak valid: %q", paymentMethod)
	}
	return nil
}

// ============================================================================
// mustahik — Field name constants
// ============================================================================

const (
	MFieldFullName        = "full_name"
	MFieldAsnafCategories = "asnaf_categories"
	MFieldPrimaryCategory = "primary_category"
	MFieldNeedsAssessment = "needs_assessment"
	MFieldAssessmentDate  = "assessment_date"
	MFieldAssessedBy      = "assessed_by"
	MFieldBranchID        = "branch_id"
	MFieldPhone           = "phone"
	MFieldAddress         = "address"
	MFieldNextReviewDate  = "next_review_date"
	MFieldStatus          = "status"
)

// mustahik — Status constants.
const (
	MStatusActive    = "active"
	MStatusInactive  = "inactive"
	MStatusGraduated = "graduated"
)

var validMustahikStatuses = map[string]bool{
	MStatusActive: true, MStatusInactive: true, MStatusGraduated: true,
}

// MustahikDescriptor mengimplementasi vernon.DomainDescriptor untuk mustahik.
type MustahikDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *MustahikDescriptor) TableName() string { return "mustahik" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo branch (autoload, optional).
func (d *MustahikDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		"branch": {
			Domain:     "branches",
			Type:       vernon.RelBelongsTo,
			FK:         MFieldBranchID,
			LocalKey:   MFieldBranchID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
	}
}

// Validate memvalidasi invariant mustahik sebelum write.
func (d *MustahikDescriptor) Validate(data map[string]any) error {
	fullName, _ := data[MFieldFullName].(string)
	if fullName == "" {
		return errors.New("full_name wajib diisi")
	}

	asnafCats := data[MFieldAsnafCategories]
	if asnafCats == nil {
		return errors.New("asnaf_categories wajib diisi")
	}

	primaryCat, _ := data[MFieldPrimaryCategory].(string)
	if primaryCat == "" {
		return errors.New("primary_category wajib diisi")
	}
	if !validAsnafCategories[primaryCat] {
		return fmt.Errorf("primary_category tidak valid: %q", primaryCat)
	}

	needsAssessment, _ := data[MFieldNeedsAssessment].(string)
	if needsAssessment == "" {
		return errors.New("needs_assessment wajib diisi")
	}

	assessmentDate, _ := data[MFieldAssessmentDate].(string)
	if assessmentDate == "" {
		return errors.New("assessment_date wajib diisi")
	}

	assessedBy, _ := data[MFieldAssessedBy].(string)
	if assessedBy == "" {
		return errors.New("assessed_by wajib diisi")
	}

	status, _ := data[MFieldStatus].(string)
	if status != "" && !validMustahikStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}

// ============================================================================
// tazir_fund — Field name constants
// ============================================================================

const (
	TFFieldDendaID       = "denda_id"
	TFFieldPinjamanID    = "pinjaman_id"
	TFFieldNasabahID     = "nasabah_id"
	TFFieldBranchID      = "branch_id"
	TFFieldAmount        = "amount"
	TFFieldCollectionDate = "collection_date"
	TFFieldDistributionID = "distribution_id"
	TFFieldNotes         = "notes"
	TFFieldStatus        = "status"
)

// tazir_fund — Status constants.
const (
	TFStatusCollected   = "collected"
	TFStatusDistributed = "distributed"
)

// tazir_fund — Relation name constants.
const (
	TFRelDenda    = "denda"
	TFRelPinjaman = "pinjaman"
	TFRelNasabah  = "nasabah"
)

var validTFStatuses = map[string]bool{
	TFStatusCollected: true, TFStatusDistributed: true,
}

// TazirFundDescriptor mengimplementasi vernon.DomainDescriptor untuk tazir_fund.
type TazirFundDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *TazirFundDescriptor) TableName() string { return "tazir_fund" }

// DefaultRels mendefinisikan relasi domain ini.
// BelongsTo denda, pinjaman, nasabah (autoload).
func (d *TazirFundDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		TFRelDenda: {
			Domain:     "denda",
			Type:       vernon.RelBelongsTo,
			FK:         TFFieldDendaID,
			LocalKey:   TFFieldDendaID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"penalty_type", "final_amount", "status"},
		},
		TFRelPinjaman: {
			Domain:     "pinjaman",
			Type:       vernon.RelBelongsTo,
			FK:         TFFieldPinjamanID,
			LocalKey:   TFFieldPinjamanID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"no_pinjaman", "status"},
		},
		TFRelNasabah: {
			Domain:     "nasabah",
			Type:       vernon.RelBelongsTo,
			FK:         TFFieldNasabahID,
			LocalKey:   TFFieldNasabahID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"nama_lengkap", "no_nasabah"},
		},
	}
}

// Validate memvalidasi invariant tazir_fund sebelum write.
func (d *TazirFundDescriptor) Validate(data map[string]any) error {
	dendaID, _ := data[TFFieldDendaID].(string)
	if dendaID == "" {
		return errors.New("denda_id wajib diisi")
	}

	pinjamanID, _ := data[TFFieldPinjamanID].(string)
	if pinjamanID == "" {
		return errors.New("pinjaman_id wajib diisi")
	}

	nasabahID, _ := data[TFFieldNasabahID].(string)
	if nasabahID == "" {
		return errors.New("nasabah_id wajib diisi")
	}

	amount, ok := data[TFFieldAmount].(float64)
	if !ok {
		return errors.New("amount wajib diisi")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih dari 0, got %.2f", amount)
	}

	status, _ := data[TFFieldStatus].(string)
	if status != "" && !validTFStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}
