# 10 - Tata Kelola Koperasi (Governance)

Dokumen ini mendeskripsikan kerangka tata kelola koperasi dalam sistem SekolahPro yang mencakup: struktur organisasi dan internal controls (ADR-K025), manajemen RAT/RALB (ADR-K026), indikator kesehatan koperasi (ADR-K031), siklus keanggotaan lengkap (ADR-K030), program pendidikan anggota (ADR-K037), prosedur pembubaran (ADR-K038), dan konsolidasi multi-cabang (ADR-K039). Keseluruhan governance framework ini dibangun di atas UU Koperasi No. 25/1992 dan peraturan OJK.

---

## ADR References

| ADR | Judul | Status |
|-----|-------|--------|
| ADR-K025 | Cooperative Governance & Internal Controls | Accepted |
| ADR-K026 | RAT (Rapat Anggota Tahunan) Management | Accepted |
| ADR-K030 | Membership Lifecycle Management | Accepted |
| ADR-K031 | Cooperative Health Indicators & Risk Management | Accepted |
| ADR-K037 | Member Education Program | Accepted |
| ADR-K038 | Cooperative Dissolution Process | Accepted |
| ADR-K039 | Multi-Branch Consolidation Reporting | Accepted |

---

## Domain Entities

### Governance & Internal Controls (K025)

| Entity | Deskripsi |
|--------|-----------|
| `governance_position` | Jabatan struktural koperasi — Ketua, Sekretaris, Bendahara, Pengawas, Manajer, DPS. Menyimpan authority limits dan status jabatan. |
| `internal_audit` | Rekaman audit internal oleh Pengawas — temuan, severitas, rekomendasi, dan status presentasi ke RAT. |
| `conflict_of_interest` | Pengungkapan konflik kepentingan oleh Pengurus/Pengawas. Wajib diisi sebelum memutuskan hal yang terkait. |

### RAT Management (K026)

| Entity | Deskripsi |
|--------|-----------|
| `rat_meeting` | Data rapat anggota (RAT/RALB/Gabungan/Pembubaran) — jadwal, kuorum, status, dan referensi dokumen. |
| `rat_agenda_item` | Item agenda rapat — proposal, hasil voting, tindak lanjut, dan penanggung jawab. |
| `rat_attendance` | Daftar kehadiran anggota di rapat — hadir langsung, kuasa, atau tidak hadir. |
| `rat_vote` | Rekaman suara per anggota per item agenda. UNIQUE per agenda_item + voter. |
| `rat_election` | Pemilihan Pengurus/Pengawas dalam RAT — posisi, masa jabatan, metode pemilihan. |
| `rat_election_candidate` | Kandidat dalam pemilihan beserta jumlah suara dan hasil akhir. |

### Membership Lifecycle (K030)

| Entity | Deskripsi |
|--------|-----------|
| `membership_lifecycle_event` | Rekaman setiap perubahan status keanggotaan — persetujuan, pengunduran diri, ekskumul, kematian, dormant, transfer. Menyimpan financial impact. |

### Health Indicators (K031)

| Entity | Deskripsi |
|--------|-----------|
| `kpi_definition` | Definisi KPI kesehatan koperasi — formula, unit, arah (higher/lower is better), threshold sehat/peringatan/kritis. |
| `kpi_measurement` | Pengukuran KPI aktual per periode — nilai, status, tren. |
| `early_warning_alert` | Alert yang dipicu saat KPI melewati threshold — tindakan korektif dan eskalasi. |
| `stress_test_scenario` | Skenario stress test — parameter, hasil proyeksi, dan apakah koperasi survive. |

### Education (K037)

| Entity | Deskripsi |
|--------|-----------|
| `education_course` | Kursus/pelatihan — tipe, konten, prasyarat, kuis, dan target audiens. |
| `education_enrollment` | Pendaftaran dan progress anggota dalam kursus — status, skor kuis, sertifikat. |

### Dissolution (K038)

| Entity | Deskripsi |
|--------|-----------|
| `dissolution_process` | Proses pembubaran koperasi — tipe, tim likuidasi, tahapan, summary keuangan. |

### Multi-Branch (K039)

| Entity | Deskripsi |
|--------|-----------|
| `branch_financial_summary` | Ringkasan keuangan per cabang dan konsolidasi — P&L, neraca, KPI. |

---

## Business Rules

### Struktur Organisasi

1. **Satu Posisi Aktif.** Setiap tipe jabatan struktural (misal: Ketua) hanya boleh diisi satu orang aktif dalam satu waktu. Enforced via UNIQUE constraint `(tenant_id, position_type, status) WHERE status = 'active'`.

2. **DPS Wajib untuk BMT.** Untuk `coop_type = "islamic"`, Dewan Pengawas Syariah (DPS) minimal 1 anggota adalah WAJIB. Approval pembiayaan di atas threshold, produk baru, dan laporan keuangan harus mendapat persetujuan DPS.

3. **Masa Jabatan Tercatat.** Setiap jabatan memiliki `term_start`, `term_end`, dan `term_number`. Tidak ada jabatan tanpa batas waktu kecuali dikonfigurasi eksplisit.

### Authority Matrix & Segregation of Duties

4. **Maker-Checker.** Setiap transaksi di atas threshold harus memiliki maker dan checker yang berbeda. Tidak boleh menyetujui transaksi sendiri (no self-approval).

5. **Four-Eyes Principle.** Transaksi bernilai besar (penarikan > Rp 100 juta, disbursement pinjaman > Rp 50 juta, write-off) memerlukan minimal 2 approval dari jabatan berbeda.

6. **Rotasi Teller.** Teller tidak boleh bertugas di posisi yang sama lebih dari 6 bulan berturut-turut.

7. **Write-Off Authority.** Penghapusan pinjaman (write-off) memerlukan proposal Manager → review Bendahara → approval Ketua. Tidak bisa dilakukan oleh satu pihak saja.

### Internal Audit

8. **Jadwal Audit Minimum.** (a) Audit rutin: minimal 1x per semester oleh Pengawas. (b) Cash opname: minimal 1x per bulan (random). (c) Compliance audit: 1x per tahun. (d) Audit khusus: atas permintaan RAT atau jika ada indikasi irregularitas.

9. **Selisih Kas.** Selisih kas > Rp 50.000 wajib dilaporkan ke supervisor. Selisih > Rp 500.000 wajib dilaporkan ke manajer. Log kas tidak dapat dihapus (immutable).

10. **Conflict of Interest.** Pengurus/Pengawas WAJIB mengungkapkan konflik kepentingan sebelum memutuskan hal apapun yang melibatkan pihak terkait. Threshold dan ketentuan:
    - Penerimaan hadiah/gratifikasi > Rp 500.000 dari pihak manapun yang berhubungan dengan bisnis koperasi wajib dilaporkan dalam 5 hari kerja. Field: `conflict_of_interest.disclosure_amount`, `conflict_of_interest.gift_value`.
    - Hadiah yang tidak dilaporkan atau diterima tanpa pelaporan = pelanggaran kode etik → sanksi sesuai AD/ART.

### RAT Management

11. **Quorum RAT.** RAT pertama: > 50% anggota berhak hadir. RAT kedua (jika quorum pertama gagal): > 33%. Perubahan AD/ART dan pembubaran: > 66%.

12. **Voting Passing Criteria.** Keputusan biasa: > 50% suara hadir. Perubahan AD/ART dan pembubaran: > 66%. Pemilihan: plurality (suara terbanyak menang).

13. **Empat Jenis Rapat.** (a) ANNUAL — RAT tahunan, wajib. (b) SPECIAL (RALB) — atas permintaan ≥ 25% anggota atau keputusan Pengawas. (c) JOINT — untuk penggabungan koperasi. (d) DISSOLUTION — untuk pembubaran.

14. **Tindak Lanjut Otomatis dari RAT.** Setelah rapat selesai, sistem secara otomatis: (a) Membuat `governance_position` baru untuk pengurus/pengawas yang terpilih. (b) Memicu alur distribusi SHU (K016) jika SHU disetujui.

### Membership Lifecycle

15. **Anggota Tidak Aktif → Dormant.** Trigger dormant: tidak ada transaksi selama 12 bulan DAN tidak hadir di 2 RAT berturut-turut DAN tidak bayar simpanan wajib selama 6 bulan. Peringatan H-30 dan H-7 sebelum status berubah.

16. **Hak Anggota Dormant.** Tabungan tetap aman, tidak hangus. Namun anggota dormant tidak berhak voting di RAT dan tidak bisa mengajukan pinjaman baru.

17. **Perhitungan Refund Keluar.** Net Refund = Simpanan Pokok + Simpanan Wajib + Saldo Tabungan + SHU Pro-rata − Pinjaman Outstanding − Denda − Biaya Administrasi.

18. **Ekskumul (Pemecatan).** Proses ekskumul wajib melalui due process: notifikasi tertulis → hearing anggota → investigasi Pengawas → keputusan Ketua + 1 Pengawas. Anggota berhak banding ke RAT. Simpanan pokok tetap dikembalikan sesuai UU Koperasi.

19. **Ahli Waris Kematian.** Jika anggota meninggal: semua rekening langsung dibekukan. Ahli waris diverifikasi. Jika ada pinjaman outstanding: cek asuransi jiwa/takaful (K035) terlebih dahulu. Untuk BMT: prioritas ahli waris mengikuti ketentuan faraidh.

20. **Masa Clearing Pengunduran Diri.** Status RESIGNING berlaku selama clearing period (H+1 s/d H+30). Seluruh perhitungan diselesaikan sebelum status berubah ke RESIGNED.

### Health Indicators

21. **Komposit Health Score (CAMEL).** Setiap KPI diberi skor 0-100 berdasarkan posisi relatif terhadap threshold. Skor komposit dihitung dengan formula berbobot:

    | Komponen | Bobot | Indikator Utama |
    |----------|-------|-----------------|
    | Capital (Modal) | 20% | CAR ≥ 8% |
    | Asset Quality | 25% | NPL ≤ 5% |
    | Management | 15% | PKG score pengurus |
    | Earnings | 20% | ROA, BOPO |
    | Liquidity | 20% | LDR ≤ 80% |

    Skor akhir = WEIGHTED_SUM(skor per komponen × bobot). Interpretasi: ≥ 80 = SEHAT, 65-79 = CUKUP SEHAT, 50-64 = KURANG SEHAT, < 50 = TIDAK SEHAT.

22. **Rating Kesehatan.** AA (85-100) = Sangat Sehat; A (70-84) = Sehat; BBB (55-69) = Cukup Sehat; BB (40-54) = Kurang Sehat; B (25-39) = Tidak Sehat; C (0-24) = Sangat Tidak Sehat.

23. **Early Warning Action.** Alert yang dipicu harus di-assign ke penanggung jawab dengan deadline tindakan korektif. Alert yang melewati deadline di-eskalasi otomatis ke jabatan lebih tinggi.

24. **Stress Testing Wajib.** Minimal 1x per tahun, jalankan skenario Mild Stress, Moderate Stress, dan Severe Stress. Dokumentasikan apakah koperasi survive setiap skenario.

### Program Pendidikan

25. **Pelatihan Wajib.** COOP_BASIC wajib diselesaikan dalam 30 hari setelah approval keanggotaan. PRODUCT_TRAINING wajib sebelum pinjaman berikutnya dapat diajukan. Untuk BMT: ISLAMIC_FINANCE wajib untuk anggota baru.

26. **Impact Tracking.** Sistem memantau perbedaan NPL rate, savings rate, dan kehadiran RAT antara anggota yang sudah terlatih vs belum terlatih.

### Pembubaran (Dissolution)

27. **Urutan Pembayaran.** Prioritas settlement saat pembubaran: (1) Biaya likuidasi, (2) Gaji staf tertunggak, (3) Utang pajak pemerintah, (4) Utang kreditor (bank, supplier), (5) Simpanan anggota, (6) SHU/surplus pro-rata.

28. **Claim Period Minimum.** Periode klaim kreditor minimal 60 hari setelah pengumuman pembubaran.

29. **Arsip 5 Tahun.** Seluruh records koperasi yang dibubarkan diarsipkan minimal 5 tahun sesuai ketentuan arsip keuangan.

### Multi-Branch Consolidation

30. **Eliminasi Inter-Branch.** Transfer antar cabang harus dieliminasi dalam laporan konsolidasi (net = 0). Tidak ada double-counting aset/kewajiban antar cabang.

    **Akun Eliminasi Inter-Branch:** Seluruh transaksi antar cabang menggunakan akun eliminasi khusus dari COA:
    - **Akun 1901** (Piutang Antar Cabang): digunakan oleh cabang pemberi untuk mencatat tagihan ke cabang penerima.
    - **Akun 2901** (Hutang Antar Cabang): digunakan oleh cabang penerima untuk mencatat kewajiban ke cabang pemberi.
    - Saat konsolidasi: akun 1901 dan 2901 saling di-eliminasi (net = 0).
    - Proses eliminasi dilakukan oleh sistem saat `ConsolidationReportRequested` event diterima.

31. **Empat Level Pelaporan.** (a) Standalone per cabang. (b) Setelah eliminasi inter-branch. (c) Konsolidasi penuh. (d) Regulatory reporting dengan data cabang sebagai lampiran.

---

## Key Decisions & Rationale

### D1. Simplified Mode untuk Koperasi Kecil (K025)

**Keputusan:** Koperasi dengan < 100 anggota dapat menggunakan simplified mode dengan single approval.

**Alasan:** Authority matrix multi-level menambah overhead signifikan untuk koperasi kecil yang biasanya dikelola oleh satu atau dua orang. Template authority matrix disediakan untuk koperasi kecil, medium, dan besar.

### D2. ElectionCompleted → Auto-Create governance_position (K026)

**Keputusan:** Saat pemilihan di RAT selesai, sistem otomatis membuat record `governance_position` untuk yang terpilih.

**Alasan:** Menghindari keharusan admin melakukan entry manual setelah RAT, yang rawan lupa atau delay. Memastikan hak akses sistem segera mencerminkan hasil RAT.

### D3. Dormant Tidak Hangus (K030)

**Keputusan:** Tabungan anggota dormant tetap aman, tidak hangus.

**Alasan:** UU Koperasi melindungi simpanan anggota. Dormant hanya membekukan hak voting dan peminjaman, bukan simpanan.

### D4. CAMELS + Sharia Score untuk BMT (K031)

**Keputusan:** Mode BMT menggunakan framework CAMELS ditambah skor kepatuhan syariah.

**Alasan:** OJK mensyaratkan aspek kepatuhan syariah sebagai dimensi tambahan untuk LKM syariah. DPS melakukan audit syariah terpisah dari audit keuangan reguler.

### D5. Pelaporan Konsolidasi Empat Level (K039)

**Keputusan:** Empat level pelaporan (standalone, post-eliminasi, konsolidasi, regulatory).

**Alasan:** Setiap level melayani kebutuhan audiens yang berbeda — manajer cabang membutuhkan standalone, Bendahara/Ketua membutuhkan konsolidasi, dan OJK/Dinas membutuhkan regulatory format.

---

## Status Lifecycle Keanggotaan (State Machine)

```
APPLIED ──────────────────────────→ APPROVED ──→ ACTIVE ──┬──→ SUSPENDED ──→ ACTIVE
             (approval)                    (bayar pokok)    │   (reinstatement)
                │                                           │
                └──→ REJECTED                               │   EKSKUMUL:
                                                            ├──→ EXPELLED
                                                            │
                                                            │   KELUAR SUKARELA:
                                                            ├──→ RESIGNING (30 hari clearing)
                                                            │         └──→ RESIGNED
                                                            │
                                                            │   KEMATIAN:
                                                            ├──→ DECEASED ──→ SETTLING ──→ SETTLED
                                                            │
                                                            │   TIDAK AKTIF:
                                                            ├──→ DORMANT ──→ ACTIVE (reactivated)
                                                            │
                                                            │   TRANSFER:
                                                            └──→ TRANSFERRED_OUT
                                                            
EXTERNAL ──→ TRANSFERRED_IN ──→ ACTIVE
```

**Keterangan Status Transfer:**

- `TRANSFERRED_OUT`: anggota yang pindah ke koperasi lain — **tidak sama dengan mengundurkan diri**. Simpanan pokok/wajib ditransfer ke koperasi tujuan via mekanisme antar-koperasi. Hak dan kewajiban diselesaikan sebelum status berubah.
- `TRANSFERRED_IN`: anggota yang masuk dari koperasi lain dengan membawa simpanan yang sudah dibayarkan. Riwayat simpanan pokok/wajib dari koperasi asal diverifikasi dan diakui.
- **Faraidh waris (BMT mode):** Untuk anggota yang meninggal dunia, distribusi ke ahli waris mengacu pada hukum faraidh. **Catatan implementasi MVP**: sistem tidak mengotomasi kalkulasi faraidh (terlalu kompleks). Sistem menampilkan data ahli waris yang terdaftar dan meminta upload Surat Keterangan Ahli Waris dari Pengadilan Agama. Distribusi dilakukan manual oleh operator berdasarkan surat tersebut.

```
```

---

## Alur RAT Tahunan

```
H-60 s/d H-30: Persiapan
├── Pengurus susun laporan tahunan
├── Pengawas susun laporan pemeriksaan
├── Bendahara siapkan laporan keuangan + SHU proposal (K016)
└── Draft agenda disusun

H-14 s/d H-7: Undangan
├── Surat undangan dikirim ke semua anggota
├── Agenda & dokumen dilampirkan
└── Deadline konfirmasi kehadiran

Hari H: Pelaksanaan Rapat
├── Registrasi & pengecekan quorum
├── Laporan Pengurus (keuangan, operasional)
├── Laporan Pengawas (audit findings)
├── Diskusi & voting per agenda
├── SHU approval
├── Pemilihan Pengurus/Pengawas (jika ada)
└── Penutupan

H+7 s/d H+30: Pasca-Rapat
├── Notulensi final disusun & disahkan
├── Keputusan RAT didistribusikan ke anggota
├── Auto-create governance_position untuk yang terpilih (SyncEngine)
├── Laporan ke Dinas Koperasi (jika wajib)
└── SHU distribution triggered (K016)
```

---

## KPI Kesehatan Default

| KPI | Formula | Sehat | Peringatan | Kritis | Arah |
|-----|---------|-------|------------|--------|------|
| CAR | Modal Sendiri / ATMR | ≥ 15% | 8-15% | < 8% | Higher |
| NPL Gross | Pinjaman NPL / Total Pinjaman | 0-5% | 5-10% | > 10% | Lower |
| NPL Net | (NPL - PPAP) / Total Pinjaman | 0-3% | 3-6% | > 6% | Lower |
| BOPO | Biaya Op / Pendapatan Op | < 70% | 70-85% | > 85% | Lower |
| ROA | Laba Bersih / Total Aset | ≥ 2% | 1-2% | < 1% | Higher |
| ROE | Laba Bersih / Modal Sendiri | ≥ 15% | 8-15% | < 8% | Higher |
| FDR | Total Pembiayaan / Total Simpanan | 78-92% | 70-78% atau 92-100% | < 70% atau > 100% | Range |
| Cash Ratio | Kas+Bank / Simpanan Jangka Pendek | ≥ 10% | 5-10% | < 5% | Higher |

---

## Integration Points

```
Governance Module (K025-K031, K037-K039)
│
├── K016 (SHU)          — RAT approval triggers distribusi SHU; SHU proposal disiapkan sebelum RAT
├── K015 (Jurnal/COA)   — Laporan keuangan RAT bersumber dari jurnal; branch P&L dari COA
├── K017 (Laporan Reg)  — Health indicators → regulatory reporting; multi-branch → single report
├── K022 (Notifikasi)   — Undangan RAT, pengingat mandatory training, early warning alert
├── K028 (Data Privacy) — CoI disclosures, audit records harus diretain 5 tahun
├── K035 (Asuransi)     — Settlement kematian cek asuransi jiwa/takaful sebelum bebankan ke ahli waris
├── K033 (Collection)   — NPL collection efficiency masuk health indicator
└── K014 (Kas)          — Cash position per branch masuk executive dashboard dan KPI likuiditas
```

---

## RBAC Summary

### Governance Positions

| Permission | Teller | Supervisor | Manager | Bendahara | Ketua | Pengawas |
|------------|--------|------------|---------|-----------|-------|----------|
| Create operational position | - | - | v | - | - | - |
| Create management position | - | - | - | v | v | - |
| Create strategic position | - | - | - | - | v (RAT) | - |
| View all positions | - | - | v | v | v | v |
| Configure authority limits | - | - | - | v | v | - |

### RAT

| Permission | Ketua | Sekretaris | Bendahara | Pengawas | Anggota | Admin |
|------------|-------|------------|-----------|----------|---------|-------|
| Create RAT meeting | v | v (draft) | - | - | - | v |
| Conduct voting | v | v (record) | - | v (witness) | v (vote) | - |
| Finalize minutes | v | v | - | v (approve) | - | v |
| Manage elections | v | v | - | v | - | v |

---

## Dual-Mode Governance

| Aspek | Koperasi Umum | BMT (Islamic) |
|-------|---------------|---------------|
| Dewan Pengawas Syariah | Tidak ada | WAJIB ≥ 1 anggota |
| Approval Pembiayaan Besar | Manager + Bendahara | + DPS |
| Produk Baru | Approval Pengurus | + DPS |
| Laporan Keuangan | PSAK/SAK ETAP | PSAK Syariah + review DPS |
| RAT — Laporan DPS | Tidak ada | WAJIB laporan kepatuhan syariah |
| Ekskumul | Alasan standar | + pelanggaran prinsip syariah |
| Dissolution | Standard | + zakat/infaq disalurkan dulu; dana sosial tidak dibagi ke anggota |
| Ahli Waris | KUHPerdata | Faraidh (hukum Islam) |
| Skenario Stress Test | Standard CAMEL | + skenario perubahan nisbah bagi hasil |
