# 14 — Database Schema (PostgreSQL 17 + Prisma)

> **Terakhir diperbarui**: 20 Maret 2026
> **Database**: PostgreSQL 17
> **ORM**: Prisma 5.x

Schema ini adalah referensi lengkap semua model. Untuk detail bisnis logic tiap model, baca file domain yang sesuai. File ini fokus pada struktur, relasi, dan index.

---

## Setup PostgreSQL 17

```prisma
// packages/database/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}
```

### Fitur PostgreSQL 17 yang Dimanfaatkan

| Fitur | Penggunaan di EDS |
|-------|------------------|
| Row Level Security (RLS) | Isolasi data per `schoolId` (multi-tenant) |
| `MERGE` statement | Upsert di sync Dapodik |
| Logical replication slot failover | HA tanpa kehilangan slot saat failover |
| `pg_walinspect` | Debugging WAL untuk audit |
| JSON schema validation | Validasi field `config` & `content` |
| `VACUUM` improvements | Lebih efisien untuk tabel transaksi besar |

### Row Level Security

```sql
-- Aktifkan RLS di semua tabel utama setelah migrasi
ALTER TABLE "Student"     ENABLE ROW LEVEL SECURITY;
ALTER TABLE "Teacher"     ENABLE ROW LEVEL SECURITY;
ALTER TABLE "AuditLog"    ENABLE ROW LEVEL SECURITY;
-- dst untuk semua tabel yang punya schoolId

-- Policy: user hanya bisa lihat data schoolId mereka
CREATE POLICY school_isolation ON "Student"
  USING (school_id = current_setting('app.current_school_id')::text);

-- Set di setiap request
SET app.current_school_id = 'clxyz123...';
```

---

## Core Domain

### SaaS & Tenant

```prisma
model Tenant {
  id             String          @id @default(cuid())
  schoolId       String          @unique
  subdomain      String          @unique
  customDomain   String?         @unique
  tier           TenantTier      @default(FREE)
  status         TenantStatus    @default(TRIAL)
  trialEndsAt    DateTime?
  billingEmail   String
  features       TenantFeature[]
  subscription   Subscription?
  addons         TenantAddon[]
  createdAt      DateTime        @default(now())
  updatedAt      DateTime        @updatedAt
}

model TenantFeature {
  id           String      @id @default(cuid())
  tenantId     String
  tenant       Tenant      @relation(fields: [tenantId], references: [id])
  moduleId     String
  enabled      Boolean     @default(false)
  config       Json?
  overrideTier TenantTier?
  expiresAt    DateTime?
  updatedBy    String
  updatedAt    DateTime    @updatedAt
  @@unique([tenantId, moduleId])
}

model Subscription {
  id                 String             @id @default(cuid())
  tenantId           String             @unique
  tenant             Tenant             @relation(fields: [tenantId], references: [id])
  stripeSubId        String?            @unique
  tier               TenantTier
  status             SubscriptionStatus
  currentPeriodStart DateTime
  currentPeriodEnd   DateTime
  cancelAtPeriodEnd  Boolean            @default(false)
  invoices           Invoice[]
  createdAt          DateTime           @default(now())
}

model TenantAddon {
  id        String   @id @default(cuid())
  tenantId  String
  tenant    Tenant   @relation(fields: [tenantId], references: [id])
  addon     String   // 'AI_QUOTA_500', 'WA_QUOTA_1000', 'STORAGE_10GB'
  quantity  Int      @default(1)
  activeFrom DateTime
  activeTo  DateTime
  createdAt DateTime @default(now())
}

model Invoice {
  id             String      @id @default(cuid())
  subscriptionId String
  subscription   Subscription @relation(fields: [subscriptionId], references: [id])
  stripeInvoiceId String?    @unique
  amount         Decimal
  status         InvoiceStatus
  paidAt         DateTime?
  dueDate        DateTime
  period         String       // "2026-03"
  pdfUrl         String?
  createdAt      DateTime     @default(now())
}
```

### School & User

```prisma
model School {
  id           String    @id @default(cuid())
  name         String
  npsn         String    @unique
  address      String
  city         String
  province     String
  logoUrl      String?
  config       Json      @default("{}")
  students     Student[]
  teachers     Teacher[]
  createdAt    DateTime  @default(now())
  updatedAt    DateTime  @updatedAt
}

model User {
  id           String    @id @default(cuid())
  schoolId     String?
  email        String    @unique
  name         String
  role         UserRole
  passwordHash String
  photoUrl     String?
  mfaEnabled   Boolean   @default(false)
  mfaSecret    String?
  mfaMethod    String?
  isActive     Boolean   @default(true)
  lastLoginAt  DateTime?
  createdAt    DateTime  @default(now())
  updatedAt    DateTime  @updatedAt

  @@index([schoolId, role])
}

model RefreshToken {
  id          String   @id @default(cuid())
  userId      String
  token       String   @unique
  deviceInfo  String?
  ipAddress   String?
  expiresAt   DateTime
  revokedAt   DateTime?
  createdAt   DateTime @default(now())
  @@index([userId])
}
```

### Student & Guardian

```prisma
model Student {
  id              String            @id @default(cuid())
  schoolId        String
  school          School            @relation(fields: [schoolId], references: [id])
  userId          String?           @unique
  nisn            String            @unique
  nik             String?
  name            String
  gender          Gender
  birthDate       DateTime
  birthPlace      String?
  religion        Religion?
  photoUrl        String?
  address         String?
  guardians       Guardian[]
  enrollments     Enrollment[]
  savingsAccount  SavingsAccount?
  studentCard     StudentCard?
  healthProfile   HealthProfile?
  portfolio       Portfolio?
  specialNeeds    SpecialNeedsProfile?
  nutrition       NutritionRecord[]
  pointsAccount   StudentPoints?
  riskScore       Float?            @default(0)
  riskUpdatedAt   DateTime?
  createdAt       DateTime          @default(now())
  updatedAt       DateTime          @updatedAt

  @@index([schoolId, name])
  @@index([nisn])
}

model Guardian {
  id            String   @id @default(cuid())
  studentId     String
  student       Student  @relation(fields: [studentId], references: [id])
  schoolId      String
  relation      String   // "Ayah", "Ibu", "Wali"
  name          String
  phone         String
  email         String?
  occupation    String?
  income        String?  // kategori penghasilan (untuk beasiswa)
  isPrimary     Boolean  @default(false)
  userId        String?  // jika ortu punya akun EDS (portal ortu)
  createdAt     DateTime @default(now())
}

model Enrollment {
  id           String    @id @default(cuid())
  studentId    String
  student      Student   @relation(fields: [studentId], references: [id])
  schoolId     String
  classId      String
  class        Class     @relation(fields: [classId], references: [id])
  academicYear String    // "2025/2026"
  semester     Int?
  status       EnrollmentStatus @default(ACTIVE)
  startDate    DateTime
  endDate      DateTime?
  createdAt    DateTime  @default(now())

  @@unique([studentId, academicYear])
}

model Class {
  id           String      @id @default(cuid())
  schoolId     String
  name         String      // "10A", "XI IPA 2"
  gradeLevel   String      // "10", "11", "12"
  academicYear String
  waliKelasId  String?     // userId wali kelas
  students     Enrollment[]
  schedules    TeachingSchedule[]
  createdAt    DateTime    @default(now())

  @@unique([schoolId, name, academicYear])
}
```

### Attendance

```prisma
model StudentAttendance {
  id          String           @id @default(cuid())
  studentId   String
  student     Student          @relation(fields: [studentId], references: [id])
  schoolId    String
  classId     String
  date        DateTime
  status      AttendanceStatus
  method      String?          // 'MANUAL', 'QR', 'FACE'
  notes       String?
  recordedBy  String
  createdAt   DateTime         @default(now())

  @@unique([studentId, date])
  @@index([schoolId, date])
  @@index([classId, date])
}

enum AttendanceStatus { PRESENT ABSENT LATE SICK PERMISSION }
```

### Academic

```prisma
model Subject {
  id          String   @id @default(cuid())
  schoolId    String
  code        String
  name        String
  gradeLevel  String
  curriculum  CurriculumType
  hoursPerWeek Int
  isActive    Boolean  @default(true)
  @@unique([schoolId, code])
}

model TeachingSchedule {
  id          String   @id @default(cuid())
  schoolId    String
  classId     String
  class       Class    @relation(fields: [classId], references: [id])
  teacherId   String
  teacher     Teacher  @relation(fields: [teacherId], references: [id])
  subject     String
  dayOfWeek   Int      // 1=Senin, 6=Sabtu
  timeSlot    String   // "07:00-08:30"
  room        String?
  academicYear String
  semester    Int
  createdAt   DateTime @default(now())
}

model Grade {
  id           String   @id @default(cuid())
  studentId    String
  student      Student  @relation(fields: [studentId], references: [id])
  schoolId     String
  classId      String
  subject      String
  kdCode       String   // kode Kompetensi Dasar
  type         GradeType // DAILY | MID | FINAL | REMEDIAL
  score        Float
  academicYear String
  semester     Int
  teacherId    String
  createdAt    DateTime @default(now())

  @@index([studentId, subject, academicYear])
}

model ReportCard {
  id           String   @id @default(cuid())
  studentId    String
  student      Student  @relation(fields: [studentId], references: [id])
  schoolId     String
  classId      String
  academicYear String
  semester     Int
  grades       Json     // { subject: { finalScore, grade, teacherNote } }
  waaliNote    String?  // narasi wali kelas (bisa AI-generated)
  absenceCount Json     // { sick, permission, alpha }
  status       ReportStatus @default(DRAFT)
  publishedAt  DateTime?
  createdAt    DateTime @default(now())

  @@unique([studentId, academicYear, semester])
}

enum GradeType    { DAILY MID FINAL PROJECT REMEDIAL }
enum ReportStatus { DRAFT PUBLISHED }
```

### Exam (M05)

```prisma
model ExamQuestion {
  id          String       @id @default(cuid())
  schoolId    String
  subject     String
  topic       String
  gradeLevel  String
  curriculum  CurriculumType
  bloomLevel  BloomLevel
  difficulty  Difficulty
  type        QuestionType
  question    String
  options     Json?
  answer      String
  explanation String?
  generatedByAI Boolean   @default(true)
  createdBy   String
  usageCount  Int         @default(0)
  exams       ExamQuestionMap[]
  createdAt   DateTime    @default(now())

  @@index([schoolId, subject, gradeLevel])
}

model Exam {
  id          String          @id @default(cuid())
  schoolId    String
  title       String
  subject     String
  classId     String
  gradeLevel  String
  config      Json            // ExamConfig snapshot
  questions   ExamQuestionMap[]
  results     ExamResult[]
  startAt     DateTime?
  endAt       DateTime?
  status      ExamStatus      @default(DRAFT)
  createdBy   String
  createdAt   DateTime        @default(now())
}

model ExamQuestionMap {
  examId      String
  exam        Exam         @relation(fields: [examId], references: [id])
  questionId  String
  question    ExamQuestion @relation(fields: [questionId], references: [id])
  order       Int
  @@id([examId, questionId])
}

model ExamResult {
  id          String   @id @default(cuid())
  examId      String
  exam        Exam     @relation(fields: [examId], references: [id])
  studentId   String
  student     Student  @relation(fields: [studentId], references: [id])
  schoolId    String
  answers     Json     // { questionId: answer }
  score       Float?
  startedAt   DateTime?
  submittedAt DateTime?
  gradedAt    DateTime?
  gradedBy    String?  // userId atau 'AUTO'
  createdAt   DateTime @default(now())

  @@unique([examId, studentId])
}
```

### Finance

```prisma
model SavingsAccount {
  id          String        @id @default(cuid())
  studentId   String        @unique
  student     Student       @relation(fields: [studentId], references: [id])
  schoolId    String
  balance     Decimal       @default(0)
  minBalance  Decimal       @default(10000)
  transactions Transaction[]
  createdAt   DateTime      @default(now())
  updatedAt   DateTime      @updatedAt
}

model Transaction {
  id          String          @id @default(cuid())
  accountId   String
  account     SavingsAccount  @relation(fields: [accountId], references: [id])
  schoolId    String
  type        TransactionType
  amount      Decimal
  balanceAfter Decimal
  description String
  cashierId   String
  approvedBy  String?
  approvedAt  DateTime?
  vernonJournalId String?     // ID journal entry di Vernon Accounting
  createdAt   DateTime        @default(now())
}

model Invoice {
  id          String        @id @default(cuid())
  schoolId    String
  studentId   String
  student     Student       @relation(fields: [studentId], references: [id])
  type        InvoiceType   // SPP | PPDB | ACTIVITY | OTHER
  month       String?       // "2026-03" untuk SPP bulanan
  amount      Decimal
  dueDate     DateTime
  status      InvoiceStatus @default(UNPAID)
  paidAt      DateTime?
  payments    Payment[]
  vernonJournalId String?
  createdAt   DateTime      @default(now())

  @@index([schoolId, studentId, status])
}

model Payment {
  id          String        @id @default(cuid())
  invoiceId   String
  invoice     Invoice       @relation(fields: [invoiceId], references: [id])
  schoolId    String
  amount      Decimal
  channel     String        // 'QRIS', 'VA_BCA', 'GOPAY', 'CASH'
  gatewayRef  String?       // ID dari Midtrans/Xendit
  status      PaymentStatus @default(PENDING)
  paidAt      DateTime?
  receiptUrl  String?
  createdAt   DateTime      @default(now())
}

enum TransactionType { DEPOSIT WITHDRAWAL PURCHASE }
enum InvoiceType     { SPP PPDB ACTIVITY CANTEEN OTHER }
enum InvoiceStatus   { UNPAID PARTIAL PAID OVERDUE WAIVED }
enum PaymentStatus   { PENDING SUCCESS FAILED EXPIRED REFUNDED }
```

### Notifications (M14)

```prisma
model NotificationTemplate {
  id          String              @id @default(cuid())
  schoolId    String?             // null = template global EDS
  event       String              // 'ATTENDANCE_ABSENT', 'SPP_REMINDER', dll
  channel     NotificationChannel
  subject     String?             // untuk email
  body        String              // template dengan {{variabel}}
  isActive    Boolean             @default(true)
  createdAt   DateTime            @default(now())

  @@unique([schoolId, event, channel])
}

model NotificationLog {
  id          String              @id @default(cuid())
  schoolId    String
  recipientId String              // userId atau guardianId
  channel     NotificationChannel
  event       String
  subject     String?
  body        String              // pesan yang terkirim (setelah template di-render)
  status      NotificationStatus  @default(PENDING)
  gatewayRef  String?             // ID dari WA/FCM/SMTP
  errorMsg    String?
  sentAt      DateTime?
  readAt      DateTime?
  createdAt   DateTime            @default(now())

  @@index([schoolId, recipientId, createdAt])
}

enum NotificationChannel { WHATSAPP PUSH_NOTIFICATION EMAIL SMS }
enum NotificationStatus  { PENDING SENT FAILED READ }
```

### Listing Marketplace

```prisma
model ListingVendor {
  id           String          @id @default(cuid())
  businessName String
  category     ListingCategory
  description  String
  photos       String[]
  contactPhone String
  contactEmail String
  website      String?
  coverageArea String[]
  tier         ListingTier     @default(BASIC)
  status       ListingStatus   @default(PENDING_REVIEW)
  verifiedAt   DateTime?
  verifiedBy   String?
  placements   ListingPlacement[]
  billing      ListingBilling?
  reviews      ListingReview[]
  createdAt    DateTime        @default(now())
}

model ListingPlacement {
  id        String        @id @default(cuid())
  vendorId  String
  vendor    ListingVendor @relation(fields: [vendorId], references: [id])
  schoolId  String?       // null = semua sekolah di area coverage
  moduleId  String        // 'M11', 'M34', 'M19', dll
  position  Int           @default(0)
  isActive  Boolean       @default(true)
  clickCount  Int         @default(0)
  bookingCount Int        @default(0)
  createdAt DateTime      @default(now())
}

model ListingBilling {
  id        String        @id @default(cuid())
  vendorId  String        @unique
  vendor    ListingVendor @relation(fields: [vendorId], references: [id])
  tier      ListingTier
  monthlyFee Decimal
  status    BillingStatus @default(ACTIVE)
  renewedAt DateTime?
  expiresAt DateTime?
  createdAt DateTime      @default(now())
}

model ListingReview {
  id        String        @id @default(cuid())
  vendorId  String
  vendor    ListingVendor @relation(fields: [vendorId], references: [id])
  schoolId  String
  userId    String
  rating    Int           // 1-5
  comment   String?
  createdAt DateTime      @default(now())
  @@unique([vendorId, userId])
}

enum ListingCategory {
  GURU_LES TEMPAT_KURSUS ANTAR_JEMPUT KATERING
  PENERBIT SERAGAM VENDOR_EVENT ASURANSI TABUNGAN ALAT_TULIS
}
enum ListingTier   { BASIC PRO PREMIUM PER_EVENT }
enum ListingStatus { PENDING_REVIEW ACTIVE SUSPENDED EXPIRED REJECTED }
enum BillingStatus { ACTIVE PAST_DUE EXPIRED CANCELLED }
```

### Audit Log (M45)

```prisma
model AuditLog {
  id             String      @id @default(cuid())
  timestamp      DateTime    @default(now())
  schoolId       String?
  userId         String
  userRole       String
  module         String
  action         AuditAction
  entityType     String
  entityId       String
  oldValue       Json?
  newValue       Json?
  ipAddress      String
  userAgent      String
  sessionId      String?
  impersonatedBy String?
  isSensitive    Boolean     @default(false)
  requestId      String?

  @@index([schoolId, timestamp])
  @@index([userId, timestamp])
  @@index([entityType, entityId])
}
```

### Safety & Health Enums (dari M28, M29, M30)

```prisma
enum AuditAction {
  CREATE UPDATE DELETE VIEW EXPORT
  LOGIN LOGOUT LOGIN_FAILED
  IMPERSONATE_START IMPERSONATE_END
  FEATURE_TOGGLE TENANT_SUSPEND
  BULK_IMPORT BULK_EXPORT
  PASSWORD_CHANGE TWO_FA_ENABLED TWO_FA_DISABLED
}

enum BloomLevel     { C1 C2 C3 C4 C5 C6 }
enum Difficulty     { EASY MEDIUM HARD }
enum QuestionType   { MULTIPLE_CHOICE ESSAY TRUE_FALSE MATCHING }
enum ExamStatus     { DRAFT PUBLISHED ONGOING COMPLETED CANCELLED }
enum Gender         { MALE FEMALE }
enum Religion       { ISLAM CHRISTIAN CATHOLIC HINDU BUDDHA CONFUCIUS }
enum CurriculumType { K13 MERDEKA }
enum TenantTier     { FREE BASIC PRO ENTERPRISE }
enum TenantStatus   { TRIAL ACTIVE SUSPENDED CHURNED }
enum SubscriptionStatus { ACTIVE PAST_DUE CANCELED UNPAID }
enum InvoiceStatus2 { DRAFT OPEN PAID VOID UNCOLLECTIBLE }
enum UserRole {
  EDS_SUPERADMIN EDS_SUPPORT EDS_SALES LISTING_MANAGER
  YAYASAN_ADMIN ADMIN_SEKOLAH
  KEPALA_SEKOLAH KEPALA_KURIKULUM BENDAHARA
  GURU WALI_KELAS GURU_BK GPK
  OPERATOR_SIMS KASIR_KOPERASI PUSTAKAWAN
  PETUGAS_UKS TATA_USAHA SECURITY_OFFICER KOMITE_SEKOLAH
  SISWA ORANG_TUA ALUMNI GURU_LES
  MITRA_DU_DI PUBLISHER DEVELOPER VOLUNTEER LISTING_VENDOR
}
enum EnrollmentStatus { ACTIVE GRADUATED TRANSFERRED DROPPED }
enum NutritionStatus  { SEVERELY_WASTED WASTED NORMAL OVERWEIGHT OBESE }
enum BloodType        { A_POS A_NEG B_POS B_NEG AB_POS AB_NEG O_POS O_NEG UNKNOWN }
```

---

## Migrations

```bash
# Generate migration dari perubahan schema
pnpm --filter @eds/database prisma migrate dev --name "nama_perubahan"

# Apply ke staging/production
pnpm --filter @eds/database prisma migrate deploy

# Reset dev database (HATI-HATI: hapus semua data)
pnpm --filter @eds/database prisma migrate reset

# Introspect database yang sudah ada
pnpm --filter @eds/database prisma db pull
```

## Seed Data

```bash
# Seed data demo (2 sekolah, user-user default, modul config)
pnpm --filter @eds/database prisma db seed

# Seed custom per environment
DATABASE_SEED_ENV=staging pnpm --filter @eds/database prisma db seed
```

## Index Strategy

Semua tabel dengan `schoolId` harus punya index `[schoolId, createdAt]` sebagai minimum. Tabel yang sering diquery berdasarkan `studentId` ditambah index tersebut. Hindari index berlebihan pada tabel append-only seperti `AuditLog` dan `NotificationLog`.

```sql
-- Contoh index komposit penting
CREATE INDEX CONCURRENTLY idx_student_school_name
  ON "Student"(school_id, name);

CREATE INDEX CONCURRENTLY idx_attendance_school_date
  ON "StudentAttendance"(school_id, date DESC);

CREATE INDEX CONCURRENTLY idx_invoice_school_student_status
  ON "Invoice"(school_id, student_id, status);

-- Partial index untuk record aktif saja
CREATE INDEX CONCURRENTLY idx_active_enrollments
  ON "Enrollment"(school_id, class_id)
  WHERE status = 'ACTIVE';
```
