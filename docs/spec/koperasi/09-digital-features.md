# 09 - Fitur Digital: Uang Saku Digital, Biometrik, dan Mobile App

Dokumen ini mendeskripsikan tiga pilar fitur digital Modul Koperasi SekolahPro: (1) Uang Saku Digital sebagai antarmuka pengeluaran siswa berbasis tabungan, (2) Autentikasi Biometrik untuk keamanan transaksi bernilai tinggi, dan (3) Strategi Mobile App yang mengakomodasi empat segmen pengguna utama. Ketiga fitur ini saling terintegrasi dan dirancang dengan prinsip *privacy by design* sesuai UU PDP (ADR-K028).

---

## ADR References

| ADR | Judul | Status |
|-----|-------|--------|
| ADR-K021 | Uang Saku Digital (Kartu Belanja Siswa) | Accepted |
| ADR-K032 | Biometric Authentication | Accepted |
| ADR-K036 | Mobile App Strategy | Accepted |

---

## Domain Entities

### Uang Saku Digital (ADR-K021)

| Entity | Deskripsi |
|--------|-----------|
| `spending_control` | Konfigurasi kontrol belanja siswa — limit harian, per transaksi, mingguan, bulanan, pembatasan kategori, dan pembatasan jam. Single source of truth untuk semua enforcement. |
| `ewallet_card` | Kartu identifikasi siswa — NFC, QR statis, QR dinamis, atau barcode. Hanya sebagai identifier; saldo tetap di rekening tabungan. |
| `ewallet_proxy` | Wali kelas yang bertindak sebagai proxy pembelian untuk siswa muda (SD/MI). Terikat permission dan limit terpisah. |

### Biometrik (ADR-K032)

| Entity | Deskripsi |
|--------|-----------|
| `biometric_enrollment` | Rekaman pendaftaran biometrik nasabah (sidik jari, wajah, suara). Menyimpan hash template, bukan data mentah biometrik jika on-device. |
| `biometric_verification_log` | Log setiap verifikasi biometrik — tujuan, hasil, confidence score. Immutable audit trail. |

### Mobile App (ADR-K036)

| Entity | Deskripsi |
|--------|-----------|
| `mobile_app_config` | Konfigurasi per tenant untuk branding app, feature flags, limit offline, dan keamanan sesi. |

---

## Business Rules

### Uang Saku Digital

1. **Bukan Uang Elektronik.** Uang Saku Digital adalah *spending interface* di atas rekening tabungan, BUKAN instrumen uang elektronik sebagaimana PBI 20/6/PBI/2018. Tidak ada stored value terpisah; semua transaksi langsung mendebit tabungan nasabah.

2. **Terminologi Publik.** Istilah "E-Wallet" TIDAK digunakan dalam UI atau dokumen publik. Gunakan: "Uang Saku Digital" atau "Kartu Belanja Siswa".

3. **Top-up = Setoran Tabungan.** Semua jalur pengisian saldo — teller, transfer bank, payroll deduction, atau transfer internal orang tua ke anak — WAJIB dicatat sebagai setoran tabungan (ADR-K011). Tidak ada jalur bypass.

4. **Pembatasan Scope Internal.** QR/NFC hanya dapat digunakan di POS yang terdaftar dalam tenant yang sama. Transaksi dari POS tidak dikenal langsung ditolak.

5. **Spending Control Merupakan Satu-Satunya Sumber Kebenaran.** Tiga sistem yang membaca kontrol belanja — K019 (POS Kantin/Toko), K021 (Kartu Belanja), dan K023 (Parent Portal) — semuanya membaca dari tabel `spending_control`. Tidak ada duplikasi konfigurasi.

6. **Validasi Bertingkat.** Setiap transaksi kartu memvalidasi secara berurutan: (a) kartu aktif, (b) spending control aktif, (c) daily/per-transaction limit, (d) time restriction, (e) category restriction, (f) available balance, (g) PIN jika di atas threshold. Gagal di satu step langsung menolak transaksi.

7. **Keamanan PIN.** PIN 6 digit di-hash (bcrypt/argon2). 3x salah PIN → kartu freeze otomatis 30 menit. 5x salah dalam 24 jam → freeze manual oleh admin.

8. **Kartu Hilang.** Saldo tidak hilang saat kartu hilang karena saldo berada di tabungan, bukan di kartu. Alur: freeze lama → deactivate → terbitkan kartu baru (menggunakan spending_control yang sama).

9. **Proxy Wali Kelas.** Wali kelas dapat bertindak sebagai proxy pembelian untuk siswa muda. Proxy tetap terikat limit yang di-set orang tua. Semua transaksi proxy tercatat dengan `created_by = proxy_user_id`.

10. **Mode Boarding.** Untuk siswa asrama, flag `is_boarding = true` memungkinkan meal plan otomatis dan emergency spending approval tanpa limit check oleh wali asrama.

11. **Eskalasi ke Luar Sekolah.** Jika scope diperluas ke merchant luar sekolah, WAJIB konsultasi legal dan kemungkinan butuh lisensi BI. ADR ini secara eksplisit membatasi scope ke lingkungan internal.

### Biometrik

12. **On-Device Preferred.** Biometric template LEBIH DIUTAMAKAN disimpan di Secure Enclave/Keystore perangkat. Server hanya menyimpan token hash dan public key. Ini adalah persyaratan kepatuhan UU PDP (data biometrik = data sensitif).

13. **Liveness Detection Wajib untuk Face Recognition.** Minimal 3 tantangan liveness aktif (kedipan, palingkan kepala, senyum). Foto, video, mask, dan deepfake harus ditolak.

14. **Threshold Biometrik per Operasi.** Penarikan > Rp 25 juta memerlukan fingerprint. Penarikan > Rp 50 juta memerlukan fingerprint + face. Login teller dan approval pinjaman selalu memerlukan fingerprint. Setoran tidak memerlukan biometrik (PIN saja cukup).

15. **Fallback Chain.** Jika biometrik gagal 3x: PIN + OTP. Jika gagal lagi: verifikasi manual oleh staf (KTP + selfie, perlu approval supervisor). Jika semua gagal: transaksi ditolak, nasabah diarahkan ke cabang.

16. **Consent Biometrik.** Enrollmen biometrik WAJIB memiliki `consent_id` yang terhubung ke `consent_record` (ADR-K028). Consent eksplisit, tidak bisa di-bundle dengan consent lain.

17. **Retensi Data.** Data biometrik disimpan 5 tahun pasca penutupan rekening. Untuk mode BMT, DPS harus approve kebijakan biometrik.

### Mobile App

18. **Flutter Multi-Platform.** Satu codebase Flutter → iOS + Android. Backend REST API sama dengan web dashboard (Vernon API pattern).

19. **Empat Varian App.** (a) Nasabah App — fitur lengkap anggota, distribusi Play Store/App Store. (b) Parent App — monitoring dan top-up anak, distribusi publik. (c) Student E-Wallet — QR payment siswa, distribusi internal APK sideload. (d) Collector App — penagihan lapangan dengan GPS dan foto bukti, distribusi internal.

20. **Offline-First.** Cache TTL: saldo 5 menit, riwayat transaksi 1 jam, profil 24 jam. Antrian offline maksimal 10 transaksi, periode offline maksimal 24 jam. Transaksi bernilai tinggi TIDAK dapat dilakukan offline.

21. **Certificate Pinning.** Semua komunikasi jaringan menggunakan certificate pinning untuk mencegah serangan MITM. TLS 1.3 minimum.

22. **Device Registration.** Maksimal 2 perangkat per nasabah. Session timeout 5 menit inaktif (dapat dikonfigurasi). Auto-logout dan clear cache saat logout.

23. **Root/Jailbreak Detection.** Perangkat yang terdeteksi root/jailbreak mendapat peringatan dan fitur terbatas. Tidak ada akses penuh ke fitur keuangan sensitif.

---

## Key Decisions & Rationale

### D1. Spending Interface, Bukan Saldo Terpisah (K021)

**Keputusan:** Uang Saku Digital bukan saldo tersendiri melainkan layer kontrol di atas tabungan.

**Alasan:** Menghindari double-balance problem (dua saldo yang harus di-reconcile), menyederhanakan laporan keuangan, dan mereuse semua infrastruktur tabungan yang sudah ada. Alternatif saldo terpisah ditolak karena menambah kompleksitas signifikan tanpa manfaat nyata.

### D2. Stored Value Card vs Identifier-Only Card (K021)

**Keputusan:** Kartu hanya berfungsi sebagai identifier — tidak menyimpan nilai apapun.

**Alasan:** Saldo tersimpan di kartu fisik tidak bisa di-reconcile real-time, berisiko kehilangan saldo jika kartu hilang/rusak, dan memerlukan hardware card-writer yang mahal.

### D3. On-Device Biometric Storage (K032)

**Keputusan:** Template biometrik disimpan di Secure Enclave/Keystore perangkat, bukan server.

**Alasan:** Jika server mengalami breach, tidak ada data biometrik yang bocor. Ini juga merupakan best practice kepatuhan UU PDP — data sensitif seminimal mungkin di server.

### D4. Flutter untuk Mobile (K036)

**Keputusan:** Flutter dipilih sebagai framework mobile.

**Alasan:** Single codebase untuk iOS dan Android mengurangi biaya development. Tim sudah familiar dengan Flutter. Performance mendekati native. Offline-first architecture lebih mudah diimplementasikan dengan Hive/Isar.

### D5. Empat Varian App, Bukan Satu (K036)

**Keputusan:** Empat app terpisah berdasarkan segmen pengguna (nasabah, orang tua, siswa, collector).

**Alasan:** Masing-masing segmen memiliki use case yang sangat berbeda dan flow keamanan yang berbeda. Menggabungkan semua dalam satu app akan menjadikannya terlalu kompleks dan membingungkan pengguna.

---

## Integration Points

### Uang Saku Digital dengan Domain Lain

```
K021 (Uang Saku Digital)
│
├── K002/K005 (Rekening/Tabungan)    — Saldo aktual; top-up = setoran tabungan
├── K011 (Transaksi)                  — Setiap debit kartu = 1 transaksi K011
├── K019 (Toko/Kantin POS)            — POS membaca spending_control untuk enforce limit
├── K022 (Notifikasi)                 — Alert harian orang tua, alert limit exceeded
├── K023 (Dashboard/Portal)           — Parent portal: set limit, monitoring spending
└── K024 (Integrasi API)              — Top-up via Virtual Account dari luar
```

### Biometrik dengan Domain Lain

```
K032 (Biometrik)
│
├── K011 (Transaksi)                  — Verifikasi sebelum penarikan besar
├── K007 (Pinjaman)                   — Verifikasi sebelum disbursement
├── K028 (Data Privacy)               — Consent record wajib sebelum enrollment
├── K036 (Mobile App)                 — Biometric login dan verifikasi in-app
└── K025 (Governance)                 — Login teller dan approval manajerial
```

### Mobile App dengan Domain Lain

```
K036 (Mobile App)
│
├── K021 (Uang Saku Digital)          — Student E-Wallet untuk QR payment
├── K022 (Notifikasi)                 — FCM push notification
├── K023 (Dashboard/Portal)           — Portal nasabah dan orang tua via mobile
├── K033 (Loan Collection)            — Collector App dengan GPS dan log kunjungan
├── K032 (Biometrik)                  — Biometric login dan verifikasi transaksi
└── K024 (Integrasi API)              — REST API sama dengan web dashboard
```

---

## Alur Pembayaran Kartu Belanja (Flow Diagram)

```
Siswa tap/scan kartu di POS
         │
         v
┌────────────────────────────┐
│ Step 1: Identifikasi        │
│ Lookup card → nasabah       │
│ Cek card.status = ACTIVE    │
└───────────┬────────────────┘
            │
            v
┌────────────────────────────┐
│ Step 2: Validasi Spending   │
│ Cek daily_limit             │
│ Cek per_transaction_limit   │
│ Cek time_restrictions       │
│ Cek category_restrictions   │
│ Cek available_balance       │
└───────────┬────────────────┘
            │ ALL PASS
            v
┌────────────────────────────┐
│ Step 3: PIN (jika required) │
│ amount > pin_threshold?     │
│ → minta PIN                 │
└───────────┬────────────────┘
            │
            v
┌────────────────────────────┐
│ Step 4: Execute             │
│ Create transaksi K011       │
│ Create penjualan K019       │
│ Update saldo tabungan       │
│ (dalam 1 DB transaction)    │
└───────────┬────────────────┘
            │
            v
┌────────────────────────────┐
│ Step 5: Receipt & Log       │
│ Print receipt POS           │
│ Log spending untuk orang tua│
│ Notifikasi K022 (opsional)  │
└────────────────────────────┘
```

---

## RBAC Summary

### Uang Saku Digital

| Permission | Teller | Supervisor | Manager | Admin | Parent |
|------------|--------|------------|---------|-------|--------|
| Issue card | v | v | v | v | - |
| Freeze card | - | v | v | v | v |
| Unfreeze card | - | v | v | v | v |
| Set spending limits | - | - | - | v | v |
| View child spending | - | - | - | - | v |
| Set PIN | - | - | - | - | v |
| Reset PIN | v | v | v | v | v |

### Biometrik

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Enroll nasabah biometric | v | v | v | v |
| Disable nasabah biometric | - | v | v | v |
| Configure thresholds | - | - | - | v |
| Purge biometric data | - | - | - | v |

---

## Catatan Regulasi

### Analisis Klasifikasi Uang Saku Digital vs PBI 20/6/PBI/2018

| Kriteria BI | Uang Saku Digital | Kesimpulan |
|-------------|-------------------|------------|
| Saldo tersimpan di server penerbit? | Tidak — di rekening tabungan | Bukan e-money |
| Ada instrumen pembayaran dengan nilai? | Kartu = identifier saja | Bukan e-money |
| Top-up hasilkan saldo terpisah? | Tidak — top-up = setoran tabungan | Bukan e-money |
| Bisa transfer antar pengguna? | Tidak — hanya di POS internal | Bukan e-money |

**Risiko yang wajib dimitigasi:** (1) Semua top-up HARUS melewati rekening tabungan — tidak ada jalur bypass. (2) QR hanya bisa discan POS terdaftar dalam tenant. (3) Jika scope diperluas ke luar sekolah, wajib konsultasi legal dan kemungkinan butuh lisensi BI.
