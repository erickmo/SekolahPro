# 03 — Komunikasi & Komunitas (M13–M17)

---

## M13 — Portal Orang Tua
**Tier**: BASIC | **Integrasi**: M01, M03, M04, M14, M18, M22

Akses real-time bagi orang tua untuk memantau perkembangan anak.

- [ ] Pantau nilai & perkembangan per KD secara real-time
- [ ] Rekap absensi harian + notifikasi otomatis jika alpha
- [ ] Saldo & riwayat transaksi tabungan koperasi anak
- [ ] Chat langsung dengan wali kelas & guru mapel
- [ ] Download rapor digital (PDF)
- [ ] Agenda & pengumuman sekolah
- [ ] Persetujuan digital (izin kegiatan, surat, dll)
- [ ] Riwayat pembayaran SPP & tagihan lainnya
- [ ] Akses ringkasan keuangan dana komite (dari M52)
- [ ] Notifikasi push dari semua modul relevan via M14

---

## M14 — Messaging & Notifikasi
**Tier**: PRO (1.000 WA/bulan, bisa tambah add-on) | **Integrasi**: semua modul

Gateway notifikasi terpusat. Semua notifikasi dari seluruh modul EDS dikirim melalui modul ini.

**Channel Priority**: WhatsApp Business API → Push Notification → Email → SMS

### Jenis Notifikasi

| Event | Trigger | Penerima |
|-------|---------|----------|
| Siswa tidak hadir | M01 absensi | Orang tua |
| Tagihan SPP jatuh tempo | M18 | Orang tua |
| Hasil ujian tersedia | M05 | Siswa + Ortu |
| Siswa tiba di sekolah | M32 Smart Gate | Orang tua |
| Transaksi tabungan | M03 | Orang tua |
| Alert EWS | M23 | Guru BK + Wali kelas |
| Pengumuman sekolah | Admin | Semua (blast) |
| Reminder remedial | M50 | Siswa + Ortu |

### Template Engine

```typescript
// Template WA dengan variabel dinamis
const templates = {
  ATTENDANCE_ABSENT: `Yth. {{guardianName}},
Putra/putri Anda *{{studentName}}* tidak hadir di {{schoolName}} pada {{date}}.
Silakan konfirmasi melalui portal: {{portalUrl}}`,

  SPP_REMINDER: `Yth. {{guardianName}},
Tagihan SPP {{studentName}} bulan {{month}} sebesar *Rp {{amount}}*
jatuh tempo {{dueDate}}. Bayar sekarang: {{paymentUrl}}`,
};
```

### Quota & Rate Limiting

```typescript
interface WAQuotaConfig {
  monthlyLimit: number;        // default 1000 di tier PRO
  used: number;
  remaining: number;
  resetAt: Date;
  // prioritas saat quota hampir habis:
  priority: ('ATTENDANCE' | 'PAYMENT' | 'EMERGENCY' | 'ANNOUNCEMENT')[];
}
```

---

## M15 — Forum & Komunitas Sekolah
**Tier**: BASIC | **Integrasi**: M04, M26b

- [ ] Forum diskusi per mata pelajaran & kelas
- [ ] Grup ekskul & angkatan
- [ ] Papan pengumuman digital (pengganti mading fisik)
- [ ] Polling & survei internal
- [ ] Q&A siswa dengan guru (terstruktur, dimoderasi)
- [ ] Moderasi konten otomatis via Claude AI (M26b)

---

## M16 — Rapat & Video Conference
**Tier**: PRO | **Integrasi**: M53, M57

- [ ] Penjadwalan rapat komite sekolah (M53)
- [ ] Slot konsultasi guru–wali murid (booking online via M13)
- [ ] Kelas online / remedial via video (Jitsi Meet self-hosted)
- [ ] Rekaman sesi (opsional, dengan explicit consent)
- [ ] Transkripsi & notulen otomatis via Claude AI
- [ ] Pengingat otomatis H-1 dan H-0 ke peserta via M14

---

## M17 — Blog & Majalah Sekolah Digital
**Tier**: BASIC | **Integrasi**: M07, M15

- [ ] Platform penerbitan karya tulis siswa
- [ ] Kategori: berita, opini, cerpen, puisi, karya ilmiah
- [ ] Workflow editorial: draft → review guru → publish
- [ ] Editor rich-text dengan gambar & video embed
- [ ] Versi digital majalah sekolah (PDF flipbook — skill `pdf`)
- [ ] Artikel terbaik otomatis tampil di M07 (website sekolah)
