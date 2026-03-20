# 07 — Fasilitas & Operasional (M31–M35, M56–M57)

> **Terakhir diperbarui**: 20 Maret 2026

M31 SDM & Presensi Guru ditenagai **module-vernon-hrm** — lihat `docs/16-vernon-hrm.md` untuk detail integrasi. Dokumen ini menjelaskan layer EDS di atas Vernon HRM.

---

## M31 — SDM & Presensi Guru
**Tier**: BASIC | **Integrasi**: module-vernon-hrm, M04, M09, M39

### Fitur
- [ ] Database guru & staff: data pribadi, NUPTK, sertifikasi, mapel, status kepegawaian
- [ ] Absensi guru: scan QR / fingerprint / face recognition → sync ke Vernon HRM
- [ ] Manajemen cuti & izin: submission online → approval kepala sekolah → Vernon HRM
- [ ] Penjadwalan pengganti otomatis saat guru absen (cari guru tersedia + bidang studi)
- [ ] PKG (Penilaian Kinerja Guru) digital — template resmi Kemendikbud
- [ ] Rekap jam mengajar per bulan untuk tunjangan profesi & sertifikasi
- [ ] Export laporan SDM (skill `xlsx`) — format Dapodik-compatible

### Schema (EDS layer — source of truth di Vernon HRM)

```prisma
model Teacher {
  id            String        @id @default(cuid())
  schoolId      String
  school        School        @relation(fields: [schoolId], references: [id])
  userId        String        @unique   // FK ke User (auth)
  nuptk         String?       @unique
  nip           String?
  name          String
  subjects      String[]
  certStatus    CertStatus    @default(NONE)
  employeeType  EmployeeType
  vernonEmpId   String?       // ID di Vernon HRM setelah sync
  schedules     TeachingSchedule[]
  leaves        LeaveRequest[]
  pkgReviews    PKGReview[]
  createdAt     DateTime      @default(now())
}

model LeaveRequest {
  id          String        @id @default(cuid())
  teacherId   String
  teacher     Teacher       @relation(fields: [teacherId], references: [id])
  schoolId    String
  type        LeaveType
  startDate   DateTime
  endDate     DateTime
  reason      String
  status      LeaveStatus   @default(PENDING)
  approvedBy  String?
  approvedAt  DateTime?
  vernonLeaveId String?     // ID di Vernon HRM setelah sync
  substitute  SubstituteAssignment?
  createdAt   DateTime      @default(now())
}

model SubstituteAssignment {
  id            String       @id @default(cuid())
  leaveId       String       @unique
  leave         LeaveRequest @relation(fields: [leaveId], references: [id])
  originalId    String       // teacherId yang absen
  substituteId  String       // teacherId pengganti
  scheduleId    String       // TeachingSchedule yang digantikan
  date          DateTime
  confirmedAt   DateTime?
}

model PKGReview {
  id            String     @id @default(cuid())
  teacherId     String
  teacher       Teacher    @relation(fields: [teacherId], references: [id])
  schoolId      String
  period        String     // "2025/2026-1" (semester)
  reviewerId    String     // kepala sekolah
  scores        Json       // { kompetensi: score, ... }
  totalScore    Float
  grade         PKGGrade
  notes         String?
  signedAt      DateTime?
  createdAt     DateTime   @default(now())
}

enum CertStatus   { NONE IN_PROCESS CERTIFIED }
enum EmployeeType { PNS PPPK HONORER GTT }
enum LeaveType    { ANNUAL SICK PERMISSION MATERNITY PATERNITY OTHER }
enum LeaveStatus  { PENDING APPROVED REJECTED CANCELLED }
enum PKGGrade     { POOR FAIR GOOD VERY_GOOD EXCELLENT }
```

### Auto-Substitute Logic
```typescript
export async function findSubstitute(
  absentTeacherId: string,
  scheduleId: string,
  date: Date
): Promise<string | null> {
  const schedule = await getSchedule(scheduleId);
  // Cari guru yang: (1) mengajar mapel sama, (2) tidak ada jadwal di slot waktu itu
  const candidates = await prisma.teacher.findMany({
    where: {
      schoolId: schedule.schoolId,
      subjects: { has: schedule.subject },
      id: { not: absentTeacherId },
      schedules: {
        none: {
          dayOfWeek: schedule.dayOfWeek,
          timeSlot: schedule.timeSlot,
          academicYear: schedule.academicYear,
        },
      },
    },
  });
  return candidates[0]?.id ?? null;
}
```

---

## M32 — Smart Gate & Keamanan
**Tier**: PRO (butuh hardware) | **Integrasi**: M01, M02, M14

### Fitur
- [ ] Gate akses via scan QR (kartu M02) atau RFID card
- [ ] Notifikasi WA ke ortu: siswa tiba di sekolah + waktu masuk (via M14)
- [ ] Notifikasi WA ke ortu: siswa pulang (checkout)
- [ ] Alert ke ortu + piket jika siswa belum scan masuk 30 menit setelah bel
- [ ] Dashboard kehadiran realtime untuk guru piket
- [ ] Log kunjungan tamu: check-in, tujuan, lencana tamu digital
- [ ] Integrasi metadata CCTV: timestamp event per kamera (bukan streaming)

### Schema

```prisma
model GateEvent {
  id          String      @id @default(cuid())
  schoolId    String
  gateId      String      // ID gate/pintu masuk
  personId    String      // studentId atau teacherId
  personType  PersonType
  direction   GateDir     // IN | OUT
  method      ScanMethod  // QR | RFID | MANUAL
  scannedAt   DateTime    @default(now())
  deviceId    String      // ID scanner hardware
  notified    Boolean     @default(false)
}

model GateDevice {
  id        String   @id @default(cuid())
  schoolId  String
  name      String   // "Pintu Utama", "Pintu Samping"
  location  String
  type      String   // "QR_SCANNER" | "RFID_READER"
  apiKey    String   @unique  // untuk autentikasi device ke API
  isActive  Boolean  @default(true)
  lastPing  DateTime?
}

model GuestVisit {
  id          String   @id @default(cuid())
  schoolId    String
  name        String
  idNumber    String?
  purpose     String
  hostId      String?  // guru/staff yang dikunjungi
  badgeToken  String   @unique @default(cuid())
  checkedInAt DateTime @default(now())
  checkedOutAt DateTime?
}

enum PersonType { STUDENT TEACHER STAFF }
enum GateDir    { IN OUT }
enum ScanMethod { QR RFID MANUAL }
```

### Device API (hardware scanner memanggil endpoint ini)

```
POST /api/v1/gate/scan
Body: { deviceApiKey, qrData | rfidTag, direction }
Response: { allowed: boolean, personName, photoUrl, message }
```

---

## M33 — Booking Ruang & Fasilitas
**Tier**: BASIC | **Integrasi**: M04, M16, M57

### Fitur
- [ ] Kalender ketersediaan semua ruangan (lab, aula, lapangan, studio)
- [ ] Booking online oleh guru atau panitia kegiatan
- [ ] Approval TU atau kepala sekolah (untuk aula & lapangan)
- [ ] Cek konflik otomatis dengan jadwal pelajaran (M04)
- [ ] Reminder H-1 ke peminjam (via M14)
- [ ] Laporan utilisasi ruangan per bulan

### Schema

```prisma
model Room {
  id          String    @id @default(cuid())
  schoolId    String
  name        String    // "Lab Komputer 1", "Aula"
  type        RoomType
  capacity    Int
  facilities  String[]  // ["proyektor", "AC", "whiteboard"]
  requiresApproval Boolean @default(false)
  bookings    RoomBooking[]
}

model RoomBooking {
  id          String        @id @default(cuid())
  roomId      String
  room        Room          @relation(fields: [roomId], references: [id])
  schoolId    String
  requestedBy String        // userId
  purpose     String
  date        DateTime
  startTime   String        // "08:00"
  endTime     String        // "10:00"
  status      BookingStatus @default(PENDING)
  approvedBy  String?
  approvedAt  DateTime?
  notes       String?
  createdAt   DateTime      @default(now())
}

enum RoomType      { CLASSROOM LAB HALL SPORTS_FIELD STUDIO OTHER }
enum BookingStatus { PENDING APPROVED REJECTED CANCELLED }
```

---

## M34 — Transportasi Sekolah
**Tier**: PRO + LISTING (vendor antar-jemput) | **Integrasi**: M01, M14

### Fitur (armada sekolah internal)
- [ ] GPS tracking bus sekolah realtime (update posisi tiap 10 detik)
- [ ] Peta rute & ETA ke tiap halte
- [ ] Notifikasi ke ortu: bus 10 menit dari halte rumah
- [ ] Absensi siswa saat naik/turun bus (scan QR kartu M02)
- [ ] Log perjalanan harian

### Fitur (listing vendor antar-jemput swasta)
- [ ] Vendor listing tampil di portal ortu (M13) sebagai opsi alternatif
- [ ] Info: area layanan, harga, kapasitas, rating
- [ ] Kontak langsung ke vendor (tidak ada transaksi in-app untuk vendor eksternal)

### Schema

```prisma
model BusRoute {
  id        String    @id @default(cuid())
  schoolId  String
  name      String    // "Rute A - Perumahan Graha"
  stops     BusStop[]
  vehicles  Vehicle[]
  schedules BusSchedule[]
}

model BusStop {
  id        String   @id @default(cuid())
  routeId   String
  route     BusRoute @relation(fields: [routeId], references: [id])
  name      String
  latitude  Float
  longitude Float
  order     Int
}

model Vehicle {
  id           String   @id @default(cuid())
  schoolId     String
  routeId      String?
  route        BusRoute? @relation(fields: [routeId], references: [id])
  plateNumber  String
  capacity     Int
  driverName   String
  driverPhone  String
  gpsDeviceId  String?
  lastLat      Float?
  lastLng      Float?
  lastPing     DateTime?
}

model BusRideLog {
  id          String    @id @default(cuid())
  schoolId    String
  vehicleId   String
  studentId   String
  direction   GateDir   // IN (naik) | OUT (turun)
  stopId      String
  scannedAt   DateTime  @default(now())
}
```

---

## M35 — Helpdesk & Ticketing
**Tier**: BASIC | **Integrasi**: M20, M57, M59 (IoT alert)

### Fitur
- [ ] Form pelaporan kerusakan fasilitas dengan foto (max 3 foto)
- [ ] Kategorisasi: listrik, air, IT/jaringan, furnitur, kebersihan, keamanan
- [ ] Auto-assign ke PIC berdasarkan kategori (configurable per sekolah)
- [ ] Tracking status: OPEN → IN_PROGRESS → RESOLVED → CLOSED
- [ ] SLA per kategori: IT < 4 jam, listrik/air < 8 jam, furnitur < 3 hari
- [ ] Rating kepuasan (1–5 bintang) setelah tiket resolved
- [ ] Alert dari M59 IoT otomatis buat tiket dengan sumber `IOT_SENSOR`
- [ ] Laporan rekap pemeliharaan bulanan → input realisasi ke M51

### Schema

```prisma
model HelpTicket {
  id           String         @id @default(cuid())
  schoolId     String
  ticketNumber String         @unique // "TKT-2026-0042"
  reportedBy   String
  category     TicketCategory
  title        String
  description  String
  photos       String[]       // URL di MinIO
  location     String?        // ruang/lokasi kerusakan
  status       TicketStatus   @default(OPEN)
  priority     TicketPriority @default(MEDIUM)
  assignedTo   String?
  slaDeadline  DateTime?      // dihitung dari createdAt + SLA per kategori
  resolvedAt   DateTime?
  closedAt     DateTime?
  rating       Int?           // 1-5
  ratingNote   String?
  source       TicketSource   @default(MANUAL)
  iotSensorId  String?
  comments     TicketComment[]
  assetId      String?        // FK ke Asset (M20) jika terkait aset
  createdAt    DateTime       @default(now())
  updatedAt    DateTime       @updatedAt
}

model TicketComment {
  id        String     @id @default(cuid())
  ticketId  String
  ticket    HelpTicket @relation(fields: [ticketId], references: [id])
  authorId  String
  message   String
  photos    String[]
  createdAt DateTime   @default(now())
}

enum TicketCategory { ELECTRICAL PLUMBING IT FURNITURE CLEANING SECURITY OTHER }
enum TicketStatus   { OPEN IN_PROGRESS RESOLVED CLOSED }
enum TicketPriority { LOW MEDIUM HIGH CRITICAL }
enum TicketSource   { MANUAL IOT_SENSOR }
```

---

## M56 — Manajemen Seragam & Atribut Sekolah
**Tier**: LISTING | **Integrasi**: M08, M03

### Fitur
- [ ] Katalog seragam resmi: foto, ukuran (S–XXL + anak), harga, keterangan
- [ ] Pemesanan seragam saat PPDB (M08) — terintegrasi di form daftar ulang
- [ ] Wizard ukuran: input tinggi & berat → rekomendasi ukuran otomatis
- [ ] Manajemen stok (di koperasi M03 atau vendor listing)
- [ ] Notifikasi ortu saat seragam siap diambil (via M14)
- [ ] Rekap pesanan per angkatan (export ke skill `xlsx`) untuk vendor

### Schema

```prisma
model UniformCatalog {
  id          String         @id @default(cuid())
  schoolId    String
  name        String         // "Seragam Putih Abu-Abu Putra"
  type        UniformType
  sizes       UniformSize[]
  vendorId    String?        // listing vendor jika bukan koperasi
  isActive    Boolean        @default(true)
  photoUrl    String?
  createdAt   DateTime       @default(now())
}

model UniformSize {
  id          String         @id @default(cuid())
  catalogId   String
  catalog     UniformCatalog @relation(fields: [catalogId], references: [id])
  size        String         // "S", "M", "L", "XL", "6A", "8A"
  price       Decimal
  stock       Int            @default(0)
  orders      UniformOrder[]
}

model UniformOrder {
  id          String       @id @default(cuid())
  sizeId      String
  size        UniformSize  @relation(fields: [sizeId], references: [id])
  schoolId    String
  studentId   String
  quantity    Int          @default(1)
  status      OrderStatus  @default(PENDING)
  paidAt      DateTime?
  readyAt     DateTime?
  pickedUpAt  DateTime?
  ppdbId      String?      // FK ke PPDB enrollment
  createdAt   DateTime     @default(now())
}

enum UniformType { SHIRT PANTS SKIRT BATIK SPORT_TOP SPORT_BOTTOM HAT TIE OTHER }
enum OrderStatus { PENDING PAID READY PICKED_UP CANCELLED }
```

---

## M57 — Event & Kepanitiaan
**Tier**: BASIC | **Integrasi**: M07, M14, M18, M33, M51, M54

### Fitur
- [ ] Perencanaan event: nama, tanggal, jenis, PIC, anggaran estimasi
- [ ] Struktur kepanitiaan digital: divisi, anggota, tugas, deadline, progress
- [ ] Checklist persiapan event dengan assignment ke panitia
- [ ] Booking ruang & fasilitas otomatis via M33
- [ ] Anggaran event dengan approval → realisasi terintegrasi ke M51
- [ ] Ticketing event berbayar (pentas seni, seminar) via M18 payment
- [ ] Pendaftaran peserta internal & eksternal dengan QR check-in
- [ ] Dokumentasi & laporan pasca-event (foto, ringkasan, evaluasi)
- [ ] Galeri event otomatis publish ke M07 (website sekolah)
- [ ] Mobilisasi volunteer via M54

### Schema

```prisma
model SchoolEvent {
  id            String          @id @default(cuid())
  schoolId      String
  name          String
  type          EventType
  description   String?
  startDate     DateTime
  endDate       DateTime
  location      String?
  estimatedBudget Decimal?
  actualBudget  Decimal?
  budgetApproved Boolean        @default(false)
  status        EventStatus     @default(PLANNING)
  picId         String          // penanggung jawab utama
  committees    EventCommittee[]
  tasks         EventTask[]
  tickets       EventTicket[]
  registrations EventRegistration[]
  documents     EventDocument[]
  createdAt     DateTime        @default(now())
}

model EventCommittee {
  id        String      @id @default(cuid())
  eventId   String
  event     SchoolEvent @relation(fields: [eventId], references: [id])
  userId    String
  division  String      // "Acara", "Konsumsi", "Dekorasi", "Dokumentasi"
  role      String      // "Ketua", "Anggota"
}

model EventTask {
  id          String      @id @default(cuid())
  eventId     String
  event       SchoolEvent @relation(fields: [eventId], references: [id])
  title       String
  assignedTo  String
  deadline    DateTime
  status      TaskStatus  @default(TODO)
  notes       String?
}

model EventTicket {
  id          String             @id @default(cuid())
  eventId     String
  event       SchoolEvent        @relation(fields: [eventId], references: [id])
  name        String             // "Tiket Umum", "Tiket VIP"
  price       Decimal
  quota       Int
  sold        Int                @default(0)
  registrations EventRegistration[]
}

model EventRegistration {
  id          String      @id @default(cuid())
  eventId     String
  event       SchoolEvent @relation(fields: [eventId], references: [id])
  ticketId    String?
  ticket      EventTicket? @relation(fields: [ticketId], references: [id])
  name        String
  email       String?
  phone       String?
  qrToken     String      @unique @default(cuid())
  checkedIn   Boolean     @default(false)
  checkedInAt DateTime?
  paymentId   String?     // FK ke M18 Payment
  createdAt   DateTime    @default(now())
}

enum EventType   { CEREMONY COMPETITION CULTURAL SPORTS SEMINAR GRADUATION OTHER }
enum EventStatus { PLANNING APPROVED ONGOING COMPLETED CANCELLED }
enum TaskStatus  { TODO IN_PROGRESS DONE OVERDUE }
```
