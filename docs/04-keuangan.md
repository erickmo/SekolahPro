# 04 — Keuangan & Bisnis (M18–M22, M51–M52)

---

## M18 — Pembayaran SPP Online
**Tier**: PRO (+ fee transaksi 0.5%) | **Integrasi**: M03, M09, M13, M51

**Entitas**: `Invoice`, `Payment`, `PaymentMethod`, `Installment`

- [ ] Generate tagihan SPP otomatis tiap awal bulan
- [ ] Multi-channel: QRIS, VA BCA/BNI/Mandiri, GoPay, OVO, ShopeePay
- [ ] Auto-reminder WA (H-7, H-3, H-0 jatuh tempo) via M14
- [ ] Cicilan dengan persetujuan kepala sekolah
- [ ] Rekap penerimaan harian/bulanan untuk bendahara
- [ ] Laporan tunggakan per siswa, per kelas
- [ ] Kwitansi digital otomatis (PDF — skill `pdf`)
- [ ] Integrasi ke M51 (RKAS) sebagai realisasi pendapatan
- [ ] Reconciliation otomatis dengan payment gateway

**Fee Model**:
```
EDS mengambil 0.5% dari setiap transaksi SPP (di atas biaya payment gateway)
Minimal fee: Rp 500/transaksi
Tercatat di Listing billing, bukan subscription sekolah
```

---

## M19 — Kantin Digital & Cashless
**Tier**: PRO + LISTING (katering vendor) | **Integrasi**: M03, M14, M30

- [ ] Menu kantin digital dengan foto & harga
- [ ] Pre-order makan siang (cut-off jam 08.00)
- [ ] Pembayaran via EDS Wallet (top-up by ortu)
- [ ] Notifikasi ke ortu saat anak membeli (item & nominal)
- [ ] Batasan pengeluaran harian per siswa (set ortu)
- [ ] Laporan gizi mingguan berdasarkan pembelian → M30
- [ ] Dashboard penjualan untuk pengelola kantin
- [ ] Dukungan multiple vendor (kantin + katering partner dari Listing)

**EDS Wallet**:
```typescript
interface WalletTopUp {
  studentId: string;
  amount: number;          // top-up oleh ortu
  channel: PaymentChannel; // transfer, QRIS
  purpose: 'CANTEEN' | 'COOPERATIVE' | 'GENERAL';
  dailySpendLimit?: number; // batas pengeluaran harian
}
```

---

## M20 — Manajemen Aset Sekolah
**Tier**: BASIC | **Integrasi**: M35, M51

- [ ] Database aset: ruangan, peralatan, furnitur, kendaraan, elektronik
- [ ] Barcode/QR label per item aset (print via skill `pdf`)
- [ ] Scan QR untuk cek status & lokasi aset
- [ ] Jadwal pemeliharaan & service berkala
- [ ] Laporan kondisi & depresiasi aset per tahun
- [ ] Alur pengadaan: request → approve → terima → input ke sistem
- [ ] Alert ke M35 (Helpdesk) saat aset dilaporkan rusak
- [ ] Export laporan aset untuk audit (skill `xlsx`)

---

## M21 — Perpustakaan Digital
**Tier**: BASIC | **Integrasi**: M04, M24

- [ ] Katalog OPAC (Online Public Access Catalog)
- [ ] Peminjaman & pengembalian via scan QR kode buku
- [ ] Koleksi e-book & jurnal digital (upload PDF)
- [ ] Reminder pengembalian otomatis via M14 (WA/push)
- [ ] Denda keterlambatan → tagihan di M18
- [ ] Rekomendasi buku berdasarkan riwayat baca & kurikulum (M04)
- [ ] Laporan statistik: buku terpopuler, utilisasi, koleksi
- [ ] Integrasi dengan learning path (M24)

---

## M22 — Beasiswa & Bantuan Siswa
**Tier**: FREE | **Integrasi**: M01, M09, M18

- [ ] Database program beasiswa (internal & eksternal, termasuk BSM/PIP)
- [ ] Formulir pendaftaran beasiswa online
- [ ] Kriteria seleksi otomatis (nilai, kondisi ekonomi, dll)
- [ ] Monitoring penerima BSM/PIP — sinkronisasi dengan Dapodik (M40)
- [ ] Pencairan & pelaporan dana bantuan
- [ ] Integrasi ke data ekonomi siswa di M01

---

## M51 — Anggaran & RKAS
**Tier**: BASIC | **Integrasi**: M09, M18, M40, M52, M53

Digitalisasi penyusunan dan monitoring RKAS (Rencana Kegiatan & Anggaran Sekolah).

**Entitas**: `BudgetPlan`, `BudgetItem`, `Realization`, `BudgetApproval`

- [ ] Penyusunan RKAS tahunan secara kolaboratif (per unit/bidang)
- [ ] Template standar sesuai format Kemendikbud
- [ ] Workflow approval: TU → Bendahara → Kepala Sekolah → Komite (M53)
- [ ] Monitoring realisasi vs. rencana real-time (per bulan, per item)
- [ ] Alert otomatis jika realisasi mendekati 90% batas anggaran
- [ ] Revisi RKAS mid-year dengan audit trail lengkap
- [ ] Sinkronisasi ke ARKAS (M40)
- [ ] Laporan pertanggungjawaban untuk rapat pleno
- [ ] Dashboard keuangan di M12

```prisma
model BudgetPlan {
  id          String         @id @default(cuid())
  schoolId    String
  fiscalYear  String         // "2025"
  status      BudgetStatus   @default(DRAFT)
  totalBOS    Decimal
  totalComite Decimal
  totalOther  Decimal
  items       BudgetItem[]
  approvals   BudgetApproval[]
  createdAt   DateTime       @default(now())
}

model BudgetItem {
  id          String      @id @default(cuid())
  planId      String
  plan        BudgetPlan  @relation(fields: [planId], references: [id])
  category    String      // "Sarana Prasarana", "Pengembangan Guru", dll
  description String
  plannedAmount  Decimal
  realizedAmount Decimal  @default(0)
  source      BudgetSource // BOS | KOMITE | OTHER
  quarter     Int?         // 1-4
}

enum BudgetStatus { DRAFT PENDING_REVIEW APPROVED ACTIVE CLOSED }
enum BudgetSource { BOS KOMITE BEASISWA OTHER }
```

---

## M52 — Dana Komite & Sumbangan Masyarakat
**Tier**: BASIC | **Integrasi**: M51, M53, M18

- [ ] Pencatatan donasi & sumbangan (tunai & transfer)
- [ ] Portal transparansi: ortu bisa lihat penggunaan dana komite (read-only)
- [ ] Mini crowdfunding untuk kegiatan sekolah (study tour, porseni, dll)
- [ ] Target & progress funding per program (tampil di M07 website)
- [ ] Laporan tahunan dana komite untuk rapat pleno
- [ ] Kwitansi digital otomatis
- [ ] Integrasi ke M51 (RKAS) sebagai sumber pendapatan non-BOS

**Catatan Transparansi**: Sekolah yang mengaktifkan M52 secara otomatis menampilkan ringkasan penggunaan dana di portal ortu (M13). Ini adalah fitur yang tidak bisa dimatikan demi akuntabilitas.
