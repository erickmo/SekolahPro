# 11 — Integrasi Eksternal & Ekosistem (M40–M44, M59–M61)

> **Terakhir diperbarui**: 20 Maret 2026

---

## M40 — Dapodik / EMIS / ARKAS Integration
**Tier**: BASIC | **Integrasi**: M01, M04, M09, M31, M51

### Jadwal Sinkronisasi

```typescript
const SYNC_SCHEDULE = {
  dapodik_full:  'awal_semester',        // Jan & Jul: semua entitas
  dapodik_daily: ['attendance', 'enrollment'], // cron 23:00 WIB
  bos:           'tiap_triwulan',
  arkas:         'setelah_rkas_diapprove',
  emis:          'awal_semester',        // khusus madrasah
};
```

### Entitas yang Di-sync ke Dapodik

| EDS Entitas | Dapodik Entitas | Arah |
|-------------|----------------|------|
| `Student` | `Peserta Didik` | Dua arah |
| `Teacher` | `Pendidik & Tenaga Kependidikan` | Dua arah |
| `Enrollment` | `Rombongan Belajar` | EDS → Dapodik |
| `StaffAttendance` | `Kehadiran PTK` | EDS → Dapodik |
| `StudentAttendance` | `Kehadiran PD` | EDS → Dapodik |

### Schema

```prisma
model DapodikSyncLog {
  id          String      @id @default(cuid())
  schoolId    String
  entity      String      // 'Student', 'Teacher', dll
  direction   SyncDir     // PUSH | PULL
  recordCount Int
  successCount Int
  errorCount  Int
  errors      Json?       // detail error per record
  startedAt   DateTime
  completedAt DateTime?
  status      SyncStatus
}

enum SyncDir    { PUSH PULL BIDIRECTIONAL }
enum SyncStatus { RUNNING SUCCESS PARTIAL_FAIL FAILED }
```

### Error Handling
```typescript
// Konflik data: Dapodik dianggap source of truth untuk NISN & NUPTK
export async function resolveSyncConflict(
  entity: 'Student' | 'Teacher',
  localData: any, dapodikData: any
): Promise<any> {
  const merged = { ...localData, ...dapodikData };
  // Field yang Dapodik selalu menang: nisn, nuptk, nip, birthDate
  // Field yang EDS menang: email, phone, photoUrl (tidak ada di Dapodik)
  merged.nisn = dapodikData.nisn ?? localData.nisn;
  merged.nuptk = dapodikData.nuptk ?? localData.nuptk;
  return merged;
}
```

---

## M41 — LMS Integration
**Tier**: PRO | **Integrasi**: M04, M05

### Platform

| Platform | Integrasi | Metode |
|----------|-----------|--------|
| Google Classroom | Kelas, tugas, nilai | Google Classroom API v1 |
| Moodle | Kursus, modul, nilai | Moodle Web Services REST |
| Ruangguru | Akun premium siswa | SSO SAML 2.0 |

### SSO Flow (EDS sebagai Identity Provider)

```typescript
// EDS generate SAML assertion untuk platform eksternal
export async function generateSAMLAssertion(userId: string, platform: string) {
  const user = await getUser(userId);
  const attributes = {
    email: user.email,
    name: user.name,
    role: user.role,
    schoolId: user.schoolId,
    studentId: user.studentId ?? undefined,
    grade: user.currentGrade ?? undefined,
  };
  return buildSAMLResponse(attributes, PLATFORM_SP_CONFIG[platform]);
}
```

### Nilai Sync (Google Classroom → EDS)
```typescript
// Cron setiap malam: tarik nilai dari Google Classroom ke M04
export async function syncGradeFromGClassroom(schoolId: string) {
  const tokens = await getGoogleTokens(schoolId);
  const courses = await gClassroom.courses.list({ teacherId: 'me' });
  for (const course of courses.data.courses ?? []) {
    const submissions = await gClassroom.courses.courseWork.studentSubmissions.list({
      courseId: course.id!, courseWorkId: '-', states: ['TURNED_IN', 'RETURNED'],
    });
    await upsertGradesFromClassroom(schoolId, course, submissions.data.studentSubmissions);
  }
}
```

---

## M42 — Multi-tenant & White-label
**Tier**: ENTERPRISE | **Integrasi**: Developer Portal

### Subdomain Routing

```typescript
// middleware/tenant.ts — Next.js middleware
export async function middleware(req: NextRequest) {
  const host = req.headers.get('host') ?? '';
  const subdomain = host.split('.')[0];
  // eds.id → marketing, dev.eds.id → developer portal
  if (subdomain === 'eds' || subdomain === 'dev') return NextResponse.next();

  const tenant = await getTenantBySubdomain(subdomain);
  if (!tenant) return NextResponse.rewrite(new URL('/404', req.url));
  if (tenant.status === 'SUSPENDED') return NextResponse.rewrite(new URL('/suspended', req.url));

  // Inject schoolId ke header untuk semua downstream requests
  const res = NextResponse.next();
  res.headers.set('x-school-id', tenant.schoolId);
  res.headers.set('x-tenant-tier', tenant.tier);
  return res;
}
```

### Custom Domain (ENTERPRISE)
```
Sekolah ingin: siakad.smanmalang.sch.id
Steps:
1. Sekolah set CNAME: siakad.smanmalang.sch.id → eds.id
2. Admin EDS verifikasi domain di Developer Portal
3. SSL auto-provision via Let's Encrypt (Caddy / cert-manager)
4. Tenant record diupdate: customDomain = "siakad.smanmalang.sch.id"
```

---

## M43 — Dashboard Yayasan
**Tier**: ENTERPRISE | **Integrasi**: M12 (tiap unit)

### Fitur
- [ ] Agregasi KPI dari semua unit sekolah (siswa aktif, kehadiran, rata2 nilai, SPP collected)
- [ ] Perbandingan performa antar unit (bar chart, ranking)
- [ ] Alokasi & monitoring anggaran per unit
- [ ] Push kebijakan terpusat ke semua unit (PPDB config, kurikulum, dll)
- [ ] Laporan konsolidasi untuk pengurus yayasan (skill `xlsx`, skill `pdf`)
- [ ] Alert otomatis jika unit perlu perhatian (tunggakan SPP tinggi, kehadiran turun)

### Schema

```prisma
model Foundation {
  id        String   @id @default(cuid())
  name      String
  schools   FoundationSchool[]
  admins    FoundationAdmin[]
  createdAt DateTime @default(now())
}

model FoundationSchool {
  id           String     @id @default(cuid())
  foundationId String
  foundation   Foundation @relation(fields: [foundationId], references: [id])
  schoolId     String     @unique
  joinedAt     DateTime   @default(now())
}

model FoundationAdmin {
  id           String     @id @default(cuid())
  foundationId String
  foundation   Foundation @relation(fields: [foundationId], references: [id])
  userId       String     @unique
  role         String     // 'YAYASAN_ADMIN' | 'YAYASAN_VIEWER'
}
```

---

## M44 — Open API & Developer Marketplace
**Tier**: ENTERPRISE | **Integrasi**: Developer Portal

### API Endpoints Publik

```
GET  /api/v1/schools/:schoolId/info       — info publik sekolah
GET  /api/v1/students/:nisn               — cek status siswa (auth required)
POST /api/v1/webhooks/register            — daftarkan webhook endpoint
GET  /api/v1/webhooks/events              — daftar event yang tersedia
```

### Webhook Events

```typescript
type WebhookEvent =
  | 'student.enrolled'      | 'student.graduated'
  | 'exam.completed'        | 'payment.received'
  | 'attendance.marked'     | 'alert.ews_triggered'
  | 'report.generated'      | 'listing.new_booking'
```

### Rate Limiting

```typescript
const API_RATE_LIMITS = {
  FREE_TIER_DEV:    { rpm: 60,   daily: 1_000 },
  PAID_DEV_BASIC:   { rpm: 300,  daily: 10_000 },
  PAID_DEV_PRO:     { rpm: 1000, daily: 100_000 },
  ENTERPRISE:       { rpm: 5000, daily: -1 },  // unlimited
};
```

---

## M59 — IoT & Smart Campus
**Tier**: ENTERPRISE (butuh hardware) | **Integrasi**: M04, M35, M51

### Perangkat yang Di-integrasikan

| Sensor | Lokasi | Data yang Dikumpulkan | Aksi |
|--------|--------|----------------------|------|
| Suhu & kelembaban | Ruang kelas | °C, %RH per 5 menit | Alert jika >32°C |
| Smart meter listrik | Panel utama | kWh per jam | Laporan ke M51 |
| Smart meter air | Tandon | Liter per hari | Alert kebocoran |
| Sensor gerak | Lab & perpustakaan | Kehadiran tanpa scan | Log kehadiran |
| Sensor pintu | Server room, gudang | Buka/tutup + siapa | Alert akses janggal |
| Panic button | UKS, ruang BK | Tekan | Alert darurat ke piket |

### Protokol

```
Hardware Device → MQTT Broker (Mosquitto) → Node-RED → EDS API
                                               ↓
                                          M35 Helpdesk (buat tiket otomatis)
                                          M12 Dashboard (realtime widget)
                                          M51 RKAS (data konsumsi energi)
```

### Schema

```prisma
model IoTDevice {
  id          String        @id @default(cuid())
  schoolId    String
  name        String
  type        IoTDeviceType
  location    String
  mqttTopic   String        @unique
  isActive    Boolean       @default(true)
  lastPing    DateTime?
  readings    IoTReading[]
}

model IoTReading {
  id        String    @id @default(cuid())
  deviceId  String
  device    IoTDevice @relation(fields: [deviceId], references: [id])
  schoolId  String
  value     Float
  unit      String    // "°C", "%RH", "kWh", "L"
  recordedAt DateTime @default(now())

  @@index([deviceId, recordedAt])
}

enum IoTDeviceType {
  TEMPERATURE_HUMIDITY SMART_METER_ELECTRIC SMART_METER_WATER
  MOTION_SENSOR DOOR_SENSOR PANIC_BUTTON
}
```

---

## M60 — Gamifikasi & Program Loyalitas
**Tier**: PRO | **Integrasi**: M01, M14, M37

### Sistem Poin

```typescript
export const POINT_RULES: PointRule[] = [
  { event: 'ATTENDANCE_ON_TIME',        points: 10, maxPerDay: 1 },
  { event: 'HOMEWORK_SUBMITTED',        points: 15, maxPerDay: null },
  { event: 'EXAM_SCORE_90_PLUS',        points: 50, maxPerExam: 1 },
  { event: 'LIBRARY_BOOK_READ',         points: 5,  maxPerDay: 3 },
  { event: 'FORUM_POST_QUALITY',        points: 3,  maxPerDay: 5 },
  { event: 'EXTRACURRICULAR_ATTENDED',  points: 10, maxPerDay: 1 },
  { event: 'REMEDIAL_PASSED',           points: 30, maxPerSession: 1 },
];
```

### Schema

```prisma
model StudentPoints {
  id          String   @id @default(cuid())
  studentId   String
  student     Student  @relation(fields: [studentId], references: [id])
  schoolId    String
  totalPoints Int      @default(0)
  streak      Int      @default(0)    // hari hadir berturut-turut
  lastStreakDate DateTime?
  transactions PointTransaction[]
  badges      StudentBadge[]
  updatedAt   DateTime @updatedAt
}

model PointTransaction {
  id          String       @id @default(cuid())
  studentId   String
  schoolId    String
  points      Int          // positif (earn) atau negatif (redeem)
  event       String       // 'ATTENDANCE_ON_TIME', 'REDEEM_REWARD', dll
  sourceId    String?      // ID sumber (examId, attendanceId, dll)
  description String
  createdAt   DateTime     @default(now())
}

model Badge {
  id          String         @id @default(cuid())
  schoolId    String
  name        String
  description String
  iconUrl     String
  criteria    Json           // { type: 'STREAK', value: 30 }
  students    StudentBadge[]
}

model StudentBadge {
  id          String   @id @default(cuid())
  studentId   String
  points      StudentPoints @relation(fields: [studentId], references: [studentId])
  badgeId     String
  badge       Badge    @relation(fields: [badgeId], references: [id])
  earnedAt    DateTime @default(now())
  @@unique([studentId, badgeId])
}
```

---

## M61 — Marketplace Konten Edukatif
**Tier**: LISTING (revenue share 70/30) | **Integrasi**: M05, M48, M49

### Schema

```prisma
model EduContent {
  id            String          @id @default(cuid())
  authorId      String          // teacherId atau alumniId
  schoolOrigin  String          // schoolId asal penulis
  title         String
  description   String
  type          ContentType
  subject       String
  gradeLevel    String
  curriculum    CurriculumType
  price         Decimal         // 0 untuk gratis
  licenseType   LicenseType
  fileUrl       String?         // konten utama di MinIO
  previewUrl    String?
  thumbnailUrl  String?
  status        ContentStatus   @default(PENDING_REVIEW)
  reviewedBy    String?
  reviewedAt    DateTime?
  reviewNotes   String?
  aiQualityScore Float?         // 0-1, dari Claude review
  downloads     Int             @default(0)
  rating        Float?
  purchases     ContentPurchase[]
  reviews       ContentReview[]
  createdAt     DateTime        @default(now())
}

model ContentPurchase {
  id          String     @id @default(cuid())
  contentId   String
  content     EduContent @relation(fields: [contentId], references: [id])
  schoolId    String     // sekolah pembeli
  buyerId     String     // userId yang membeli
  amount      Decimal
  royaltyPaid Decimal    // 70% dari amount
  paymentId   String     // FK ke M18 Payment
  purchasedAt DateTime   @default(now())
}

enum ContentType   { MODULE VIDEO RPP RUBRIC EXAM_PACK OTHER }
enum LicenseType   { SINGLE_SCHOOL MULTI_SCHOOL UNLIMITED }
enum ContentStatus { PENDING_REVIEW APPROVED REJECTED SUSPENDED }
```

### AI Quality Review
```typescript
export async function reviewContentQuality(contentId: string): Promise<number> {
  const content = await getContentForReview(contentId);
  const response = await callClaude(
    `Kamu adalah reviewer konten edukatif Indonesia. Nilai kualitas konten berikut
     dari 0.0 hingga 1.0 berdasarkan: relevansi kurikulum, kedalaman materi,
     kejelasan penyajian, dan potensi plagiarisme. Jawab HANYA dengan angka desimal.`,
    JSON.stringify({ type: content.type, title: content.title, preview: content.previewText }),
    { maxTokens: 10, moduleId: 'M61' }
  );
  return parseFloat(response.trim());
}
```
