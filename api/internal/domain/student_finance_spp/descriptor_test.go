package student_finance_spp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/boilerplate/internal/domain/student_finance_spp"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// ============================================================================
// FeeTypeDescriptor tests
// ============================================================================

func TestFeeTypeDescriptor_TableName(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	assert.Equal(t, "fee_types", d.TableName())
}

func TestFeeTypeDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &student_finance_spp.FeeTypeDescriptor{}
}

func TestFeeTypeDescriptor_DefaultRels(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 0, "fee_types should have no relations")
}

func validFeeTypeData() map[string]any {
	return map[string]any{
		student_finance_spp.FTFieldName:         "SPP Bulanan",
		student_finance_spp.FTFieldCode:         "SPP",
		student_finance_spp.FTFieldFeeCategory:  student_finance_spp.FeeCategoryMonthly,
		student_finance_spp.FTFieldDefaultAmount: float64(500000),
	}
}

func TestFeeTypeDescriptor_Validate_Success(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	err := d.Validate(validFeeTypeData())
	assert.NoError(t, err)
}

func TestFeeTypeDescriptor_Validate_AllFeeCategories(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	categories := []string{
		student_finance_spp.FeeCategoryMonthly,
		student_finance_spp.FeeCategoryAnnual,
		student_finance_spp.FeeCategoryOneTime,
		student_finance_spp.FeeCategoryIncidental,
	}
	for _, cat := range categories {
		data := validFeeTypeData()
		data[student_finance_spp.FTFieldFeeCategory] = cat
		err := d.Validate(data)
		assert.NoError(t, err, "fee_category=%q should be valid", cat)
	}
}

func TestFeeTypeDescriptor_Validate_MissingName(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	data := validFeeTypeData()
	delete(data, student_finance_spp.FTFieldName)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "name wajib diisi")
}

func TestFeeTypeDescriptor_Validate_MissingCode(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	data := validFeeTypeData()
	delete(data, student_finance_spp.FTFieldCode)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "code wajib diisi")
}

func TestFeeTypeDescriptor_Validate_MissingFeeCategory(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	data := validFeeTypeData()
	delete(data, student_finance_spp.FTFieldFeeCategory)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "fee_category wajib diisi")
}

func TestFeeTypeDescriptor_Validate_InvalidFeeCategory(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	data := validFeeTypeData()
	data[student_finance_spp.FTFieldFeeCategory] = "quarterly"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "fee_category tidak valid")
}

func TestFeeTypeDescriptor_Validate_MissingDefaultAmount(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	data := validFeeTypeData()
	delete(data, student_finance_spp.FTFieldDefaultAmount)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "default_amount wajib diisi")
}

func TestFeeTypeDescriptor_Validate_ZeroDefaultAmount(t *testing.T) {
	d := &student_finance_spp.FeeTypeDescriptor{}
	data := validFeeTypeData()
	data[student_finance_spp.FTFieldDefaultAmount] = float64(0)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "default_amount harus lebih dari 0")
}

// ============================================================================
// InvoiceDescriptor tests
// ============================================================================

func TestInvoiceDescriptor_TableName(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	assert.Equal(t, "student_invoices", d.TableName())
}

func TestInvoiceDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &student_finance_spp.InvoiceDescriptor{}
}

func TestInvoiceDescriptor_DefaultRels_Count(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 3, "student_invoices should have 3 BelongsTo relations")
}

func TestInvoiceDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_finance_spp.InvRelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_finance_spp.InvFieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func TestInvoiceDescriptor_DefaultRels_AcademicYearRelation(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_finance_spp.InvRelAcademicYear]
	require.True(t, ok, "should have academic_year relation")

	assert.Equal(t, "academic_years", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_finance_spp.InvFieldAcademicYearID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"name", "code"}, rel.Fields)
}

func TestInvoiceDescriptor_DefaultRels_FeeTypeRelation(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_finance_spp.InvRelFeeType]
	require.True(t, ok, "should have fee_type relation")

	assert.Equal(t, "fee_types", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_finance_spp.InvFieldFeeTypeID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"name", "code", "fee_category"}, rel.Fields)
}

func validInvoiceData() map[string]any {
	return map[string]any{
		student_finance_spp.InvFieldStudentID:      "00000000-0000-0000-0000-000000000001",
		student_finance_spp.InvFieldAcademicYearID: "00000000-0000-0000-0000-000000000002",
		student_finance_spp.InvFieldFeeTypeID:      "00000000-0000-0000-0000-000000000003",
		student_finance_spp.InvFieldInvoiceNo:      "INV-2026-001",
		student_finance_spp.InvFieldPeriodYear:     float64(2026),
		student_finance_spp.InvFieldAmount:         float64(500000),
		student_finance_spp.InvFieldTotalAmount:    float64(500000),
		student_finance_spp.InvFieldDueDate:        "2026-04-30",
		student_finance_spp.InvFieldStatus:         student_finance_spp.InvoiceStatusUnpaid,
	}
}

func TestInvoiceDescriptor_Validate_Success(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	err := d.Validate(validInvoiceData())
	assert.NoError(t, err)
}

func TestInvoiceDescriptor_Validate_AllStatuses(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	statuses := []string{
		student_finance_spp.InvoiceStatusUnpaid,
		student_finance_spp.InvoiceStatusPartial,
		student_finance_spp.InvoiceStatusPaid,
		student_finance_spp.InvoiceStatusOverdue,
		student_finance_spp.InvoiceStatusWaived,
	}
	for _, status := range statuses {
		data := validInvoiceData()
		data[student_finance_spp.InvFieldStatus] = status
		err := d.Validate(data)
		assert.NoError(t, err, "status=%q should be valid", status)
	}
}

func TestInvoiceDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	data := validInvoiceData()
	delete(data, student_finance_spp.InvFieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestInvoiceDescriptor_Validate_MissingAcademicYearID(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	data := validInvoiceData()
	delete(data, student_finance_spp.InvFieldAcademicYearID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "academic_year_id wajib diisi")
}

func TestInvoiceDescriptor_Validate_MissingFeeTypeID(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	data := validInvoiceData()
	delete(data, student_finance_spp.InvFieldFeeTypeID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "fee_type_id wajib diisi")
}

func TestInvoiceDescriptor_Validate_MissingStatus(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	data := validInvoiceData()
	delete(data, student_finance_spp.InvFieldStatus)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status wajib diisi")
}

func TestInvoiceDescriptor_Validate_InvalidStatus(t *testing.T) {
	d := &student_finance_spp.InvoiceDescriptor{}
	data := validInvoiceData()
	data[student_finance_spp.InvFieldStatus] = "cancelled"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "status tidak valid")
}

// ============================================================================
// PaymentDescriptor tests
// ============================================================================

func TestPaymentDescriptor_TableName(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	assert.Equal(t, "student_payments", d.TableName())
}

func TestPaymentDescriptor_ImplementsInterface(t *testing.T) {
	var _ vernon.DomainDescriptor = &student_finance_spp.PaymentDescriptor{}
}

func TestPaymentDescriptor_DefaultRels_Count(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	rels := d.DefaultRels()
	assert.Len(t, rels, 2, "student_payments should have 2 BelongsTo relations")
}

func TestPaymentDescriptor_DefaultRels_InvoiceRelation(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_finance_spp.PayRelInvoice]
	require.True(t, ok, "should have invoice relation")

	assert.Equal(t, "student_invoices", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_finance_spp.PayFieldInvoiceID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"invoice_no", "total_amount"}, rel.Fields)
}

func TestPaymentDescriptor_DefaultRels_StudentRelation(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	rels := d.DefaultRels()

	rel, ok := rels[student_finance_spp.PayRelStudent]
	require.True(t, ok, "should have student relation")

	assert.Equal(t, "students", rel.Domain)
	assert.Equal(t, vernon.RelBelongsTo, rel.Type)
	assert.Equal(t, student_finance_spp.PayFieldStudentID, rel.FK)
	assert.True(t, rel.IsAutoload)
	assert.Equal(t, []string{"full_name", "nis"}, rel.Fields)
}

func validPaymentData() map[string]any {
	return map[string]any{
		student_finance_spp.PayFieldInvoiceID:     "00000000-0000-0000-0000-000000000001",
		student_finance_spp.PayFieldStudentID:     "00000000-0000-0000-0000-000000000002",
		student_finance_spp.PayFieldReceiptNo:     "RCT-2026-001",
		student_finance_spp.PayFieldAmount:        float64(500000),
		student_finance_spp.PayFieldPaymentMethod: student_finance_spp.PaymentMethodCash,
		student_finance_spp.PayFieldPaymentDate:   "2026-04-17",
		student_finance_spp.PayFieldReceivedBy:    "00000000-0000-0000-0000-000000000003",
	}
}

func TestPaymentDescriptor_Validate_Success(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	err := d.Validate(validPaymentData())
	assert.NoError(t, err)
}

func TestPaymentDescriptor_Validate_AllPaymentMethods(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	methods := []string{
		student_finance_spp.PaymentMethodCash,
		student_finance_spp.PaymentMethodTransfer,
		student_finance_spp.PaymentMethodDebit,
		student_finance_spp.PaymentMethodQris,
		student_finance_spp.PaymentMethodVA,
	}
	for _, method := range methods {
		data := validPaymentData()
		data[student_finance_spp.PayFieldPaymentMethod] = method
		err := d.Validate(data)
		assert.NoError(t, err, "payment_method=%q should be valid", method)
	}
}

func TestPaymentDescriptor_Validate_MissingInvoiceID(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	data := validPaymentData()
	delete(data, student_finance_spp.PayFieldInvoiceID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "invoice_id wajib diisi")
}

func TestPaymentDescriptor_Validate_MissingStudentID(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	data := validPaymentData()
	delete(data, student_finance_spp.PayFieldStudentID)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "student_id wajib diisi")
}

func TestPaymentDescriptor_Validate_MissingPaymentMethod(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	data := validPaymentData()
	delete(data, student_finance_spp.PayFieldPaymentMethod)
	err := d.Validate(data)
	assert.ErrorContains(t, err, "payment_method wajib diisi")
}

func TestPaymentDescriptor_Validate_InvalidPaymentMethod(t *testing.T) {
	d := &student_finance_spp.PaymentDescriptor{}
	data := validPaymentData()
	data[student_finance_spp.PayFieldPaymentMethod] = "bitcoin"
	err := d.Validate(data)
	assert.ErrorContains(t, err, "payment_method tidak valid")
}
