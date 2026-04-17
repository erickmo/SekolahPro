import { createVernonService } from './vernon.service'
import type {
  FeeType,
  StudentInvoice,
  StudentPayment,
  RaporTemplate,
  RaporRecord,
  LeaveType,
  LeaveBalance,
  LeaveRequest,
  DutySchedule,
  TeacherSubstitution,
  CounselingCase,
  CounselingSession,
  TeacherAttendance,
  TeacherWorkload,
  TeacherWorkloadItem,
} from '@/types/sprint6.types'

// ── FeeType ──────────────────────────────────────────────────────────────────

function transformFeeType(raw: Record<string, unknown>): FeeType {
  return {
    id: raw.id as string,
    name: (raw.name ?? '') as string,
    code: (raw.code ?? '') as string,
    feeCategory: (raw.fee_category ?? 'monthly') as FeeType['feeCategory'],
    defaultAmount: (raw.default_amount ?? 0) as number,
    description: (raw.description ?? '') as string,
    isActive: (raw.is_active ?? true) as boolean,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const feeTypeService = createVernonService<FeeType, Record<string, unknown>>(
  '/fee_types',
  transformFeeType,
)

// ── StudentInvoice ───────────────────────────────────────────────────────────

function transformStudentInvoice(raw: Record<string, unknown>): StudentInvoice {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  const ftData = (data.fee_type ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    feeTypeId: (raw.fee_type_id ?? '') as string,
    feeTypeName: (ftData.name ?? '') as string,
    feeTypeCode: (ftData.code ?? '') as string,
    feeCategory: (ftData.fee_category ?? 'monthly') as StudentInvoice['feeCategory'],
    invoiceNo: (raw.invoice_no ?? '') as string,
    periodMonth: (raw.period_month ?? null) as number | null,
    periodYear: (raw.period_year ?? 0) as number,
    amount: (raw.amount ?? 0) as number,
    totalAmount: (raw.total_amount ?? 0) as number,
    paidAmount: (raw.paid_amount ?? 0) as number,
    dueDate: (raw.due_date ?? '') as string,
    status: (raw.status ?? 'unpaid') as StudentInvoice['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const studentInvoiceService = createVernonService<StudentInvoice, Record<string, unknown>>(
  '/student_invoices',
  transformStudentInvoice,
)

// ── StudentPayment ───────────────────────────────────────────────────────────

function transformStudentPayment(raw: Record<string, unknown>): StudentPayment {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const invoiceData = (data.invoice ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    invoiceId: (raw.invoice_id ?? '') as string,
    invoiceNo: (invoiceData.invoice_no ?? '') as string,
    invoiceTotalAmount: (invoiceData.total_amount ?? 0) as number,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    receiptNo: (raw.receipt_no ?? '') as string,
    amount: (raw.amount ?? 0) as number,
    paymentMethod: (raw.payment_method ?? 'cash') as StudentPayment['paymentMethod'],
    paymentDate: (raw.payment_date ?? '') as string,
    receivedBy: (raw.received_by ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const studentPaymentService = createVernonService<StudentPayment, Record<string, unknown>>(
  '/student_payments',
  transformStudentPayment,
)

// ── RaporTemplate ────────────────────────────────────────────────────────────

function transformRaporTemplate(raw: Record<string, unknown>): RaporTemplate {
  return {
    id: raw.id as string,
    name: (raw.name ?? '') as string,
    curriculumType: (raw.curriculum_type ?? 'merdeka') as RaporTemplate['curriculumType'],
    gradeLevels: (raw.grade_levels ?? []) as string[],
    description: (raw.description ?? '') as string,
    layoutConfig: (raw.layout_config ?? {}) as Record<string, unknown>,
    headerConfig: (raw.header_config ?? {}) as Record<string, unknown>,
    isActive: (raw.is_active ?? true) as boolean,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const raporTemplateService = createVernonService<RaporTemplate, Record<string, unknown>>(
  '/rapor_templates',
  transformRaporTemplate,
)

// ── RaporRecord ──────────────────────────────────────────────────────────────

function transformRaporRecord(raw: Record<string, unknown>): RaporRecord {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  const crData = (data.class_room ?? {}) as Record<string, unknown>
  const tplData = (data.template ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    classRoomId: (raw.class_room_id ?? '') as string,
    classRoomName: (crData.name ?? '') as string,
    templateId: (raw.template_id ?? '') as string,
    templateName: (tplData.name ?? '') as string,
    curriculumType: (tplData.curriculum_type ?? 'merdeka') as RaporRecord['curriculumType'],
    semester: (raw.semester ?? 'ganjil') as RaporRecord['semester'],
    status: (raw.status ?? 'draft') as RaporRecord['status'],
    notes: (raw.notes ?? '') as string,
    pdfUrl: (raw.pdf_url ?? '') as string,
    generatedAt: (raw.generated_at ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const raporRecordService = createVernonService<RaporRecord, Record<string, unknown>>(
  '/rapor_records',
  transformRaporRecord,
)

// ── LeaveType (ADR-S030) ─────────────────────────────────────────────────────

function transformLeaveType(raw: Record<string, unknown>): LeaveType {
  return {
    id: raw.id as string,
    code: (raw.code ?? '') as LeaveType['code'],
    name: (raw.name ?? '') as string,
    description: (raw.description ?? '') as string,
    maxDaysPerYear: (raw.max_days_per_year ?? null) as number | null,
    isPaid: (raw.is_paid ?? true) as boolean,
    applicableTo: (raw.applicable_to ?? 'all') as LeaveType['applicableTo'],
    approvalLevels: (raw.approval_levels ?? 1) as number,
    requiresDocument: (raw.requires_document ?? false) as boolean,
    isActive: (raw.is_active ?? true) as boolean,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const leaveTypeService = createVernonService<LeaveType, Record<string, unknown>>(
  '/leave_types',
  transformLeaveType,
)

// ── LeaveBalance (ADR-S030) ──────────────────────────────────────────────────

function transformLeaveBalance(raw: Record<string, unknown>): LeaveBalance {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  const ltData = (data.leave_type ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    teacherEmployeeType: (teacherData.employee_type ?? '') as string,
    leaveTypeId: (raw.leave_type_id ?? '') as string,
    leaveTypeCode: (ltData.code ?? '') as string,
    leaveTypeName: (ltData.name ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    year: (raw.year ?? 0) as number,
    initialBalance: (raw.initial_balance ?? 0) as number,
    used: (raw.used ?? 0) as number,
    remaining: (raw.remaining ?? 0) as number,
    carryOver: (raw.carry_over ?? 0) as number,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const leaveBalanceService = createVernonService<LeaveBalance, Record<string, unknown>>(
  '/leave_balances',
  transformLeaveBalance,
)

// ── LeaveRequest (ADR-S030) ──────────────────────────────────────────────────

function transformLeaveRequest(raw: Record<string, unknown>): LeaveRequest {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  const ltData = (data.leave_type ?? {}) as Record<string, unknown>
  const lbData = (data.leave_balance ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    teacherEmployeeType: (teacherData.employee_type ?? '') as string,
    leaveTypeId: (raw.leave_type_id ?? '') as string,
    leaveTypeName: (ltData.name ?? '') as string,
    leaveTypeCode: (ltData.code ?? '') as string,
    leaveBalanceId: (raw.leave_balance_id ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    startDate: (raw.start_date ?? '') as string,
    endDate: (raw.end_date ?? '') as string,
    totalDays: (raw.total_days ?? 0) as number,
    reason: (raw.reason ?? '') as string,
    status: (raw.status ?? 'draft') as LeaveRequest['status'],
    documentUrl: (raw.document_url ?? '') as string,
    approvalNotes: (raw.approval_notes ?? '') as string,
    currentApprovalLevel: (raw.current_approval_level ?? 0) as number,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const leaveRequestService = createVernonService<LeaveRequest, Record<string, unknown>>(
  '/leave_requests',
  transformLeaveRequest,
)

// ── DutySchedule (ADR-S032) ──────────────────────────────────────────────────

function transformDutySchedule(raw: Record<string, unknown>): DutySchedule {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    teacherEmployeeType: (teacherData.employee_type ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    semester: (raw.semester ?? 'ganjil') as DutySchedule['semester'],
    dayOfWeek: (raw.day_of_week ?? 1) as number,
    dutyType: (raw.duty_type ?? 'piket_pagi') as DutySchedule['dutyType'],
    startTime: (raw.start_time ?? '') as string,
    endTime: (raw.end_time ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const dutyScheduleService = createVernonService<DutySchedule, Record<string, unknown>>(
  '/duty_schedules',
  transformDutySchedule,
)

// ── TeacherSubstitution (ADR-S032) ───────────────────────────────────────────

function transformTeacherSubstitution(raw: Record<string, unknown>): TeacherSubstitution {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const origTeacherData = (data.original_teacher ?? {}) as Record<string, unknown>
  const subTeacherData = (data.substitute_teacher ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  const crData = (data.class_room ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    originalTeacherId: (raw.original_teacher_id ?? '') as string,
    originalTeacherName: (origTeacherData.full_name ?? '') as string,
    substituteTeacherId: (raw.substitute_teacher_id ?? '') as string,
    substituteTeacherName: (subTeacherData.full_name ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    substitutionDate: (raw.substitution_date ?? '') as string,
    reasonType: (raw.reason_type ?? 'lain_lain') as TeacherSubstitution['reasonType'],
    classRoomId: (raw.class_room_id ?? '') as string,
    classRoomName: (crData.name ?? '') as string,
    status: (raw.status ?? 'pending') as TeacherSubstitution['status'],
    scheduleEntryId: (raw.schedule_entry_id ?? '') as string,
    timeSlotId: (raw.time_slot_id ?? '') as string,
    originalSubjectId: (raw.original_subject_id ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teacherSubstitutionService = createVernonService<TeacherSubstitution, Record<string, unknown>>(
  '/teacher_substitutions',
  transformTeacherSubstitution,
)

// ── Counseling Cases (ADR-S017) ────────────────────────────────────────────────

function transformCounselingCase(raw: Record<string, unknown>): CounselingCase {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    caseNo: (raw.case_no ?? '') as string,
    category: (raw.category ?? 'other') as CounselingCase['category'],
    title: (raw.title ?? '') as string,
    description: (raw.description ?? '') as string,
    severity: (raw.severity ?? 'low') as CounselingCase['severity'],
    status: (raw.status ?? 'open') as CounselingCase['status'],
    openedDate: (raw.opened_date ?? '') as string,
    closedDate: (raw.closed_date ?? '') as string,
    resolution: (raw.resolution ?? '') as string,
    counselorId: (raw.counselor_id ?? '') as string,
    referredTo: (raw.referred_to ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const counselingCaseService = createVernonService<CounselingCase, Record<string, unknown>>(
  '/counseling_cases',
  transformCounselingCase,
)

// ── Counseling Sessions (ADR-S017) ─────────────────────────────────────────────

function transformCounselingSession(raw: Record<string, unknown>): CounselingSession {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const caseData = (data.case ?? {}) as Record<string, unknown>
  const studentData = (data.student ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    caseId: (raw.case_id ?? '') as string,
    caseNo: (caseData.case_no ?? '') as string,
    caseTitle: (caseData.title ?? '') as string,
    studentId: (raw.student_id ?? '') as string,
    studentName: (studentData.full_name ?? '') as string,
    studentNis: (studentData.nis ?? '') as string,
    sessionDate: (raw.session_date ?? '') as string,
    sessionType: (raw.session_type ?? 'individual') as CounselingSession['sessionType'],
    durationMinutes: (raw.duration_minutes ?? 0) as number,
    notes: (raw.notes ?? '') as string,
    recommendation: (raw.recommendation ?? '') as string,
    counselorId: (raw.counselor_id ?? '') as string,
    parentPresent: (raw.parent_present ?? false) as boolean,
    followUpDate: (raw.follow_up_date ?? '') as string,
    followUpNote: (raw.follow_up_note ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const counselingSessionService = createVernonService<CounselingSession, Record<string, unknown>>(
  '/counseling_sessions',
  transformCounselingSession,
)

// ── Teacher Attendance (ADR-S026) ──────────────────────────────────────────────

function transformTeacherAttendance(raw: Record<string, unknown>): TeacherAttendance {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    attendanceDate: (raw.attendance_date ?? '') as string,
    status: (raw.status ?? 'present') as TeacherAttendance['status'],
    clockIn: (raw.clock_in ?? '') as string,
    clockOut: (raw.clock_out ?? '') as string,
    lateMinutes: (raw.late_minutes ?? 0) as number,
    earlyLeaveMinutes: (raw.early_leave_minutes ?? 0) as number,
    clockInMethod: (raw.clock_in_method ?? '') as string,
    clockOutMethod: (raw.clock_out_method ?? '') as string,
    note: (raw.note ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teacherAttendanceService = createVernonService<TeacherAttendance, Record<string, unknown>>(
  '/teacher_attendances',
  transformTeacherAttendance,
)

// ── Teacher Workload (ADR-S027) ────────────────────────────────────────────────

function transformTeacherWorkload(raw: Record<string, unknown>): TeacherWorkload {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    semester: (raw.semester ?? 'ganjil') as TeacherWorkload['semester'],
    teachingHours: Number(raw.teaching_hours ?? 0),
    additionalHours: Number(raw.additional_hours ?? 0),
    totalHours: Number(raw.total_hours ?? 0),
    minimumRequired: Number(raw.minimum_required ?? 24),
    isFulfilled: (raw.is_fulfilled ?? false) as boolean,
    fulfillmentStatus: (raw.fulfillment_status ?? 'kurang') as TeacherWorkload['fulfillmentStatus'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teacherWorkloadService = createVernonService<TeacherWorkload, Record<string, unknown>>(
  '/teacher_workloads',
  transformTeacherWorkload,
)

// ── Teacher Workload Items (ADR-S027) ──────────────────────────────────────────

function transformTeacherWorkloadItem(raw: Record<string, unknown>): TeacherWorkloadItem {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    workloadId: (raw.workload_id ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    itemType: (raw.item_type ?? 'mengajar') as TeacherWorkloadItem['itemType'],
    description: (raw.description ?? '') as string,
    hoursPerWeek: Number(raw.hours_per_week ?? 0),
    referenceType: (raw.reference_type ?? '') as string,
    referenceId: (raw.reference_id ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teacherWorkloadItemService = createVernonService<TeacherWorkloadItem, Record<string, unknown>>(
  '/teacher_workload_items',
  transformTeacherWorkloadItem,
)
