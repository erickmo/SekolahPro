# 08 — Pedagogis & Kurikulum Lanjutan (M48–M50, M55)

> **Terakhir diperbarui**: 20 Maret 2026

Kelompok ini mendukung siklus pembelajaran penuh sesuai Merdeka Belajar: RPP → Asesmen → Remedial. M55 ABK mendapat enkripsi data dan kontrol akses yang lebih ketat.

---

## M48 — RPP Digital
**Tier**: BASIC | **Integrasi**: M04, M31, M39

### Fitur
- [ ] Template RPP standar Kemendikbud (K13 & Merdeka Belajar)
- [ ] Generate draft RPP via Claude AI dari silabus + topik + KD
- [ ] Workflow: guru draft → kepala kurikulum review → approve
- [ ] Bank RPP sekolah: simpan, filter, fork (gunakan sebagai dasar)
- [ ] Version history per RPP (bisa revert ke versi sebelumnya)
- [ ] Sinkronisasi RPP ke kalender akademik M04
- [ ] Export RPP ke PDF (skill `pdf`) dan Word (skill `docx`)
- [ ] Dashboard supervisi: kepala kurikulum lihat status RPP semua guru

### AI Generate RPP

```typescript
// packages/ai/prompts/rpp.ts
export const RPP_SYSTEM_PROMPT = `
Kamu adalah ahli kurikulum Indonesia. Buat Rencana Pelaksanaan Pembelajaran
sesuai format resmi Kemendikbud untuk kurikulum {curriculum}.
Sertakan: tujuan pembelajaran, langkah kegiatan (pendahuluan/inti/penutup),
penilaian, dan sumber belajar.
Output hanya JSON dengan struktur RPPDocument.
`;

export async function generateRPP(params: {
  subject: string; topic: string; gradeLevel: string;
  curriculum: 'K13' | 'MERDEKA'; kdList: string[];
  durationMinutes: number;
}): Promise<RPPDocument> {
  const raw = await callClaude(
    RPP_SYSTEM_PROMPT.replace('{curriculum}', params.curriculum),
    JSON.stringify(params),
    { maxTokens: 2048, schoolId: params.schoolId, moduleId: 'M48' }
  );
  return JSON.parse(raw) as RPPDocument;
}
```

### Schema

```prisma
model LessonPlan {
  id            String          @id @default(cuid())
  schoolId      String
  teacherId     String
  teacher       Teacher         @relation(fields: [teacherId], references: [id])
  subject       String
  gradeLevel    String
  topic         String
  curriculum    CurriculumType
  durationMins  Int
  content       Json            // struktur RPP lengkap
  status        LessonPlanStatus @default(DRAFT)
  reviewedBy    String?
  reviewedAt    DateTime?
  reviewNotes   String?
  generatedByAI Boolean        @default(false)
  academicYear  String
  semester      Int
  versions      LessonPlanVersion[]
  createdAt     DateTime        @default(now())
  updatedAt     DateTime        @updatedAt
}

model LessonPlanVersion {
  id          String     @id @default(cuid())
  planId      String
  plan        LessonPlan @relation(fields: [planId], references: [id])
  version     Int
  content     Json       // snapshot konten saat versi ini dibuat
  savedBy     String
  savedAt     DateTime   @default(now())
}

enum CurriculumType   { K13 MERDEKA }
enum LessonPlanStatus { DRAFT SUBMITTED APPROVED REJECTED }
```

---

## M49 — Penilaian Proyek, KKTP & Asesmen Autentik
**Tier**: PRO | **Integrasi**: M04, M05, M37

### Fitur
- [ ] Builder rubrik penilaian proyek dengan kriteria & level kustom per guru
- [ ] Penilaian peer-to-peer: siswa menilai karya teman (dimoderasi guru)
- [ ] Observasi & checklist perilaku (penilaian sikap — Profil Pelajar Pancasila)
- [ ] KKTP: penetapan target capaian per tujuan pembelajaran & tracking progress
- [ ] Portofolio bukti belajar per KD (foto, video, dokumen) → terhubung ke M37
- [ ] Asesmen diagnostik awal semester (baseline kemampuan)
- [ ] Asesmen formatif berkala (bukan hanya ujian sumatif M05)
- [ ] Laporan capaian KKTP per siswa per mapel

### Schema

```prisma
model AssessmentRubric {
  id          String         @id @default(cuid())
  schoolId    String
  teacherId   String
  name        String
  subject     String
  gradeLevel  String
  criteria    RubricCriteria[]
  assessments ProjectAssessment[]
  createdAt   DateTime       @default(now())
}

model RubricCriteria {
  id          String           @id @default(cuid())
  rubricId    String
  rubric      AssessmentRubric @relation(fields: [rubricId], references: [id])
  name        String           // "Kualitas Isi", "Kreativitas", "Presentasi"
  weight      Float            // bobot persen (total semua criteria = 100)
  levels      Json             // [{ score: 4, desc: "Sangat Baik" }, ...]
}

model ProjectAssessment {
  id          String            @id @default(cuid())
  schoolId    String
  rubricId    String
  rubric      AssessmentRubric  @relation(fields: [rubricId], references: [id])
  studentId   String
  assessorId  String            // guru atau siswa (peer)
  assessorType AssessorType
  scores      Json              // { criteriaId: score, ... }
  totalScore  Float
  feedback    String?
  evidenceUrls String[]         // link ke portofolio M37
  assessedAt  DateTime          @default(now())
}

model KKTPTarget {
  id              String      @id @default(cuid())
  schoolId        String
  subject         String
  gradeLevel      String
  academicYear    String
  semester        Int
  learningGoals   KKTPGoal[]
  createdBy       String
  createdAt       DateTime    @default(now())
}

model KKTPGoal {
  id          String      @id @default(cuid())
  targetId    String
  target      KKTPTarget  @relation(fields: [targetId], references: [id])
  code        String      // "TP-1.1"
  description String
  minScore    Float       // skor minimum untuk dinyatakan tercapai
  progresses  KKTPProgress[]
}

model KKTPProgress {
  id          String    @id @default(cuid())
  goalId      String
  goal        KKTPGoal  @relation(fields: [goalId], references: [id])
  studentId   String
  score       Float
  achieved    Boolean
  assessedAt  DateTime
  evidence    String?
}

enum AssessorType { TEACHER PEER SELF }
```

---

## M50 — Program Remedial & Pengayaan
**Tier**: BASIC | **Integrasi**: M04, M05, M14, M23

### Fitur
- [ ] Identifikasi otomatis siswa butuh remedial (nilai < KKM setelah sumatif M04)
- [ ] Identifikasi siswa layak pengayaan (nilai ≥ 90)
- [ ] Penjadwalan sesi remedial & pengayaan di luar jam reguler
- [ ] Modul belajar remedial ditargetkan per KD yang belum tuntas
- [ ] Generate soal remedial via AI M05 (difficulty EASY, topik KD bermasalah)
- [ ] Tracking: ujian ulang → tuntas / perlu intervensi lanjut
- [ ] Laporan efektivitas program ke kepala kurikulum
- [ ] Notifikasi ke ortu saat anak dijadwalkan remedial (via M14)

### Schema

```prisma
model RemediationProgram {
  id            String              @id @default(cuid())
  schoolId      String
  subject       String
  gradeLevel    String
  classId       String
  academicYear  String
  semester      Int
  afterExamId   String              // FK ke Exam (M05) pemicu
  sessions      RemediationSession[]
  createdAt     DateTime            @default(now())
}

model RemediationSession {
  id            String              @id @default(cuid())
  programId     String
  program       RemediationProgram  @relation(fields: [programId], references: [id])
  schoolId      String
  type          RemediationType
  scheduledAt   DateTime
  teacherId     String
  targetKDs     String[]            // KD yang difokuskan
  participants  RemediationParticipant[]
  examId        String?             // ujian remedial (M05) terkait
  createdAt     DateTime            @default(now())
}

model RemediationParticipant {
  id          String             @id @default(cuid())
  sessionId   String
  session     RemediationSession @relation(fields: [sessionId], references: [id])
  studentId   String
  preScore    Float              // nilai sebelum remedial
  postScore   Float?             // nilai setelah remedial
  status      RemediationStatus  @default(ENROLLED)
  completedAt DateTime?
}

enum RemediationType   { REMEDIAL ENRICHMENT }
enum RemediationStatus { ENROLLED ATTENDED PASSED FAILED ABSENT }
```

### Trigger Otomatis
```typescript
// Dipanggil setelah hasil sumatif di-input di M04
export async function triggerRemediationCheck(examId: string, classId: string) {
  const results = await getExamResults(examId, classId);
  const kkm = await getKKM(results[0].subject, classId);

  const needRemedial  = results.filter(r => r.score < kkm);
  const needEnrichment = results.filter(r => r.score >= 90);

  if (needRemedial.length > 0) {
    await createRemediationProgram({ type: 'REMEDIAL',  students: needRemedial });
    await notifyParents(needRemedial, 'REMEDIAL_SCHEDULED'); // via M14
  }
  if (needEnrichment.length > 0) {
    await createRemediationProgram({ type: 'ENRICHMENT', students: needEnrichment });
  }
}
```

---

## M55 — Manajemen Siswa Berkebutuhan Khusus (ABK)
**Tier**: PRO | **Integrasi**: M01, M04, M05, M27, M28

### Fitur
- [ ] Profil kebutuhan khusus: diagnosis, kategori, akomodasi yang dibutuhkan
- [ ] IEP (Individualized Education Program) digital — target per semester
- [ ] Akomodasi ujian: waktu tambahan, format berbeda, pendamping (flag otomatis di M05)
- [ ] Koordinasi GPK (Guru Pendamping Khusus) — role terpisah di RBAC
- [ ] Tracking perkembangan ABK per semester dengan indikator yang disesuaikan
- [ ] Laporan perkembangan ke ortu & dinas PLB (skill `docx`)
- [ ] Saran akomodasi otomatis untuk guru baru yang mengajar siswa ABK (Claude AI)

### Privasi ABK
Data diagnosis ABK dienkripsi dengan **kunci terpisah** dari data siswa reguler. Key management menggunakan envelope encryption: data key per `schoolId` di-wrap dengan master key di KMS.

Akses **hanya untuk**: `GPK`, `GURU_BK`, `KEPALA_SEKOLAH`. Guru reguler tidak bisa lihat profil ABK, hanya dapat notifikasi "siswa ini butuh akomodasi tertentu saat ujian" tanpa detail diagnosis.

### Schema

```prisma
model SpecialNeedsProfile {
  id               String       @id @default(cuid())
  studentId        String       @unique
  student          Student      @relation(fields: [studentId], references: [id])
  schoolId         String
  // Data terenkripsi — tidak tersimpan plaintext
  encryptedData    Bytes        // AES-256-GCM envelope encryption
  // Field non-sensitif untuk keperluan akomodasi ujian
  requiresExtraTime Boolean     @default(false)
  extraTimeMinutes  Int?
  requiresAssistant Boolean     @default(false)
  requiresLargeFont Boolean     @default(false)
  ieps              IEP[]
  progressReports   ABKProgressReport[]
  createdAt         DateTime    @default(now())
  updatedAt         DateTime    @updatedAt
}

model IEP {
  id            String              @id @default(cuid())
  profileId     String
  profile       SpecialNeedsProfile @relation(fields: [profileId], references: [id])
  schoolId      String
  academicYear  String
  semester      Int
  goals         IEPGoal[]
  gpkId         String              // GPK yang menyusun
  approvedBy    String?             // kepala sekolah
  approvedAt    DateTime?
  createdAt     DateTime            @default(now())
}

model IEPGoal {
  id          String    @id @default(cuid())
  iepId       String
  iep         IEP       @relation(fields: [iepId], references: [id])
  domain      String    // "Akademik", "Motorik", "Komunikasi", "Sosial"
  targetDesc  String
  strategy    String
  evaluation  String    // cara mengukur capaian
  progress    Float?    // 0-100% capaian akhir semester
  achieved    Boolean   @default(false)
}

model ABKProgressReport {
  id          String              @id @default(cuid())
  profileId   String
  profile     SpecialNeedsProfile @relation(fields: [profileId], references: [id])
  schoolId    String
  period      String              // "2025/2026-1"
  summary     String
  iepProgress Json                // snapshot progress semua IEP goal
  gpkNotes    String?
  parentSharedAt DateTime?
  createdAt   DateTime            @default(now())
}
```

### AI Saran Akomodasi
```typescript
// Dipanggil saat guru baru pertama kali membuka absensi kelas yang ada ABK
export async function getAccommodationSuggestions(
  studentId: string, subject: string, teacherId: string
): Promise<string> {
  // Hanya baca field non-sensitif dari profil ABK
  const profile = await getABKNonSensitiveProfile(studentId);

  return callClaude(
    `Kamu adalah konsultan pendidikan inklusif Indonesia.
     Berikan saran praktis akomodasi pembelajaran untuk siswa dengan kebutuhan berikut
     dalam mata pelajaran ${subject}. Jangan menyebut diagnosis spesifik.
     Fokus pada strategi pedagogis.`,
    JSON.stringify(profile.accommodationNeeds),
    { maxTokens: 500, moduleId: 'M55' }
  );
}
```
