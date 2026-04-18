import { createVernonService } from './vernon.service'
import type {
  LibraryBook, LibraryCopy, LibraryBorrow,
  Laboratory, LabEquipment, LabUsageLog,
  Asset, AssetMaintenance,
  Facility, FacilityBooking,
  EvaluationCompetency, TeacherEvaluation, EvaluationScore,
  TeacherCertification, DevelopmentActivity, CreditSummary,
  PayrollConfig, PayrollPeriod, PayrollEntry, PayrollComponent,
} from '@/types/sprint7.types'

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S038: Library Management
// ═══════════════════════════════════════════════════════════════════════════════

function transformLibraryBook(raw: Record<string, unknown>): LibraryBook {
  return {
    id: raw.id as string,
    title: (raw.title ?? '') as string,
    author: (raw.author ?? '') as string,
    isbn: (raw.isbn ?? '') as string,
    publisher: (raw.publisher ?? '') as string,
    publicationYear: (raw.publication_year ?? null) as number | null,
    category: (raw.category ?? 'other') as LibraryBook['category'],
    language: (raw.language ?? 'id') as LibraryBook['language'],
    ddcClassification: (raw.ddc_classification ?? '') as string,
    description: (raw.description ?? '') as string,
    coverUrl: (raw.cover_url ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const libraryBookService = createVernonService<LibraryBook, Record<string, unknown>>(
  '/library_books',
  transformLibraryBook,
)

function transformLibraryCopy(raw: Record<string, unknown>): LibraryCopy {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const bookData = (data.book ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    bookId: (raw.book_id ?? '') as string,
    bookTitle: (bookData.title ?? '') as string,
    bookIsbn: (bookData.isbn ?? '') as string,
    copyNumber: (raw.copy_number ?? '') as string,
    barcode: (raw.barcode ?? '') as string,
    condition: (raw.condition ?? 'good') as LibraryCopy['condition'],
    locationShelf: (raw.location_shelf ?? '') as string,
    acquisitionDate: (raw.acquisition_date ?? '') as string,
    acquisitionSource: (raw.acquisition_source ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const libraryCopyService = createVernonService<LibraryCopy, Record<string, unknown>>(
  '/library_copies',
  transformLibraryCopy,
)

function transformLibraryBorrow(raw: Record<string, unknown>): LibraryBorrow {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const copyData = (data.copy ?? {}) as Record<string, unknown>
  const bookData = (data.book ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    copyId: (raw.copy_id ?? '') as string,
    copyBarcode: (copyData.barcode ?? '') as string,
    copyCondition: (copyData.condition ?? '') as string,
    bookId: (raw.book_id ?? '') as string,
    bookTitle: (bookData.title ?? '') as string,
    bookIsbn: (bookData.isbn ?? '') as string,
    borrowerType: (raw.borrower_type ?? 'student') as LibraryBorrow['borrowerType'],
    borrowerId: (raw.borrower_id ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    borrowDate: (raw.borrow_date ?? '') as string,
    dueDate: (raw.due_date ?? '') as string,
    returnDate: (raw.return_date ?? '') as string,
    actualReturnDate: (raw.actual_return_date ?? '') as string,
    fineAmount: Number(raw.fine_amount ?? 0),
    fineStatus: (raw.fine_status ?? 'none') as LibraryBorrow['fineStatus'],
    note: (raw.note ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const libraryBorrowService = createVernonService<LibraryBorrow, Record<string, unknown>>(
  '/library_borrows',
  transformLibraryBorrow,
)

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S039: Laboratory Management
// ═══════════════════════════════════════════════════════════════════════════════

function transformLaboratory(raw: Record<string, unknown>): Laboratory {
  return {
    id: raw.id as string,
    name: (raw.name ?? '') as string,
    type: (raw.type ?? 'ipa') as Laboratory['type'],
    capacity: (raw.capacity ?? null) as number | null,
    building: (raw.building ?? '') as string,
    floor: (raw.floor ?? null) as number | null,
    roomNumber: (raw.room_number ?? '') as string,
    equipmentCount: (raw.equipment_count ?? 0) as number,
    safetyRating: (raw.safety_rating ?? null) as Laboratory['safetyRating'],
    lastInspectionDate: (raw.last_inspection_date ?? '') as string,
    status: (raw.status ?? 'active') as Laboratory['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const labService = createVernonService<Laboratory, Record<string, unknown>>(
  '/laboratories',
  transformLaboratory,
)

function transformLabEquipment(raw: Record<string, unknown>): LabEquipment {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const labData = (data.lab ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    labId: (raw.lab_id ?? '') as string,
    labName: (labData.name ?? '') as string,
    labType: (labData.type ?? '') as string,
    labRoomNumber: (labData.room_number ?? '') as string,
    name: (raw.name ?? '') as string,
    category: (raw.category ?? 'other') as LabEquipment['category'],
    brand: (raw.brand ?? '') as string,
    model: (raw.model ?? '') as string,
    serialNumber: (raw.serial_number ?? '') as string,
    condition: (raw.condition ?? 'good') as LabEquipment['condition'],
    quantity: (raw.quantity ?? 1) as number,
    unit: (raw.unit ?? 'unit') as string,
    purchaseDate: (raw.purchase_date ?? '') as string,
    purchasePrice: (raw.purchase_price ?? null) as number | null,
    calibrationDate: (raw.calibration_date ?? '') as string,
    nextCalibration: (raw.next_calibration ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const labEquipmentService = createVernonService<LabEquipment, Record<string, unknown>>(
  '/lab_equipment',
  transformLabEquipment,
)

function transformLabUsageLog(raw: Record<string, unknown>): LabUsageLog {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const labData = (data.lab ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  const crData = (data.class_room ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    labId: (raw.lab_id ?? '') as string,
    labName: (labData.name ?? '') as string,
    labType: (labData.type ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    classRoomId: (raw.class_room_id ?? '') as string,
    classRoomName: (crData.name ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    subjectId: (raw.subject_id ?? '') as string,
    usageDate: (raw.usage_date ?? '') as string,
    startTime: (raw.start_time ?? '') as string,
    endTime: (raw.end_time ?? '') as string,
    topic: (raw.topic ?? '') as string,
    participantCount: (raw.participant_count ?? null) as number | null,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const labUsageLogService = createVernonService<LabUsageLog, Record<string, unknown>>(
  '/lab_usage_logs',
  transformLabUsageLog,
)

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S040: Asset & Inventory Management
// ═══════════════════════════════════════════════════════════════════════════════

function transformAsset(raw: Record<string, unknown>): Asset {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const roomData = (data.room ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    assetCode: (raw.asset_code ?? '') as string,
    name: (raw.name ?? '') as string,
    category: (raw.category ?? 'peralatan') as Asset['category'],
    description: (raw.description ?? '') as string,
    acquisitionDate: (raw.acquisition_date ?? '') as string,
    acquisitionCost: (raw.acquisition_cost ?? null) as number | null,
    ownershipType: (raw.ownership_type ?? 'owned') as Asset['ownershipType'],
    condition: (raw.condition ?? 'good') as Asset['condition'],
    locationRoomId: (raw.location_room_id ?? '') as string,
    locationRoomName: (roomData.name ?? '') as string,
    locationBuilding: (roomData.building ?? '') as string,
    responsiblePerson: (raw.responsible_person ?? '') as string,
    vendor: (raw.vendor ?? '') as string,
    warrantyExpiry: (raw.warranty_expiry ?? '') as string,
    depreciationMethod: (raw.depreciation_method ?? 'straight_line') as string,
    usefulLifeYears: (raw.useful_life_years ?? null) as number | null,
    bookValue: (raw.book_value ?? null) as number | null,
    lifecycleStatus: (raw.lifecycle_status ?? 'in_use') as Asset['lifecycleStatus'],
    disposalDate: (raw.disposal_date ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const assetService = createVernonService<Asset, Record<string, unknown>>(
  '/assets',
  transformAsset,
)

function transformAssetMaintenance(raw: Record<string, unknown>): AssetMaintenance {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const assetData = (data.asset ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    assetId: (raw.asset_id ?? '') as string,
    assetCode: (assetData.asset_code ?? '') as string,
    assetName: (assetData.name ?? '') as string,
    assetCondition: (assetData.condition ?? '') as string,
    type: (raw.type ?? 'preventive') as AssetMaintenance['type'],
    description: (raw.description ?? '') as string,
    startDate: (raw.start_date ?? '') as string,
    completionDate: (raw.completion_date ?? '') as string,
    cost: (raw.cost ?? null) as number | null,
    vendor: (raw.vendor ?? '') as string,
    technician: (raw.technician ?? '') as string,
    status: (raw.status ?? 'scheduled') as AssetMaintenance['status'],
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const assetMaintenanceService = createVernonService<AssetMaintenance, Record<string, unknown>>(
  '/asset_maintenances',
  transformAssetMaintenance,
)

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S041: Facility Booking
// ═══════════════════════════════════════════════════════════════════════════════

function transformFacility(raw: Record<string, unknown>): Facility {
  return {
    id: raw.id as string,
    name: (raw.name ?? '') as string,
    type: (raw.type ?? 'classroom') as Facility['type'],
    building: (raw.building ?? '') as string,
    floor: (raw.floor ?? null) as number | null,
    roomNumber: (raw.room_number ?? '') as string,
    capacity: (raw.capacity ?? null) as number | null,
    isBookable: (raw.is_bookable ?? true) as boolean,
    requiresApproval: (raw.requires_approval ?? false) as boolean,
    bookingAdvanceMinDays: (raw.booking_advance_min_days ?? 1) as number,
    bookingAdvanceMaxDays: (raw.booking_advance_max_days ?? 30) as number,
    status: (raw.status ?? 'active') as Facility['status'],
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const facilityService = createVernonService<Facility, Record<string, unknown>>(
  '/facilities',
  transformFacility,
)

function transformFacilityBooking(raw: Record<string, unknown>): FacilityBooking {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const facilityData = (data.facility ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    facilityId: (raw.facility_id ?? '') as string,
    facilityName: (facilityData.name ?? '') as string,
    facilityType: (facilityData.type ?? '') as string,
    facilityBuilding: (facilityData.building ?? '') as string,
    facilityRoomNumber: (facilityData.room_number ?? '') as string,
    requesterType: (raw.requester_type ?? 'teacher') as FacilityBooking['requesterType'],
    requesterId: (raw.requester_id ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    bookingDate: (raw.booking_date ?? '') as string,
    startTime: (raw.start_time ?? '') as string,
    endTime: (raw.end_time ?? '') as string,
    purpose: (raw.purpose ?? '') as string,
    participantCount: (raw.participant_count ?? null) as number | null,
    status: (raw.status ?? 'pending') as FacilityBooking['status'],
    approvedBy: (raw.approved_by ?? '') as string,
    approvedAt: (raw.approved_at ?? '') as string,
    rejectionReason: (raw.rejection_reason ?? '') as string,
    recurrencePattern: (raw.recurrence_pattern ?? null) as FacilityBooking['recurrencePattern'],
    recurrenceEndDate: (raw.recurrence_end_date ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const facilityBookingService = createVernonService<FacilityBooking, Record<string, unknown>>(
  '/facility_bookings',
  transformFacilityBooking,
)

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S028: Teacher Evaluation (PKG)
// ═══════════════════════════════════════════════════════════════════════════════

function transformEvaluationCompetency(raw: Record<string, unknown>): EvaluationCompetency {
  return {
    id: raw.id as string,
    name: (raw.name ?? '') as string,
    area: (raw.area ?? 'pedagogic') as EvaluationCompetency['area'],
    indicatorCount: (raw.indicator_count ?? 0) as number,
    weight: Number(raw.weight ?? 1),
    description: (raw.description ?? '') as string,
    isActive: (raw.is_active ?? true) as boolean,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const evaluationCompetencyService = createVernonService<EvaluationCompetency, Record<string, unknown>>(
  '/teacher_evaluation_competencies',
  transformEvaluationCompetency,
)

function transformTeacherEvaluation(raw: Record<string, unknown>): TeacherEvaluation {
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
    semester: (raw.semester ?? 'ganjil') as TeacherEvaluation['semester'],
    evaluatorId: (raw.evaluator_id ?? '') as string,
    totalScore: (raw.total_score ?? null) as number | null,
    grade: (raw.grade ?? null) as TeacherEvaluation['grade'],
    workflowStatus: (raw.workflow_status ?? 'draft') as TeacherEvaluation['workflowStatus'],
    evaluationDate: (raw.evaluation_date ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teacherEvaluationService = createVernonService<TeacherEvaluation, Record<string, unknown>>(
  '/teacher_evaluations',
  transformTeacherEvaluation,
)

function transformEvaluationScore(raw: Record<string, unknown>): EvaluationScore {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const evalData = (data.evaluation ?? {}) as Record<string, unknown>
  const compData = (data.competency ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    evaluationId: (raw.evaluation_id ?? '') as string,
    evaluationTeacherId: (evalData.teacher_id ?? '') as string,
    evaluationSemester: (evalData.semester ?? '') as string,
    evaluationWorkflowStatus: (evalData.workflow_status ?? '') as string,
    competencyId: (raw.competency_id ?? '') as string,
    competencyName: (compData.name ?? '') as string,
    competencyArea: (compData.area ?? '') as string,
    competencyWeight: Number(compData.weight ?? 1),
    assessorType: (raw.assessor_type ?? 'self') as EvaluationScore['assessorType'],
    assessorId: (raw.assessor_id ?? '') as string,
    score: (raw.score ?? 0) as number,
    evidence: (raw.evidence ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const evaluationScoreService = createVernonService<EvaluationScore, Record<string, unknown>>(
  '/teacher_evaluation_scores',
  transformEvaluationScore,
)

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S029: Professional Development (PKB)
// ═══════════════════════════════════════════════════════════════════════════════

function transformTeacherCertification(raw: Record<string, unknown>): TeacherCertification {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    certificationType: (raw.certification_type ?? 'profesi') as TeacherCertification['certificationType'],
    certificationNumber: (raw.certification_number ?? '') as string,
    issueDate: (raw.issue_date ?? '') as string,
    expiryDate: (raw.expiry_date ?? '') as string,
    issuingBody: (raw.issuing_body ?? '') as string,
    status: (raw.status ?? 'active') as TeacherCertification['status'],
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const teacherCertificationService = createVernonService<TeacherCertification, Record<string, unknown>>(
  '/teacher_certifications',
  transformTeacherCertification,
)

function transformDevelopmentActivity(raw: Record<string, unknown>): DevelopmentActivity {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    activityType: (raw.activity_type ?? '') as string,
    title: (raw.title ?? '') as string,
    organizer: (raw.organizer ?? '') as string,
    startDate: (raw.start_date ?? '') as string,
    endDate: (raw.end_date ?? '') as string,
    location: (raw.location ?? '') as string,
    creditPoints: Number(raw.credit_points ?? 0),
    certificateNumber: (raw.certificate_number ?? '') as string,
    status: (raw.status ?? 'registered') as DevelopmentActivity['status'],
    description: (raw.description ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const developmentActivityService = createVernonService<DevelopmentActivity, Record<string, unknown>>(
  '/teacher_development_activities',
  transformDevelopmentActivity,
)

function transformCreditSummary(raw: Record<string, unknown>): CreditSummary {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    teacherEmployeeType: (teacherData.employee_type ?? '') as string,
    currentRank: (raw.current_rank ?? 'I/a') as CreditSummary['currentRank'],
    targetRank: (raw.target_rank ?? '') as string,
    totalCredits: Number(raw.total_credits ?? 0),
    requiredCredits: Number(raw.required_credits ?? 0),
    creditGap: (raw.credit_gap ?? null) as number | null,
    lastPromotionDate: (raw.last_promotion_date ?? '') as string,
    nextEligibleDate: (raw.next_eligible_date ?? '') as string,
    skpScore: (raw.skp_score ?? null) as number | null,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const creditSummaryService = createVernonService<CreditSummary, Record<string, unknown>>(
  '/teacher_credit_summaries',
  transformCreditSummary,
)

// ═══════════════════════════════════════════════════════════════════════════════
// ADR-S031: Staff Payroll
// ═══════════════════════════════════════════════════════════════════════════════

function transformPayrollConfig(raw: Record<string, unknown>): PayrollConfig {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    teacherEmployeeType: (teacherData.employee_type ?? '') as string,
    employeeType: (raw.employee_type ?? 'honorer') as PayrollConfig['employeeType'],
    baseSalary: Number(raw.base_salary ?? 0),
    transportAllowance: Number(raw.transport_allowance ?? 0),
    mealAllowance: Number(raw.meal_allowance ?? 0),
    positionAllowance: Number(raw.position_allowance ?? 0),
    familyAllowance: Number(raw.family_allowance ?? 0),
    riceAllowance: Number(raw.rice_allowance ?? 0),
    pphStatus: (raw.pph_status ?? 'non_pkp') as PayrollConfig['pphStatus'],
    bankName: (raw.bank_name ?? '') as string,
    bankAccount: (raw.bank_account ?? '') as string,
    isActive: (raw.is_active ?? true) as boolean,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const payrollConfigService = createVernonService<PayrollConfig, Record<string, unknown>>(
  '/payroll_configs',
  transformPayrollConfig,
)

function transformPayrollPeriod(raw: Record<string, unknown>): PayrollPeriod {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const ayData = (data.academic_year ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    periodName: (raw.period_name ?? '') as string,
    academicYearId: (raw.academic_year_id ?? '') as string,
    academicYearName: (ayData.name ?? '') as string,
    month: (raw.month ?? 1) as number,
    year: (raw.year ?? 2026) as number,
    startDate: (raw.start_date ?? '') as string,
    endDate: (raw.end_date ?? '') as string,
    workflowStatus: (raw.workflow_status ?? 'draft') as PayrollPeriod['workflowStatus'],
    totalEmployees: (raw.total_employees ?? 0) as number,
    totalGross: Number(raw.total_gross ?? 0),
    totalDeductions: Number(raw.total_deductions ?? 0),
    totalNet: Number(raw.total_net ?? 0),
    processedBy: (raw.processed_by ?? '') as string,
    processedAt: (raw.processed_at ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const payrollPeriodService = createVernonService<PayrollPeriod, Record<string, unknown>>(
  '/payroll_periods',
  transformPayrollPeriod,
)

function transformPayrollEntry(raw: Record<string, unknown>): PayrollEntry {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const periodData = (data.period ?? {}) as Record<string, unknown>
  const teacherData = (data.teacher ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    periodId: (raw.period_id ?? '') as string,
    periodName: (periodData.period_name ?? '') as string,
    periodMonth: (periodData.month ?? 0) as number,
    periodYear: (periodData.year ?? 0) as number,
    periodWorkflowStatus: (periodData.workflow_status ?? '') as string,
    teacherId: (raw.teacher_id ?? '') as string,
    teacherName: (teacherData.full_name ?? '') as string,
    teacherNip: (teacherData.nip ?? '') as string,
    employeeType: (raw.employee_type ?? 'honorer') as PayrollEntry['employeeType'],
    baseSalary: Number(raw.base_salary ?? 0),
    totalAllowances: Number(raw.total_allowances ?? 0),
    grossSalary: Number(raw.gross_salary ?? 0),
    totalDeductions: Number(raw.total_deductions ?? 0),
    netSalary: Number(raw.net_salary ?? 0),
    pph21Amount: Number(raw.pph21_amount ?? 0),
    workingDays: (raw.working_days ?? 0) as number,
    presentDays: (raw.present_days ?? 0) as number,
    absentDays: (raw.absent_days ?? 0) as number,
    leaveDays: (raw.leave_days ?? 0) as number,
    payslipNumber: (raw.payslip_number ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const payrollEntryService = createVernonService<PayrollEntry, Record<string, unknown>>(
  '/payroll_entries',
  transformPayrollEntry,
)

function transformPayrollComponent(raw: Record<string, unknown>): PayrollComponent {
  const data = (raw._data ?? {}) as Record<string, unknown>
  const entryData = (data.entry ?? {}) as Record<string, unknown>
  return {
    id: raw.id as string,
    entryId: (raw.entry_id ?? '') as string,
    entryPayslipNumber: (entryData.payslip_number ?? '') as string,
    entryTeacherId: (entryData.teacher_id ?? '') as string,
    componentType: (raw.component_type ?? 'earning') as PayrollComponent['componentType'],
    category: (raw.category ?? '') as string,
    name: (raw.name ?? '') as string,
    amount: Number(raw.amount ?? 0),
    calculationMethod: (raw.calculation_method ?? 'fixed') as PayrollComponent['calculationMethod'],
    isRecurring: (raw.is_recurring ?? true) as boolean,
    referenceId: (raw.reference_id ?? '') as string,
    notes: (raw.notes ?? '') as string,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const payrollComponentService = createVernonService<PayrollComponent, Record<string, unknown>>(
  '/payroll_components',
  transformPayrollComponent,
)
