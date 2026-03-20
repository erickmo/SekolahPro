# 16 — Vernon HRM Integration

> **Terakhir diperbarui**: 20 Maret 2026
> **Status**: Placeholder — modul akan diisi saat `module-vernon-hrm` tersedia di `/mnt/skills/`

`module-vernon-hrm` digunakan sebagai engine utama untuk **M31 — SDM & Presensi Guru**. EDS tidak membangun logika HRM dari nol; sebaliknya, M31 adalah adapter layer yang menghubungkan data EDS ke Vernon HRM.

---

## Arsitektur Integrasi

```
EDS (M31 SDM & Presensi Guru)
    │
    │  REST API / event bridge
    ▼
module-vernon-hrm
    │
    ├── Employee Management    ← mapped ke Teacher/Staff di EDS
    ├── Attendance Engine      ← absensi guru EDS
    ├── Leave Management       ← cuti, izin, sakit
    ├── Payroll Engine         ← tunjangan & sertifikasi
    └── Performance (PKG)      ← Penilaian Kinerja Guru
```

---

## Mapping Entitas

| EDS Entitas | Vernon HRM Entitas | Catatan |
|-------------|-------------------|---------|
| `Teacher` | `Employee` | NIP → employeeId, schoolId → tenantId |
| `StaffAttendance` | `AttendanceRecord` | QR/fingerprint → sumber absensi |
| `LeaveRequest` | `LeaveApplication` | Tipe: cuti tahunan, sakit, izin |
| `TeachingHours` | `WorkLog` | Rekap jam mengajar per bulan |
| `PKGScore` | `PerformanceReview` | Siklus: semester |
| `Substitute` | `ShiftSwap` | Pengganti mengajar |

---

## Konfigurasi Tenant Vernon HRM

Setiap sekolah di EDS dipetakan sebagai tenant terpisah di Vernon HRM:

```typescript
// packages/hrm/client.ts
export async function initVernon HRMTenant(schoolId: string, schoolName: string) {
  return await vernonHRM.tenants.create({
    tenantId: `eds_${schoolId}`,
    name: schoolName,
    timezone: 'Asia/Jakarta',
    workingHours: { start: '07:00', end: '16:00' },
    workDays: ['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT'],
    leavePolicy: INDONESIAN_SCHOOL_LEAVE_POLICY,
  });
}
```

---

## Fitur M31 yang Ditenagai Vernon HRM

### Absensi Guru
- [ ] Sync rekaman absensi (QR/fingerprint) dari EDS → Vernon HRM attendance log
- [ ] Laporan kehadiran otomatis dari Vernon HRM engine
- [ ] Rekap bulanan untuk tunjangan → Vernon HRM payroll input

### Manajemen Cuti & Izin
- [ ] Submission cuti via portal EDS → Vernon HRM leave workflow
- [ ] Approval chain: Guru → Kepala Sekolah (sesuai RBAC EDS)
- [ ] Saldo cuti real-time dari Vernon HRM
- [ ] Integrasi ke penjadwalan pengganti otomatis

### Penilaian Kinerja Guru (PKG)
- [ ] Template PKG Kemendikbud di Vernon HRM performance module
- [ ] Input observasi per semester oleh kepala sekolah
- [ ] Generate laporan PKG PDF (skill `pdf`)
- [ ] Rekap PKG → input ke syarat kenaikan pangkat/sertifikasi

### Payroll Guru (opsional, tier ENTERPRISE)
- [ ] Komponen gaji: gaji pokok, tunjangan profesi, TPP daerah
- [ ] Rekap jam mengajar sebagai dasar perhitungan
- [ ] Slip gaji digital per guru
- [ ] Export untuk transfer bank massal

---

## Event Bridge EDS ↔ Vernon HRM

```typescript
// Setiap aksi di M31 EDS memicu event ke Vernon HRM
type HRMEvent =
  | { type: 'attendance.checkin';   employeeId: string; timestamp: Date; method: 'QR' | 'FINGERPRINT' }
  | { type: 'attendance.checkout';  employeeId: string; timestamp: Date }
  | { type: 'leave.requested';      employeeId: string; type: string; dates: Date[] }
  | { type: 'leave.approved';       employeeId: string; approvedBy: string }
  | { type: 'substitute.assigned';  originalId: string; substituteId: string; scheduleId: string }
  | { type: 'performance.reviewed'; employeeId: string; period: string; score: number }
```

---

## Instalasi & Konfigurasi

> Baca `module-vernon-hrm/SKILL.md` terlebih dahulu sebelum implementasi.
> Path skill: `/mnt/skills/user/module-vernon-hrm/SKILL.md` (install terlebih dahulu)

```bash
# Install Vernon HRM sebagai service terpisah
docker-compose -f infrastructure/docker-compose.yml up vernon-hrm -d

# Inisialisasi tenant untuk semua sekolah aktif
pnpm --filter @eds/hrm run init-tenants

# Sync data guru yang sudah ada ke Vernon HRM
pnpm --filter @eds/hrm run migrate-existing-staff
```

---

## Environment Variables

```env
VERNON_HRM_API_URL="http://vernon-hrm:5100"   # internal docker network
VERNON_HRM_API_KEY="..."
VERNON_HRM_TENANT_PREFIX="eds_"
VERNON_HRM_WEBHOOK_SECRET="..."               # untuk validasi event dari Vernon
```

---

## Catatan Penting

1. **Vernon HRM adalah source of truth untuk data kepegawaian** — data di tabel `Teacher` EDS hanya cache/read-only setelah integrasi aktif.
2. **Multi-tenant isolation tetap berlaku** — Vernon HRM harus dikonfigurasi dengan tenant per `schoolId`, bukan satu tenant global.
3. **Sinkronisasi dua arah**: perubahan data guru di EDS (M01 admin) → push ke Vernon HRM; perubahan di Vernon HRM (misal: cuti diapprove) → webhook ke EDS untuk update status.
4. **Offline graceful**: jika Vernon HRM down, absensi tetap dicatat lokal di EDS dan di-sync saat koneksi pulih (eventual consistency).
