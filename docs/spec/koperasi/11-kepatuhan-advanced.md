# 11 - Kepatuhan Lanjutan: AML/CFT, Privasi Data (UU PDP), dan Business Continuity

Dokumen ini menguraikan tiga pilar kepatuhan lanjutan Modul Koperasi SekolahPro: (1) Anti-Money Laundering dan Counter Financing of Terrorism (AML/CFT) sesuai UU TPPU No. 8/2010 dan POJK, (2) Perlindungan Data Pribadi sesuai UU PDP No. 27/2022, dan (3) Business Continuity & Disaster Recovery (BCP/DRP) sesuai POJK 13/POJK.03/2016. Ketiga domain ini saling berkaitan — BCP melindungi keberlangsungan data yang wajib diretain oleh AML dan UU PDP.

---

## ADR References

| ADR | Judul | Status |
|-----|-------|--------|
| ADR-K027 | AML/CFT Compliance | Accepted |
| ADR-K028 | Data Privacy & Personal Data Protection (UU PDP) | Accepted |
| ADR-K029 | Business Continuity & Disaster Recovery (BCP/DRP) | Accepted |

---

## Domain Entities

### AML/CFT (K027)

| Entity | Deskripsi |
|--------|-----------|
| `aml_customer_due_diligence` | CDD per nasabah — risk level, risk score, CDD level (SDD/CDD/EDD), screening PEP, beneficial owner, status, next review date. |
| `aml_transaction_monitoring` | Alert monitoring transaksi — rule yang dipicu, tipe alert, severity, hasil investigasi, referensi LTKM ke PPATK. |
| `aml_monitoring_rule` | Definisi aturan monitoring — threshold, velocity, pattern, anomaly. Parameter dapat dikonfigurasi per tenant. |
| `aml_report` | Laporan LTKM/TKM yang dikirim ke PPATK — referensi, status pengakuan. |

### Data Privacy (K028)

| Entity | Deskripsi |
|--------|-----------|
| `consent_record` | Rekaman persetujuan pengolahan data per nasabah — tipe consent, legal basis, teks yang ditampilkan, versi, metode. Mendukung parental consent untuk nasabah di bawah 17 tahun. |
| `data_subject_request` | Permintaan hak subjek data (akses, koreksi, hapus, portabilitas, keberatan) — SLA, status, penanggung jawab. |
| `data_breach` | Rekaman insiden pelanggaran data — tipe, dampak, jumlah nasabah terdampak, notifikasi ke otoritas dan subjek. |
| `data_processing_agreement` | Perjanjian pemrosesan data dengan pihak ketiga (DPA) — nama pemroses, kategori data, lokasi server, cross-border flag. |

### BCP/DRP (K029)

| Entity | Deskripsi |
|--------|-----------|
| `incident_record` | Rekaman insiden sistem — severity (P1-P4), tipe, timeline, dampak, tindakan, root cause, status. |

---

## Business Rules

### AML/CFT

1. **Risk-Based Approach.** Setiap nasabah memiliki risk assessment dengan empat dimensi: Customer Risk, Product Risk, Geographic Risk, dan Delivery Channel Risk. Skor gabungan menentukan level CDD yang diperlukan.

2. **Tiga Level CDD.**
   - **SDD (Simplified):** Nasabah risiko rendah. Verifikasi identitas dasar. Review 3 tahun.
   - **CDD (Standard):** Nasabah risiko menengah. Verifikasi identitas + tujuan membuka rekening. Review 2 tahun.
   - **EDD (Enhanced):** Nasabah risiko tinggi — WAJIB. Verifikasi mendalam + sumber dana + sumber kekayaan. Approval senior management. Review 1 tahun.

3. **EDD Wajib Untuk.** PEP (Politically Exposed Person) dan kerabat/rekan dekat PEP; nasabah dari negara risiko tinggi (FATF list); transaksi unusual/suspicious; nasabah yang menolak memberikan informasi CDD; entity dengan struktur kepemilikan kompleks.

4. **Monitoring Rules Default.** Delapan aturan monitoring default aktif sejak awal:
   - `THRESHOLD_001`: Cash ≥ Rp 500 juta dalam satu transaksi
   - `THRESHOLD_002`: Transfer ≥ Rp 1 miliar dalam satu transaksi
   - `VELOCITY_001`: ≥ 3 transaksi > Rp 50 juta dalam 24 jam
   - `VELOCITY_002`: ≥ 5 setoran dalam 1 hari oleh orang yang sama
   - `STRUCTURE_001`: Multiple deposit tepat di bawah threshold (structuring pattern)
   - `RAPID_001`: Dana masuk → keluar dalam 24 jam
   - `DORMANT_001`: Rekening dormant tiba-tiba aktif dengan nominal besar (> Rp 50 juta)
   - `MISMATCH_001`: Transaksi tidak sesuai profil nasabah

5. **Pelaporan LTKM ke PPATK.** Wajib dilaporkan ≤ 3 hari kerja setelah indikasi suspicious terdeteksi. Format sesuai template PPATK. Identitas pelapor dilindungi (whistleblower protection).

6. **Pelaporan TKM Bulanan.** Cash transaction ≥ Rp 500 juta atau transfer ≥ Rp 1 miliar wajib dilaporkan bulanan ke PPATK, paling lambat tanggal 10 bulan berikutnya.

7. **Retensi Record AML.** CDD records diretain selama relasi bisnis aktif + 5 tahun. Transaction records: 5 tahun. AML alerts dan investigasi: 5 tahun. LTKM/TKM reports: 5 tahun. PEP screening results: 5 tahun.

8. **Auto-Block Hati-Hati.** Rule dengan `auto_block = true` dapat memblokir transaksi legitimate. Default semua rule menggunakan `auto_alert = true, auto_block = false`. Auto-block hanya diaktifkan setelah tuning dan validasi.

### Data Privacy (UU PDP)

9. **Klasifikasi Data Empat Tingkat.** PUBLIC → INTERNAL → CONFIDENTIAL → HIGHLY_CONFIDENTIAL. Data keuangan sensitif (saldo, transaksi, pinjaman), data biometrik, data KTP/NPWP, data anak di bawah umur, dan password/PIN masuk kategori HIGHLY_CONFIDENTIAL.

10. **Legal Basis per Use Case.** Setiap pemrosesan data harus memiliki legal basis yang jelas:
    - Membuka rekening: Contractual (tidak perlu consent terpisah)
    - Laporan ke OJK/PPATK: Legal Obligation (tidak perlu consent)
    - Marketing produk baru: Consent — wajib opt-in eksplisit
    - Credit scoring: Legitimate Interest (wajib disclosure)
    - Berbagi data ke orang tua siswa: Contractual + Consent

11. **Consent untuk Anak < 17 Tahun.** Wajib parental consent untuk nasabah di bawah umur 17 tahun. Consent harus mengidentifikasi orang tua/wali dan hubungan keluarga.

12. **Hak Subjek Data — SLA.** Acknowledgment wajib ≤ 72 jam. Penyelesaian: Access/Erasure/Portability ≤ 14 hari; Rectification/Restriction/Withdraw Consent ≤ 7 hari.

13. **Erasure dan Kewajiban Retensi.** Data yang wajib disimpan regulasi (transaksi, CDD, PPATK) TIDAK BOLEH dihapus. Solusinya adalah anonymize — hapus identitas, pertahankan record untuk audit. Nasabah wajib diberikan penjelasan data mana yang tidak bisa dihapus dan alasannya.

14. **Data Breach Notification Timeline.** ≤ 72 jam setelah mengetahui breach → notifikasi ke otoritas (Kementerian/Komisi PDP). Notifikasi langsung ke subjek data jika risk tinggi. Notifikasi harus berisi: jenis breach, data terdampak, langkah mitigasi, kontak darurat.

15. **DPA dengan Pihak Ketiga.** WAJIB ada Data Processing Agreement (DPA) tertulis sebelum sharing data ke pihak ketiga. Pihak ketiga tidak boleh menggunakan data untuk tujuan lain. Pihak ketiga wajib melaporkan breach dalam 24 jam ke koperasi.

16. **Data Sovereignty.** Data nasabah TIDAK BOLEH disimpan di luar Indonesia tanpa consent nasabah eksplisit. Cloud provider yang digunakan harus beroperasi di region Indonesia (Jakarta).

17. **Privacy by Design — Lima Prinsip.** (a) Data Minimization: hanya kumpulkan yang benar-benar dibutuhkan. (b) Purpose Limitation: data KYC tidak boleh dipakai marketing tanpa consent terpisah. (c) Storage Limitation: anonymize otomatis setelah periode retensi berakhir. (d) Access Control: teller hanya lihat data yang relevan untuk transaksinya. (e) Encryption: data sensitif di-encrypt at-rest dan in-transit; KTP/NPWP di-mask di display (tampil 4 digit terakhir saja).

18. **Biometrik — Data Paling Sensitif.** Data biometrik tidak boleh disimpan di server jika bisa dilakukan on-device (ref ADR-K032). Jika harus di server: enkripsi AES-256, key management terpisah, akses hanya oleh biometric verification service.

### Business Continuity (BCP/DRP)

19. **Recovery Objectives.** RPO: Critical data (saldo, transaksi) = 0 (zero data loss). RTO: Transaction processing ≤ 15 menit (failover); Online banking ≤ 30 menit; Full system ≤ 4 jam.

20. **Backup Strategy.** PostgreSQL: WAL archiving kontinu + full backup harian (02:00 WIB) + incremental setiap 4 jam. Retensi: 30 hari harian, 12 minggu mingguan, 12 bulan bulanan, 5 tahun tahunan. Storage di primary + off-site (region berbeda).

21. **Tiga Tier DR per Ukuran Koperasi.**
    - **Tier 3 (Cold):** Koperasi < 100 anggota. RPO: 24 jam, RTO: 4 jam. Minimum.
    - **Tier 2 (Warm):** 100-1000 anggota. RPO: < 1 menit, RTO: 30 menit.
    - **Tier 1 (Active-Active):** > 1000 anggota. RPO: 0, RTO: 15 menit.

22. **Empat Level Severity Insiden.**
    - **P1 (Critical):** Sistem down, data loss, security breach → respons < 15 menit, eskalasi on-call → manajer → ketua.
    - **P2 (High):** Major feature down → respons < 30 menit.
    - **P3 (Medium):** Minor feature impaired, ada workaround → respons < 2 jam.
    - **P4 (Low):** Kosmetik, tidak mendesak → next business day.

23. **Prosedur Offline.** Jika sistem benar-benar down: setoran menggunakan form manual triplicate; penarikan hanya ≤ Rp 2 juta dengan catatan manual; angsuran diterima dengan kwitansi manual. Semua transaksi manual di-input ulang saat sistem recovery, dengan verifikasi supervisor sebelum re-open ke nasabah.

24. **Jadwal Testing BCP.**
    - Restore test dari backup: bulanan
    - Failover test: triwulanan
    - Full DR test: tahunan
    - Incident response drill: semester
    - Penetration test: tahunan
    - BCP review: tahunan

25. **Vernon Cache Tidak Perlu Backup.** Redis/Elasticsearch read cache dapat direbuild dari PostgreSQL (target ≤ 30 menit, prioritas tenant aktif). Tidak perlu mekanisme backup terpisah.

---

## Key Decisions & Rationale

### D1. Rule-Based Monitoring Dulu, ML Nanti (K027)

**Keputusan:** Monitoring AML Phase 1 menggunakan rule-based (threshold, velocity, pattern), bukan ML.

**Alasan:** Rule-based lebih transparan, mudah diaudit oleh regulator, dan tidak memerlukan dataset historis yang belum ada. ML-based anomaly detection direncanakan di Phase 3 (2028) setelah ada data historis yang cukup (ADR-K040).

### D2. Anonymize vs Hapus untuk Erasure (K028)

**Keputusan:** Data yang terkena kewajiban retensi regulasi di-anonymize, bukan dihapus total.

**Alasan:** UU TPPU dan OJK mewajibkan retensi record transaksi dan CDD minimal 5 tahun. Menghapus total akan melanggar regulasi. Anonymize mempertahankan record untuk audit sambil menghilangkan identitas nasabah.

### D3. Tier DR Berdasarkan Ukuran Koperasi (K029)

**Keputusan:** Tiga tier DR dengan biaya berbeda disesuaikan dengan ukuran koperasi.

**Alasan:** Active-Active HA terlalu mahal untuk koperasi kecil dengan 50 anggota. Cold standby sudah cukup untuk koperasi kecil dengan risiko yang proporsional. Koperasi besar dengan > 1000 anggota memiliki exposure lebih tinggi sehingga justifies biaya Tier 1.

### D4. SLA 72 Jam untuk Acknowledgment Data Subject Request (K028)

**Keputusan:** Acknowledgment dalam 72 jam sesuai UU PDP, penyelesaian 7-14 hari.

**Alasan:** UU PDP Indonesia mengadopsi standar GDPR untuk timeline respons. 72 jam acknowledgment memberikan waktu bagi koperasi untuk melakukan triase sebelum proses penuh dimulai.

---

## Klasifikasi Data

```
Data Classification Matrix:
┌─────────────────────────────────────────────────────────────────┐
│ HIGHLY CONFIDENTIAL                                             │
│ ├── KTP, NPWP, foto, biometrik (sidik jari, wajah)             │
│ ├── Saldo rekening, riwayat transaksi detail, outstanding loan  │
│ ├── Data siswa < 17 tahun (e-wallet, spending history)          │
│ └── Password, PIN, kunci enkripsi                               │
├─────────────────────────────────────────────────────────────────┤
│ CONFIDENTIAL                                                    │
│ ├── Nama, alamat, nomor HP, email                               │
│ ├── Nomor anggota, status keanggotaan                           │
│ └── Produk yang dimiliki (tanpa saldo)                          │
├─────────────────────────────────────────────────────────────────┤
│ INTERNAL                                                        │
│ ├── Kebijakan internal (tanpa data pribadi)                     │
│ └── Laporan aggregate anonim                                    │
├─────────────────────────────────────────────────────────────────┤
│ PUBLIC                                                          │
│ ├── Nama dan alamat koperasi                                    │
│ └── Produk yang ditawarkan (deskripsi umum)                     │
└─────────────────────────────────────────────────────────────────┘
```

---

## Compliance Checklist

### AML/CFT

- [ ] CDD dilakukan untuk semua nasabah baru
- [ ] EDD dilakukan untuk nasabah risiko tinggi dan PEP
- [ ] 8 monitoring rules aktif dan dikonfigurasi
- [ ] Alert yang terpicu diinvestigasi dalam 3 hari kerja
- [ ] LTKM disiapkan dan dikirim ke PPATK dalam 3 hari kerja
- [ ] TKM bulanan dikirim paling lambat tanggal 10 bulan berikutnya
- [ ] Retensi record AML 5 tahun terkonfigurasi
- [ ] Staff compliance mendapat AML training berkala
- [ ] Review CDD nasabah EDD setiap tahun, CDD setiap 2 tahun, SDD setiap 3 tahun

### Data Privacy (UU PDP)

- [ ] Setiap pemrosesan data memiliki legal basis yang terdokumentasi
- [ ] Consent nasabah baru direkam sebelum pembukaan rekening
- [ ] Parental consent ada untuk nasabah < 17 tahun
- [ ] DPA sudah ditandatangani dengan semua pihak ketiga yang memproses data
- [ ] Prosedur respons data subject request (DSR) tersedia (SLA 72 jam / 7-14 hari)
- [ ] Privacy Impact Assessment dilakukan sebelum fitur baru diluncurkan
- [ ] Prosedur data breach notification ≤ 72 jam ke otoritas terdokumentasi
- [ ] Auto-anonymize data setelah periode retensi terkonfigurasi
- [ ] Data sensitif di-encrypt at-rest (AES-256) dan in-transit (TLS 1.3)
- [ ] Display masking aktif untuk KTP/NPWP (tampil 4 digit terakhir)

### BCP/DRP

- [ ] Tier DR dipilih sesuai ukuran koperasi
- [ ] Backup harian terkonfigurasi dan dimonitor
- [ ] Prosedur offline (manual form) tersedia di setiap kantor cabang
- [ ] Restore test dari backup dilakukan setiap bulan
- [ ] Failover test dilakukan setiap kuartal
- [ ] On-call engineer tersedia 7x24 untuk P1
- [ ] Incident response drill dilakukan setiap semester
- [ ] BCP document di-review dan diupdate setiap tahun

---

## Integration Points

### AML/CFT

```
K027 (AML/CFT)
│
├── K001 (Nasabah)        — CDD terhubung ke data KYC nasabah
├── K011 (Transaksi)      — TransaksiCreatedEvent → trigger monitoring rules
├── K022 (Notifikasi)     — Alert escalation → notifikasi compliance officer
├── K028 (Data Privacy)   — CDD records termasuk data highly confidential
├── K031 (Health Indicators) — NPL, suspicious transactions → health score
└── K024 (Integrasi API)  — Ekspor LTKM/TKM ke PPATK (manual atau otomatis)
```

### Data Privacy

```
K028 (Data Privacy)
│
├── K001 (Nasabah)        — Consent direkam saat onboarding nasabah baru
├── K032 (Biometrik)      — Biometric enrollment memerlukan consent_id dari K028
├── K021 (Uang Saku Digital) — Data anak < 17 tahun = HIGHLY_CONFIDENTIAL
├── K022 (Notifikasi)     — Nomor HP dan email nasabah harus diretain aman
├── K018 (Zakat/BMT)      — Data mustahik = HIGHLY_CONFIDENTIAL
└── K024 (Integrasi API)  — DPA wajib untuk WhatsApp API, payment gateway, dll
```

### BCP/DRP

```
K029 (BCP/DRP)
│
├── Semua domain         — Data keuangan wajib RPO=0 (WAL archiving)
├── K027 (AML/CFT)       — CDD records wajib retensi 5 tahun; backup mencakup ini
├── K028 (Data Privacy)  — Backup terenkripsi; recovery tidak boleh expose PII
├── K022 (Notifikasi)    — Template WhatsApp untuk komunikasi insiden pre-defined
└── K025 (Governance)    — Untuk BMT: DPS di-notify saat P1/P2; verifikasi post-recovery
```

---

## RBAC Summary

### AML/CFT

| Permission | Teller | Compliance | Manager | Admin |
|------------|--------|------------|---------|-------|
| Perform CDD (basic) | v | v | v | v |
| Perform EDD | - | v | v | v |
| View monitoring alerts | - | v | v | v |
| Investigate alerts | - | v | v | v |
| File LTKM to PPATK | - | - | v | v |
| Configure monitoring rules | - | - | - | v |

### Data Privacy

| Permission | Teller | Compliance | Manager | Admin |
|------------|--------|------------|---------|-------|
| Record consent | v | v | v | v |
| Process data subject request | - | v | v | v |
| Report breach to authority | - | - | v | v |
| Anonymize/delete data | - | v | v | v |
| Configure data classification | - | - | - | v |

### BCP/DRP

| Permission | On-Call | Manager | Admin | Ketua |
|------------|---------|---------|-------|-------|
| Create/acknowledge incident | v | v | v | v |
| Acknowledge P1/P2 | - | v | v | v |
| Trigger failover | - | - | v | v |
| Approve manual transactions (offline) | - | v | v | - |
| Configure backup schedule | - | - | v | - |
| Approve BCP changes | - | - | v | v |

---

## Dual-Mode Perbedaan

| Aspek | Koperasi Umum | BMT (Islamic) |
|-------|---------------|---------------|
| **AML** — Monitoring | Standard flows | + zakat/infaq/sadaqah flows |
| **AML** — PEP screening | Standard | + ulama/influencer PEP |
| **AML** — Produk risiko tinggi | Cash-intensive | + hawala-like informal transfer |
| **Data Privacy** — Sensitive data | Standar | + data ibadah (zakat, infaq), data mustahik |
| **Data Privacy** — DPS akses | Tidak ada | DPS dapat akses data untuk audit syariah |
| **BCP** — Priority data | Transaksi, saldo | + zakat/infaq records |
| **BCP** — Post-recovery | Standard | DPS harus verifikasi post-recovery; dana sosial harus verified |
| **BCP** — DPS notification | Tidak perlu | DPS di-notify saat P1/P2 |
