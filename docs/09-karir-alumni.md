# 09 — Karir, Alumni & Pengembangan (M36–M39)

> **Terakhir diperbarui**: 20 Maret 2026

---

## M36 — Alumni & Tracer Study
**Tier**: BASIC | **Integrasi**: M01, M37, M54

### Fitur
- [ ] Database alumni otomatis terbentuk saat siswa lulus (dari M01)
- [ ] Survei tracer study otomatis: 1, 3, 5 tahun setelah lulus (via M14 WA/email)
- [ ] Program mentoring: alumni profesional mendaftar sebagai mentor adik kelas
- [ ] Portal lowongan pekerjaan / magang dari alumni & mitra industri
- [ ] Event reunian: alumni RSVP, bayar tiket via M18
- [ ] Kontribusi alumni: donasi, beasiswa, sponsor kegiatan → M52

### Schema

```prisma
model AlumniProfile {
  id            String         @id @default(cuid())
  studentId     String         @unique
  student       Student        @relation(fields: [studentId], references: [id])
  schoolId      String
  graduationYear Int
  major         String?        // jurusan saat SMA (IPA/IPS/SMK)
  currentEdu    String?        // PT yang dimasuki setelah lulus
  currentJob    String?
  currentCompany String?
  city          String?
  linkedinUrl   String?
  isMentor      Boolean        @default(false)
  mentorSubjects String[]      // mata pelajaran / bidang yang mau di-mentor
  tracerStudies AlumniTracerStudy[]
  createdAt     DateTime       @default(now())
  updatedAt     DateTime       @updatedAt
}

model AlumniTracerStudy {
  id            String        @id @default(cuid())
  alumniId      String
  alumni        AlumniProfile @relation(fields: [alumniId], references: [id])
  schoolId      String
  surveyYear    Int           // tahun pelaksanaan survei
  yearsSince    Int           // 1 | 3 | 5
  employmentStatus EmploymentStatus
  currentActivity String?
  monthlyIncome IncomeRange?
  relevance     Int?          // 1-5: relevansi pendidikan SMA dengan karir
  feedback      String?
  submittedAt   DateTime      @default(now())
}

model JobPosting {
  id          String   @id @default(cuid())
  schoolId    String
  postedBy    String   // alumniId atau adminId
  title       String
  company     String
  type        JobType
  location    String
  description String
  applyUrl    String?
  applyEmail  String?
  deadline    DateTime?
  isActive    Boolean  @default(true)
  createdAt   DateTime @default(now())
}

enum EmploymentStatus { EMPLOYED SELF_EMPLOYED STUDYING UNEMPLOYED OTHER }
enum IncomeRange      { BELOW_3M RANGE_3_5M RANGE_5_10M ABOVE_10M }
enum JobType          { FULL_TIME PART_TIME INTERNSHIP FREELANCE }
```

---

## M37 — Portofolio Digital Siswa
**Tier**: FREE | **Integrasi**: M06, M38, M49

E-portfolio yang dapat dibagikan ke institusi lanjutan atau perusahaan.

### Fitur
- [ ] Otomatis terpopulasi dari: nilai rapor (M04), ekskul (M06), ujian (M05), asesmen proyek (M49)
- [ ] Upload karya manual: dokumen, gambar, video, link
- [ ] Sertifikat & penghargaan (upload scan + verifikasi guru)
- [ ] Endorsement dari guru (teks singkat, tampil di portofolio)
- [ ] Shareable public link: `eds.id/portfolio/[slug]`
- [ ] Privacy control: pilih item mana yang publik / hanya untuk sekolah / privat
- [ ] Export portofolio sebagai PDF (skill `pdf`)

### Schema

```prisma
model Portfolio {
  id          String           @id @default(cuid())
  studentId   String           @unique
  student     Student          @relation(fields: [studentId], references: [id])
  schoolId    String
  slug        String           @unique   // URL-friendly username
  bio         String?
  isPublic    Boolean          @default(false)
  items       PortfolioItem[]
  endorsements PortfolioEndorsement[]
  views       Int              @default(0)
  createdAt   DateTime         @default(now())
  updatedAt   DateTime         @updatedAt
}

model PortfolioItem {
  id          String        @id @default(cuid())
  portfolioId String
  portfolio   Portfolio     @relation(fields: [portfolioId], references: [id])
  type        PortfolioItemType
  title       String
  description String?
  fileUrl     String?
  linkUrl     String?
  thumbnailUrl String?
  isPublic    Boolean       @default(true)
  sourceModule String?      // 'M04', 'M05', 'M06', dll — jika auto-populated
  sourceId    String?       // ID di modul sumber
  date        DateTime?
  order       Int           @default(0)
  createdAt   DateTime      @default(now())
}

model PortfolioEndorsement {
  id          String    @id @default(cuid())
  portfolioId String
  portfolio   Portfolio @relation(fields: [portfolioId], references: [id])
  teacherId   String
  message     String    // max 280 karakter
  subject     String?
  createdAt   DateTime  @default(now())
}

enum PortfolioItemType {
  ACADEMIC_ACHIEVEMENT CREATIVE_WORK PROJECT CERTIFICATE
  EXTRACURRICULAR INTERNSHIP OTHER
}
```

---

## M38 — Prakerin / Magang Siswa (SMK)
**Tier**: PRO | **Integrasi**: M01, M37

### Fitur
- [ ] Database DU/DI mitra (Dunia Usaha/Industri) — bisa dari Listing M10
- [ ] Matching siswa ke tempat prakerin berdasarkan kompetensi keahlian
- [ ] Jurnal harian digital: siswa input aktivitas, pembimbing sekolah pantau
- [ ] Nilai dari pembimbing DU/DI via portal mitra (role `MITRA_DU_DI`)
- [ ] Absensi di tempat prakerin (siswa check-in manual, PIC konfirmasi)
- [ ] Laporan akhir prakerin (upload PDF)
- [ ] Sertifikat prakerin digital — di-generate otomatis + tanda tangan digital (skill `pdf`)
- [ ] Hasil prakerin auto-masuk ke portofolio M37

### Schema

```prisma
model InternshipPartner {
  id          String          @id @default(cuid())
  schoolId    String
  companyName String
  industry    String
  address     String
  city        String
  picName     String
  picEmail    String
  picPhone    String
  capacity    Int             // maks siswa per periode
  competencies String[]       // kompetensi keahlian yang relevan
  isActive    Boolean         @default(true)
  placements  InternshipPlacement[]
  createdAt   DateTime        @default(now())
}

model InternshipPlacement {
  id            String             @id @default(cuid())
  schoolId      String
  studentId     String
  student       Student            @relation(fields: [studentId], references: [id])
  partnerId     String
  partner       InternshipPartner  @relation(fields: [partnerId], references: [id])
  supervisorId  String             // teacherId pembimbing dari sekolah
  startDate     DateTime
  endDate       DateTime
  status        InternshipStatus   @default(ACTIVE)
  finalScore    Float?
  certificateUrl String?
  journals      InternshipJournal[]
  createdAt     DateTime           @default(now())
}

model InternshipJournal {
  id           String             @id @default(cuid())
  placementId  String
  placement    InternshipPlacement @relation(fields: [placementId], references: [id])
  date         DateTime
  activities   String
  learnings    String?
  photos       String[]
  supervisorNote String?
  supervisorApproved Boolean     @default(false)
  studentSubmittedAt DateTime    @default(now())
}

enum InternshipStatus { ACTIVE COMPLETED WITHDRAWN }
```

---

## M39 — Pelatihan & CPD Guru
**Tier**: BASIC | **Integrasi**: M31, M48, module-vernon-hrm

### Fitur
- [ ] Katalog pelatihan: internal (sekolah) & eksternal (Kemendikbud, P4TK, lembaga lain)
- [ ] Pendaftaran online & approval kepala sekolah
- [ ] Tracking kehadiran & jam pelatihan (sync ke Vernon HRM sebagai work log)
- [ ] Rekam sertifikasi & kompetensi hasil pelatihan per guru
- [ ] Poin CPD (Continuing Professional Development) — terakumulasi per tahun
- [ ] Refleksi & jurnal mengajar digital pasca-pelatihan
- [ ] Rekomendasi pelatihan berdasarkan hasil PKG (M31) dan gap kompetensi RPP (M48)

### Schema

```prisma
model TrainingCourse {
  id            String         @id @default(cuid())
  schoolId      String
  title         String
  organizer     String         // "P4TK", "Sekolah Internal", dll
  type          TrainingType
  mode          TrainingMode   // ONLINE | OFFLINE | HYBRID
  startDate     DateTime
  endDate       DateTime
  hours         Float          // jam pelatihan
  maxParticipants Int?
  registrations TrainingRegistration[]
  createdAt     DateTime       @default(now())
}

model TrainingRegistration {
  id            String         @id @default(cuid())
  courseId      String
  course        TrainingCourse @relation(fields: [courseId], references: [id])
  teacherId     String
  teacher       Teacher        @relation(fields: [teacherId], references: [id])
  schoolId      String
  status        RegStatus      @default(PENDING)
  approvedBy    String?
  attended      Boolean        @default(false)
  certificateUrl String?
  reflection    String?
  cpd_points    Float?         // dihitung dari hours * faktor jenis pelatihan
  completedAt   DateTime?
  createdAt     DateTime       @default(now())
}

enum TrainingType { SUBJECT_CONTENT PEDAGOGY CURRICULUM TECHNOLOGY LEADERSHIP OTHER }
enum TrainingMode { ONLINE OFFLINE HYBRID }
enum RegStatus    { PENDING APPROVED REJECTED COMPLETED CANCELLED }
```
