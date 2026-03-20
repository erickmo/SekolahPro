# 06 — Kesehatan & Kesejahteraan (M27–M30)

> **Terakhir diperbarui**: 20 Maret 2026

Data kesehatan mendapat enkripsi khusus terpisah dari data siswa reguler. Akses dikontrol ketat — guru reguler tidak punya akses ke rekam medis.

---

## M27 — UKS & Rekam Medis Digital
**Tier**: BASIC | **Integrasi**: M01, M14, M28, M30

### Fitur
- [ ] Rekam kesehatan per siswa: golongan darah, alergi, kondisi kronis, riwayat vaksinasi
- [ ] Log kunjungan UKS: tanggal, keluhan, tindakan, obat diberikan, petugas
- [ ] Notifikasi realtime ke ortu saat siswa ke UKS dengan ringkasan kondisi (via M14)
- [ ] Rekap kunjungan untuk laporan bulanan ke puskesmas mitra
- [ ] Jadwal imunisasi & skrining berkala (integrasi kalender M04)
- [ ] Surat pengantar rujukan ke puskesmas/RS — auto-isi data siswa (skill `docx`)
- [ ] Dashboard tren: penyakit paling sering, puncak kunjungan per bulan
- [ ] Stok obat & alat UKS dengan alert minimum stok

### Schema

```prisma
model HealthProfile {
  id            String        @id @default(cuid())
  studentId     String        @unique
  student       Student       @relation(fields: [studentId], references: [id])
  schoolId      String
  bloodType     BloodType?
  allergies     String[]
  chronicConds  String[]
  vaccinations  Vaccination[]
  uksVisits     UKSVisit[]
  encryptedData Bytes?        // AES-256-GCM, kunci per schoolId
  createdAt     DateTime      @default(now())
  updatedAt     DateTime      @updatedAt
}

model UKSVisit {
  id             String        @id @default(cuid())
  profileId      String
  profile        HealthProfile @relation(fields: [profileId], references: [id])
  schoolId       String
  visitedAt      DateTime      @default(now())
  complaint      String
  diagnosis      String?
  treatment      String
  medicinesGiven String[]
  referredTo     String?
  staffId        String
  parentNotified Boolean       @default(false)
  notifiedAt     DateTime?
}

model Vaccination {
  id          String        @id @default(cuid())
  profileId   String
  profile     HealthProfile @relation(fields: [profileId], references: [id])
  schoolId    String
  vaccineName String
  doseNumber  Int
  givenAt     DateTime
  nextDueAt   DateTime?
  batchNumber String?
}

enum BloodType { A_POS A_NEG B_POS B_NEG AB_POS AB_NEG O_POS O_NEG UNKNOWN }
```

### Akses Kontrol
Hanya `PETUGAS_UKS` (CRUD), `GURU_BK` (read), `KEPALA_SEKOLAH` (agregat), `ORANG_TUA` (anak sendiri).

### API
```
GET    /api/v1/uks/profile/:studentId
POST   /api/v1/uks/visits
GET    /api/v1/uks/visits?studentId=&month=
GET    /api/v1/uks/dashboard
POST   /api/v1/uks/referral/:visitId        — generate surat rujukan
GET    /api/v1/uks/report/monthly           — export PDF/Excel
```

---

## M28 — Konseling & BK Digital
**Tier**: BASIC | **Integrasi**: M23, M27, M29, M55

### Fitur
- [ ] Booking sesi — siswa self-book, guru referral, atau EWS otomatis (M23)
- [ ] Catatan sesi terenkripsi AES-256 (hanya BK & kepsek — kepsek tidak bisa baca isi)
- [ ] Tagging topik: akademik, sosial, keluarga, karir, mental health, bullying
- [ ] Timeline riwayat sesi per siswa
- [ ] Program bimbingan karir kelas XII: minat, bakat, rekomendasi jurusan
- [ ] Surat rujukan ke psikolog/konselor profesional (skill `docx`)
- [ ] Dashboard BK: siswa butuh follow-up, distribusi topik, tren
- [ ] Alert dari M23 EWS otomatis buat task di inbox BK

### Schema

```prisma
model CounselingSession {
  id             String            @id @default(cuid())
  studentId      String
  student        Student           @relation(fields: [studentId], references: [id])
  schoolId       String
  counselorId    String
  scheduledAt    DateTime
  actualStart    DateTime?
  actualEnd      DateTime?
  status         SessionStatus     @default(SCHEDULED)
  topics         CounselingTopic[]
  encryptedNotes Bytes             // AES-256-GCM
  followUpDate   DateTime?
  referralNote   String?
  sourceAlertId  String?           // ewsAlertId dari M23
  createdAt      DateTime          @default(now())
}

model CounselingTopic {
  id          String            @id @default(cuid())
  sessionId   String
  session     CounselingSession @relation(fields: [sessionId], references: [id])
  category    TopicCategory
}

enum SessionStatus { SCHEDULED COMPLETED CANCELLED NO_SHOW }
enum TopicCategory {
  ACADEMIC SOCIAL FAMILY CAREER MENTAL_HEALTH
  BULLYING SUBSTANCE SELF_HARM OTHER
}
```

### Business Rules
- Catatan konseling tidak bisa dibaca kepala sekolah tanpa persetujuan BK
- Jika topik `SELF_HARM` atau `SUBSTANCE` → alert otomatis ke kepsek (tanpa isi catatan)
- Sesi dari M23 EWS ditandai `sourceAlertId` untuk ukur efektivitas intervensi

---

## M29 — Anti-Bullying & Safety Reporting
**Tier**: FREE — **tidak bisa dimatikan** | **Integrasi**: M28, M12, M14

### Fitur
- [ ] Form pelaporan anonim multi-langkah (siswa, guru, atau ortu)
- [ ] Token unik untuk pelapor pantau progress tanpa membuka identitas
- [ ] Kategorisasi: bullying fisik, verbal, cyberbullying, pelecehan seksual, diskriminasi
- [ ] Eskalasi otomatis: BK (24 jam) → Kepsek (48 jam) → Dinas (72 jam)
- [ ] Timeline penanganan transparan untuk pelapor via token
- [ ] Dashboard kasus di M12 (tren bulanan, kategori, waktu penanganan rata-rata)
- [ ] Konten edukasi anti-bullying yang bisa dikurasi per sekolah

### Schema

```prisma
model SafetyReport {
  id             String             @id @default(cuid())
  schoolId       String
  trackingToken  String             @unique @default(cuid())
  reporterType   ReporterType
  reporterUserId String?
  category       IncidentCategory
  description    String
  incidentDate   DateTime?
  location       String?
  status         ReportStatus       @default(RECEIVED)
  priority       ReportPriority     @default(MEDIUM)
  assignedTo     String?
  timeline       ReportTimeline[]
  escalations    ReportEscalation[]
  createdAt      DateTime           @default(now())
  updatedAt      DateTime           @updatedAt
}

model ReportTimeline {
  id        String       @id @default(cuid())
  reportId  String
  report    SafetyReport @relation(fields: [reportId], references: [id])
  action    String
  actor     String?
  note      String?
  createdAt DateTime     @default(now())
}

model ReportEscalation {
  id          String       @id @default(cuid())
  reportId    String
  report      SafetyReport @relation(fields: [reportId], references: [id])
  level       Int          // 1=BK, 2=Kepsek, 3=Dinas
  escalatedAt DateTime     @default(now())
  reason      String
  notifiedTo  String[]
}

enum ReporterType     { STUDENT TEACHER PARENT ANONYMOUS }
enum IncidentCategory { PHYSICAL_BULLYING VERBAL_BULLYING CYBERBULLYING
                        SEXUAL_HARASSMENT DISCRIMINATION OTHER }
enum ReportStatus     { RECEIVED INVESTIGATING IN_PROGRESS RESOLVED CLOSED }
enum ReportPriority   { LOW MEDIUM HIGH CRITICAL }
```

### Eskalasi Otomatis
```typescript
// Cron job tiap jam
export async function checkSafetyReportSLA() {
  const now = new Date();
  // Level 1: belum ada respons BK dalam 24 jam
  await escalateOverdue(1, subHours(now, 24));
  // Level 2: BK sudah respons tapi belum resolved dalam 48 jam
  await escalateOverdue(2, subHours(now, 48));
  // Level 3: belum resolved dalam 72 jam → notifikasi ke dinas
  await escalateOverdue(3, subHours(now, 72));
}
```

---

## M30 — Pemantauan Gizi Siswa
**Tier**: BASIC | **Integrasi**: M01, M19, M27

### Fitur
- [ ] Input antropometri per semester: BB, TB, IMT otomatis, z-score WHO
- [ ] Grafik pertumbuhan per siswa (standar WHO 2007)
- [ ] Klasifikasi: gizi buruk / kurang / baik / lebih / obesitas
- [ ] Alert ke wali kelas jika ada siswa status gizi berisiko
- [ ] Laporan gizi ke puskesmas (format standar Kemenkes) — skill `xlsx`
- [ ] PMT (Pemberian Makanan Tambahan): daftar penerima, distribusi, tracking
- [ ] Rekomendasi menu ke M19 (Kantin) berdasarkan profil gizi kelas

### Schema

```prisma
model NutritionRecord {
  id              String          @id @default(cuid())
  studentId       String
  student         Student         @relation(fields: [studentId], references: [id])
  schoolId        String
  measuredAt      DateTime
  measuredBy      String
  weightKg        Float
  heightCm        Float
  bmi             Float
  bmiZScore       Float?
  nutritionStatus NutritionStatus
  academicYear    String
  semester        Int
  notes           String?
  createdAt       DateTime        @default(now())
}

model PMTProgram {
  id             String            @id @default(cuid())
  schoolId       String
  name           String
  startDate      DateTime
  endDate        DateTime
  targetStudents String[]
  menuPlan       Json
  distributions  PMTDistribution[]
  createdAt      DateTime          @default(now())
}

model PMTDistribution {
  id            String     @id @default(cuid())
  programId     String
  program       PMTProgram @relation(fields: [programId], references: [id])
  distributedAt DateTime
  studentId     String
  menuItem      String
  received      Boolean    @default(true)
}

enum NutritionStatus {
  SEVERELY_WASTED WASTED NORMAL OVERWEIGHT OBESE
}
```

### Kalkulasi BMI & Z-Score
```typescript
export function calcNutritionStatus(
  weightKg: number, heightCm: number,
  ageMonths: number, gender: 'MALE' | 'FEMALE'
) {
  const bmi = weightKg / Math.pow(heightCm / 100, 2);
  const zScore = lookupWHO2007ZScore(bmi, ageMonths, gender);
  const status =
    zScore < -3 ? 'SEVERELY_WASTED' :
    zScore < -2 ? 'WASTED' :
    zScore <  1 ? 'NORMAL' :
    zScore <  2 ? 'OVERWEIGHT' : 'OBESE';
  return { bmi: +bmi.toFixed(1), zScore: +zScore.toFixed(2), status };
}
```
