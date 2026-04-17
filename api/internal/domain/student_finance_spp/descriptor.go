// Package student_finance_spp adalah domain Vernon untuk keuangan siswa (SPP).
//
// Terdiri dari 3 sub-descriptor:
//   - FeeTypeDescriptor: master jenis tagihan (tanpa BelongsTo).
//   - InvoiceDescriptor: tagihan per siswa (3 BelongsTo: student, academic_year, fee_type).
//   - PaymentDescriptor: pembayaran terhadap invoice (2 BelongsTo: invoice, student).
//
// Aturan layer Vernon:
//   - Descriptor hanya mendefinisikan metadata — tidak ada framework dependency.
//   - Validate() hanya untuk invariant dari data itu sendiri.
package student_finance_spp

import (
	"fmt"

	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// FeeType — Master jenis tagihan
// ============================================================================

// Fee type field constants.
const (
	FTFieldName         = "name"
	FTFieldCode         = "code"
	FTFieldFeeCategory  = "fee_category"
	FTFieldDefaultAmount = "default_amount"
	FTFieldDescription  = "description"
	FTFieldIsActive     = "is_active"
)

// Fee category constants.
const (
	FeeCategoryMonthly    = "monthly"
	FeeCategoryAnnual     = "annual"
	FeeCategoryOneTime    = "one_time"
	FeeCategoryIncidental = "incidental"
)

var validFeeCategories = map[string]bool{
	FeeCategoryMonthly: true, FeeCategoryAnnual: true,
	FeeCategoryOneTime: true, FeeCategoryIncidental: true,
}

// FeeTypeDescriptor mengimplementasi vernon.DomainDescriptor untuk fee_types.
type FeeTypeDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *FeeTypeDescriptor) TableName() string { return "fee_types" }

// DefaultRels — fee_types tidak memiliki relasi (standalone config).
func (d *FeeTypeDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{}
}

// Validate memvalidasi invariant fee_types.
func (d *FeeTypeDescriptor) Validate(data map[string]any) error {
	name, _ := data[FTFieldName].(string)
	if name == "" {
		return fmt.Errorf("name wajib diisi")
	}

	code, _ := data[FTFieldCode].(string)
	if code == "" {
		return fmt.Errorf("code wajib diisi")
	}

	feeCategory, _ := data[FTFieldFeeCategory].(string)
	if feeCategory == "" {
		return fmt.Errorf("fee_category wajib diisi")
	}
	if !validFeeCategories[feeCategory] {
		return fmt.Errorf("fee_category tidak valid: %q", feeCategory)
	}

	defaultAmount, ok := data[FTFieldDefaultAmount].(float64)
	if !ok {
		return fmt.Errorf("default_amount wajib diisi")
	}
	if defaultAmount <= 0 {
		return fmt.Errorf("default_amount harus lebih dari 0")
	}

	return nil
}

// ============================================================================
// Invoice — Tagihan per siswa
// ============================================================================

// Invoice field constants.
const (
	InvFieldStudentID      = "student_id"
	InvFieldAcademicYearID = "academic_year_id"
	InvFieldFeeTypeID      = "fee_type_id"
	InvFieldInvoiceNo      = "invoice_no"
	InvFieldPeriodYear     = "period_year"
	InvFieldPeriodMonth    = "period_month"
	InvFieldAmount         = "amount"
	InvFieldTotalAmount    = "total_amount"
	InvFieldPaidAmount     = "paid_amount"
	InvFieldDueDate        = "due_date"
	InvFieldStatus         = "status"
)

// Invoice status constants.
const (
	InvoiceStatusUnpaid  = "unpaid"
	InvoiceStatusPartial = "partial"
	InvoiceStatusPaid    = "paid"
	InvoiceStatusOverdue = "overdue"
	InvoiceStatusWaived  = "waived"
)

// Invoice relation name constants.
const (
	InvRelStudent      = "student"
	InvRelAcademicYear = "academic_year"
	InvRelFeeType      = "fee_type"
)

var validInvoiceStatuses = map[string]bool{
	InvoiceStatusUnpaid: true, InvoiceStatusPartial: true,
	InvoiceStatusPaid: true, InvoiceStatusOverdue: true,
	InvoiceStatusWaived: true,
}

// InvoiceDescriptor mengimplementasi vernon.DomainDescriptor untuk student_invoices.
type InvoiceDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *InvoiceDescriptor) TableName() string { return "student_invoices" }

// DefaultRels mendefinisikan 3 BelongsTo autoload: student, academic_year, fee_type.
func (d *InvoiceDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		InvRelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         InvFieldStudentID,
			LocalKey:   InvFieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
		InvRelAcademicYear: {
			Domain:     "academic_years",
			Type:       vernon.RelBelongsTo,
			FK:         InvFieldAcademicYearID,
			LocalKey:   InvFieldAcademicYearID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code"},
		},
		InvRelFeeType: {
			Domain:     "fee_types",
			Type:       vernon.RelBelongsTo,
			FK:         InvFieldFeeTypeID,
			LocalKey:   InvFieldFeeTypeID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"name", "code", "fee_category"},
		},
	}
}

// Validate memvalidasi invariant student_invoices.
func (d *InvoiceDescriptor) Validate(data map[string]any) error {
	if err := validateInvoiceRequired(data); err != nil {
		return err
	}
	return validateInvoiceEnums(data)
}

func validateInvoiceRequired(data map[string]any) error {
	studentID, _ := data[InvFieldStudentID].(string)
	if studentID == "" {
		return fmt.Errorf("student_id wajib diisi")
	}

	academicYearID, _ := data[InvFieldAcademicYearID].(string)
	if academicYearID == "" {
		return fmt.Errorf("academic_year_id wajib diisi")
	}

	feeTypeID, _ := data[InvFieldFeeTypeID].(string)
	if feeTypeID == "" {
		return fmt.Errorf("fee_type_id wajib diisi")
	}

	invoiceNo, _ := data[InvFieldInvoiceNo].(string)
	if invoiceNo == "" {
		return fmt.Errorf("invoice_no wajib diisi")
	}

	periodYear, _ := data[InvFieldPeriodYear].(float64)
	if periodYear == 0 {
		return fmt.Errorf("period_year wajib diisi")
	}

	amount, ok := data[InvFieldAmount].(float64)
	if !ok {
		return fmt.Errorf("amount wajib diisi")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih dari 0")
	}

	totalAmount, ok := data[InvFieldTotalAmount].(float64)
	if !ok {
		return fmt.Errorf("total_amount wajib diisi")
	}
	if totalAmount < 0 {
		return fmt.Errorf("total_amount tidak boleh negatif")
	}

	dueDate, _ := data[InvFieldDueDate].(string)
	if dueDate == "" {
		return fmt.Errorf("due_date wajib diisi")
	}

	return nil
}

func validateInvoiceEnums(data map[string]any) error {
	status, _ := data[InvFieldStatus].(string)
	if status == "" {
		return fmt.Errorf("status wajib diisi")
	}
	if !validInvoiceStatuses[status] {
		return fmt.Errorf("status tidak valid: %q", status)
	}
	return nil
}

// ============================================================================
// Payment — Pembayaran terhadap invoice
// ============================================================================

// Payment field constants.
const (
	PayFieldInvoiceID     = "invoice_id"
	PayFieldStudentID     = "student_id"
	PayFieldReceiptNo     = "receipt_no"
	PayFieldAmount        = "amount"
	PayFieldPaymentMethod = "payment_method"
	PayFieldPaymentDate   = "payment_date"
	PayFieldReceivedBy    = "received_by"
	PayFieldNotes         = "notes"
)

// Payment method constants.
const (
	PaymentMethodCash     = "cash"
	PaymentMethodTransfer = "transfer"
	PaymentMethodDebit    = "debit"
	PaymentMethodQris     = "qris"
	PaymentMethodVA       = "va"
)

// Payment relation name constants.
const (
	PayRelInvoice = "invoice"
	PayRelStudent = "student"
)

var validPaymentMethods = map[string]bool{
	PaymentMethodCash: true, PaymentMethodTransfer: true,
	PaymentMethodDebit: true, PaymentMethodQris: true,
	PaymentMethodVA: true,
}

// PaymentDescriptor mengimplementasi vernon.DomainDescriptor untuk student_payments.
type PaymentDescriptor struct{}

// TableName mengembalikan nama tabel PostgreSQL.
func (d *PaymentDescriptor) TableName() string { return "student_payments" }

// DefaultRels mendefinisikan 2 BelongsTo autoload: invoice, student.
func (d *PaymentDescriptor) DefaultRels() map[string]vernon.RelDef {
	return map[string]vernon.RelDef{
		PayRelInvoice: {
			Domain:     "student_invoices",
			Type:       vernon.RelBelongsTo,
			FK:         PayFieldInvoiceID,
			LocalKey:   PayFieldInvoiceID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"invoice_no", "total_amount"},
		},
		PayRelStudent: {
			Domain:     "students",
			Type:       vernon.RelBelongsTo,
			FK:         PayFieldStudentID,
			LocalKey:   PayFieldStudentID,
			ForeignKey: "id",
			IsAutoload: true,
			Fields:     []string{"full_name", "nis"},
		},
	}
}

// Validate memvalidasi invariant student_payments.
func (d *PaymentDescriptor) Validate(data map[string]any) error {
	if err := validatePaymentRequired(data); err != nil {
		return err
	}
	return validatePaymentEnums(data)
}

func validatePaymentRequired(data map[string]any) error {
	invoiceID, _ := data[PayFieldInvoiceID].(string)
	if invoiceID == "" {
		return fmt.Errorf("invoice_id wajib diisi")
	}

	studentID, _ := data[PayFieldStudentID].(string)
	if studentID == "" {
		return fmt.Errorf("student_id wajib diisi")
	}

	receiptNo, _ := data[PayFieldReceiptNo].(string)
	if receiptNo == "" {
		return fmt.Errorf("receipt_no wajib diisi")
	}

	amount, ok := data[PayFieldAmount].(float64)
	if !ok {
		return fmt.Errorf("amount wajib diisi")
	}
	if amount <= 0 {
		return fmt.Errorf("amount harus lebih dari 0")
	}

	paymentDate, _ := data[PayFieldPaymentDate].(string)
	if paymentDate == "" {
		return fmt.Errorf("payment_date wajib diisi")
	}

	receivedBy, _ := data[PayFieldReceivedBy].(string)
	if receivedBy == "" {
		return fmt.Errorf("received_by wajib diisi")
	}

	return nil
}

func validatePaymentEnums(data map[string]any) error {
	paymentMethod, _ := data[PayFieldPaymentMethod].(string)
	if paymentMethod == "" {
		return fmt.Errorf("payment_method wajib diisi")
	}
	if !validPaymentMethods[paymentMethod] {
		return fmt.Errorf("payment_method tidak valid: %q", paymentMethod)
	}
	return nil
}
