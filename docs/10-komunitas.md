# 10 — Komunitas & Tata Kelola (M53–M54)

> **Terakhir diperbarui**: 20 Maret 2026

---

## M53 — Komite Sekolah Digital
**Tier**: BASIC | **Integrasi**: M16, M51, M52

Digitalisasi peran komite sekolah sebagai mitra resmi yang punya hak suara dalam kebijakan sekolah.

### Fitur
- [ ] Profil & struktur kepengurusan komite (ketua, sekretaris, bendahara, anggota)
- [ ] Rapat digital dengan agenda, notulen, voting (via M16 Video Conference)
- [ ] Akses laporan keuangan M51 & M52 (read-only, sesuai wewenang komite)
- [ ] Komunikasi formal: komite ↔ kepala sekolah ↔ dinas pendidikan
- [ ] Arsip digital keputusan rapat (immutable setelah ditandatangani)
- [ ] Persetujuan digital untuk kebijakan yang butuh tanda tangan komite (e-signature)
- [ ] Notifikasi rapat & agenda (via M14)

### Schema

```prisma
model CommitteeBoard {
  id          String            @id @default(cuid())
  schoolId    String            @unique
  periodStart DateTime
  periodEnd   DateTime
  members     CommitteeMember[]
  meetings    CommitteeMeeting[]
  decisions   CommitteeDecision[]
  isActive    Boolean           @default(true)
  createdAt   DateTime          @default(now())
}

model CommitteeMember {
  id          String         @id @default(cuid())
  boardId     String
  board       CommitteeBoard @relation(fields: [boardId], references: [id])
  userId      String         // akun EDS user dengan role KOMITE_SEKOLAH
  name        String
  position    CommitteePos
  phone       String?
  email       String
  isActive    Boolean        @default(true)
}

model CommitteeMeeting {
  id          String            @id @default(cuid())
  boardId     String
  board       CommitteeBoard    @relation(fields: [boardId], references: [id])
  schoolId    String
  title       String
  agenda      String[]
  scheduledAt DateTime
  location    String?
  videoCallUrl String?         // dari M16
  minutesUrl  String?          // URL notulen PDF setelah rapat
  attendees   String[]         // userId peserta
  decisions   CommitteeDecision[]
  status      MeetingStatus    @default(SCHEDULED)
  createdAt   DateTime         @default(now())
}

model CommitteeDecision {
  id          String           @id @default(cuid())
  meetingId   String
  meeting     CommitteeMeeting @relation(fields: [meetingId], references: [id])
  boardId     String
  board       CommitteeBoard   @relation(fields: [boardId], references: [id])
  schoolId    String
  title       String
  description String
  votesFor    Int              @default(0)
  votesAgainst Int             @default(0)
  abstentions Int              @default(0)
  result      VoteResult
  signedByIds String[]         // userId komite yang menandatangani
  signedAt    DateTime?
  isBinding   Boolean          @default(true)
  createdAt   DateTime         @default(now())
}

enum CommitteePos { CHAIR SECRETARY TREASURER MEMBER }
enum MeetingStatus { SCHEDULED ONGOING COMPLETED CANCELLED }
enum VoteResult    { APPROVED REJECTED DEFERRED }
```

### Akses Laporan Keuangan
```typescript
// Komite dapat akses read-only ke summary keuangan, bukan detail transaksi
export async function getCommitteeFinanceSummary(schoolId: string) {
  const [rkas, committeeFund] = await Promise.all([
    getRKASSummary(schoolId),   // M51 - hanya total per kategori
    getCommitteeFundSummary(schoolId), // M52 - total penerimaan & pengeluaran
  ]);
  return { rkas, committeeFund };
  // Tidak termasuk: detail SPP per siswa, detail tabungan koperasi
}
```

---

## M54 — Volunteer & Partisipasi Komunitas
**Tier**: FREE | **Integrasi**: M57, M07

### Fitur
- [ ] Pendaftaran volunteer: ortu, alumni, atau profesional eksternal
- [ ] Profil volunteer: keahlian, ketersediaan waktu, area kontribusi
- [ ] Matching keahlian dengan kebutuhan sekolah
- [ ] Jadwal & koordinasi kegiatan volunteer
- [ ] Rekam jam & jenis kontribusi
- [ ] Sertifikat kontribusi digital (skill `pdf`)
- [ ] Program "orang tua mengajar": jadwal kelas tamu, evaluasi
- [ ] Integrasi ke M57 (Event) untuk mobilisasi volunteer kegiatan besar

### Schema

```prisma
model VolunteerProfile {
  id            String            @id @default(cuid())
  schoolId      String
  userId        String?           // jika sudah punya akun EDS (ortu/alumni)
  name          String
  email         String
  phone         String?
  relation      VolunteerRelation // PARENT | ALUMNI | EXTERNAL
  skills        String[]          // keahlian yang ditawarkan
  availability  Json              // { days: string[], timeRange: string }
  totalHours    Float             @default(0)
  activities    VolunteerActivity[]
  certificates  VolunteerCertificate[]
  isActive      Boolean           @default(true)
  createdAt     DateTime          @default(now())
}

model VolunteerActivity {
  id            String           @id @default(cuid())
  profileId     String
  profile       VolunteerProfile @relation(fields: [profileId], references: [id])
  schoolId      String
  title         String           // "Kelas Tamu Coding", "Juri Lomba"
  description   String?
  eventId       String?          // FK ke SchoolEvent (M57) jika terkait event
  date          DateTime
  hours         Float
  coordinatedBy String           // userId TU atau admin yang koordinasi
  feedback      String?
  createdAt     DateTime         @default(now())
}

model VolunteerCertificate {
  id            String           @id @default(cuid())
  profileId     String
  profile       VolunteerProfile @relation(fields: [profileId], references: [id])
  schoolId      String
  title         String
  totalHours    Float
  issuedAt      DateTime         @default(now())
  fileUrl       String           // PDF di MinIO
  academicYear  String
}

enum VolunteerRelation { PARENT ALUMNI EXTERNAL_PROFESSIONAL }
```
