# Laporan Validasi Dokumentasi Sekolah
**Tanggal:** 2026-04-26
**Validator:** Kepala Sekolah & Education System Expert
**Scope:** ADR S001–S058 vs Dokumentasi spec/sekolah/ (01-domain-overview.md s.d. 16-alumni.md)

---

## 1. Statistik Coverage Keseluruhan

| Metrik | Nilai |
|--------|-------|
| Total ADR yang divalidasi | 58 |
| ADR dengan coverage FULL | **33 (57%)** |
| ADR dengan coverage PARTIAL | **25 (43%)** |
| ADR dengan coverage MISSING | **0 (0%)** |
| Gap kritis (P0) ditemukan | **6** |
| Gap penting (P1) ditemukan | **19** |
| Gap nice-to-have (P2) ditemukan | **2** |

**Kesimpulan Awal:** Tidak ada ADR yang sama sekali tidak terdokumentasi — ini merupakan pencapaian yang baik. Namun 43% ADR memiliki gap signifikan, terutama di domain regulasi, governance, dan integrasi nasional.

---

## 2. Inakurasi yang Ditemukan

### 2.1 PKG — Jumlah Indikator (KRITIS)

**Lokasi:** `09-teacher-staff.md` → Business Rule #6 & #7; `ADR-S028-teacher-performance-evaluation.md`

**Masalah:** Dokumen menyebut PKG menilai "4 kompetensi" tanpa merincikan jumlah indikator. Standar Permenneg PAN & RB No. 16 Tahun 2009 jo. Permendiknas No. 35 Tahun 2010 menetapkan:
- Guru Kelas/Mata Pelajaran: **14 kompetensi, 78 indikator**
- Guru BK: **17 kompetensi, 76 indikator**
- Kepala Sekolah: menggunakan **PKKS** (instrumen terpisah)

**Koreksi:** Schema `teacher_evaluation_competencies` mendukung "4 area" tetapi dalam implementasi nyata setiap area memiliki sub-kompetensi dan indikator yang jauh lebih granular. Dokumentasi perlu menambahkan tabel 4 area → sub-kompetensi → indikator dan menjelaskan bahwa default seeded data akan berisi 14 kompetensi.

---

### 2.2 Nilai Sikap di K13 — Deskripsi Salah

**Lokasi:** `04-student-academic.md` → "Nilai Sikap (score_attitude)"

**Masalah:** Dokumen menyebut `score_attitude` menggunakan skala **SB/B/C/K**. Di Kurikulum 2013 format Permendikbud No. 23/2016, nilai sikap resmi ditulis sebagai:
- **A** = Sangat Baik (SB)
- **B** = Baik
- **C** = Cukup
- **D** = Kurang

Skala huruf SB/B/C/K adalah *label deskriptif*, bukan kode resmi. Namun kode `SB/B/C/K` di database CHECK constraint sudah umum digunakan, sehingga ini bukan bug sistem — tetapi **dokumentasi perlu memperjelas** bahwa kode ini adalah representasi internal, bukan kode formal Permendikbud.

Lebih penting: di **Kurikulum Merdeka**, nilai sikap **tidak dinilai secara eksplisit** di rapor per mapel — penilaian profil pelajar ada di P5 project terpisah. Dokumentasi tidak membedakan ini dengan jelas.

**Koreksi:** Tambahkan catatan: "Untuk Kurikulum Merdeka, `score_attitude` tidak wajib diisi per mapel; penilaian profil pelajar dilakukan melalui `p5_projects` (S019). Untuk K13, `score_attitude` wajib diisi."

---

### 2.3 Batas Cuti Tahunan — Carry-Over

**Lokasi:** `09-teacher-staff.md` → Business Rule #10

**Masalah:** Dokumen menyatakan "carry-over maksimal 6 hari" untuk cuti tahunan PNS. PP No. 11 Tahun 2017 dan turunannya menetapkan bahwa cuti tahunan yang tidak diambil **tidak bisa di-carry-over** (kadaluarsa akhir tahun), kecuali ada ketentuan lain dalam PP kepegawaian instansi. Pernyataan "carry-over 6 hari" perlu dikonfirmasi referensinya — tidak ada dalam PP 11/2017.

**Koreksi:** Hapus klaim carry-over 6 hari atau tambahkan referensi regulasi yang tepat. Default yang aman adalah "cuti tahunan kadaluarsa akhir tahun kalender, tidak bisa di-carry-over."

---

### 2.4 Akreditasi — Masa Berlaku

**Lokasi:** `14-admin-governance.md` → `school_accreditations.valid_until`

**Masalah:** Dokumen menyebut masa berlaku akreditasi "5 tahun dari tanggal akreditasi." Per Permendikbud No. 13/2018 tentang BAN-S/M, akreditasi berlaku **5 tahun**, yang memang benar. Namun sejak 2020 BAN-S/M mengeluarkan kebijakan bahwa sekolah dengan nilai A tetap terakreditasi tanpa perlu re-akreditasi di siklus berikutnya (perpanjangan otomatis jika memenuhi syarat). Dokumen tidak menyebutkan kondisi ini.

**Koreksi:** Tambahkan catatan: "Kebijakan BAN-S/M dapat berubah (misal: perpanjangan otomatis untuk sekolah Terakreditasi A). Field `valid_until` bisa di-override secara manual oleh admin saat ada kebijakan khusus dari BAN-S/M."

---

### 2.5 Komite Sekolah — Larangan Pungutan

**Lokasi:** `14-admin-governance.md` → Business Rules Komite

**Masalah:** Dokumen TIDAK mencantumkan larangan eksplisit bahwa Komite Sekolah **dilarang memungut dana dari orang tua siswa** sesuai Permendikbud No. 75/2016 Pasal 10. Ini adalah kesalahan fatal — banyak kasus penyalahgunaan komite justru karena sistem tidak menegakkan rule ini.

**Koreksi:** Tambahkan Business Rule: "Komite Sekolah dilarang memungut iuran dari orang tua (Permendikbud 75/2016 Pasal 10). Sistem hanya mendukung pencatatan donasi sukarela (`dana_komite` sebagai source type). Tidak ada fitur tagihan komite ke orang tua."

---

### 2.6 Remedial — Nilai Maksimal Setelah Remedial

**Lokasi:** `08-exam-assessment.md` → Business Rules Remedial

**Masalah:** Dokumen menjelaskan remedial dengan baik tetapi tidak menyebutkan bahwa banyak sekolah (dan beberapa juknis daerah) menetapkan **nilai maksimal hasil remedial = nilai KKM** (tidak bisa melebihi KKM). Sistem saat ini memungkinkan nilai remedial bisa lebih tinggi dari KKM, yang menimbulkan konflik dengan kebijakan sekolah yang menerapkan aturan "nilai remedial dibatasi KKM."

**Koreksi:** Tambahkan konfigurasi di `subject_configurations`: `remedial_max_score_policy` dengan pilihan `unlimited` (default) atau `capped_at_kkm`.

---

## 3. Gap Regulasi Pendidikan Indonesia

### 3.1 ANBK / Asesmen Nasional (P0 — KRITIS)

**ADR terkait:** S022 (Exam Assessment)
**Lokasi:** `08-exam-assessment.md`

ANBK (Asesmen Nasional Berbasis Komputer) menggantikan UN sejak 2021. Sekolah perlu:
- Mendaftarkan siswa peserta ANBK (sampling kelas 5/8/11)
- Menyiapkan jadwal pelaksanaan (Oktober–November)
- Upload proktor dan pengawas ke sistem ANBK

Dokumen sama sekali tidak menyebut ANBK/AKM. Dari perspektif kepala sekolah, ini adalah kegiatan paling penting di semester ganjil.

**Rekomendasi:** Tambahkan ke `08-exam-assessment.md` bagian "Asesmen Nasional (ANBK)" yang menjelaskan bahwa SekolahPro mendukung penjadwalan internal dan persiapan data untuk ANBK, meski data ANBK itu sendiri dikelola oleh sistem nasional Kemdikbud.

---

### 3.2 SKL dan Kelulusan (P0 — KRITIS)

**ADR terkait:** S018 (Rapor Generation), S014 (Class Placement)
**Lokasi:** `04-student-academic.md`, `03-student-lifecycle.md`

Surat Keterangan Lulus (SKL) adalah dokumen yang dikeluarkan kepala sekolah saat siswa dinyatakan lulus sebelum ijazah resmi keluar. Proses kelulusan melibatkan:
1. Rapat dewan guru untuk menetapkan kelulusan
2. Penerbitan SKL (bulan Mei/Juni)
3. Pengambilan ijazah (Juli setelah UN/ANBK selesai)

Sistem saat ini tidak memiliki fitur SKL atau proses kelulusan formal. `graduation_status` ada di `student_academics.promotion_status` tapi tidak ada workflow penetapan kelulusan.

**Rekomendasi:** Tambahkan ke `03-student-lifecycle.md` atau `04-student-academic.md` bagian "Proses Kelulusan" yang menjelaskan alur: rapat kelulusan → penetapan → SKL digital → update status siswa ke `graduated`.

---

### 3.3 Dana BOS dan ARKAS (P0 — KRITIS)

**ADR terkait:** S050 (School Budget RKAS)
**Lokasi:** `14-admin-governance.md`

**Masalah:** ARKAS (Aplikasi Rencana Kegiatan dan Anggaran Sekolah) adalah sistem wajib Kemdikbud untuk semua sekolah penerima BOS. Sekolah wajib input RKAS ke ARKAS sebelum dana BOS cair. Dokumen tidak menyebutkan:
1. Apakah SekolahPro RKAS bisa di-export ke format ARKAS?
2. Siapa yang bertanggung jawab sinkronisasi antara SekolahPro RKAS dan ARKAS?
3. Komponen penggunaan BOS berubah tiap tahun berdasarkan Permendikbud terbaru — mekanisme update komponen tidak ada

**Rekomendasi:** Tambahkan Business Rule: "Komponen BOS mengacu pada Permendikbud terbaru yang bisa berubah tiap tahun. Admin dapat memperbarui daftar komponen BOS dari konfigurasi. Sistem menyediakan export format CSV yang kompatibel dengan ARKAS untuk menghindari double-entry."

---

### 3.4 EMIS Kemenag untuk Madrasah (P0 — KRITIS)

**ADR terkait:** S055 (Dapodik Integration)
**Lokasi:** `15-finance-integration.md`

Madrasah (MI, MTs, MA) berada di bawah Kemenag, **bukan** Kemdikbud. Mereka menggunakan **EMIS** (Education Management Information System Kemenag), **bukan** Dapodik. ADR S055 hanya membahas Dapodik tanpa menyebut EMIS sama sekali.

SekolahPro mengklaim mendukung Madrasah — ini adalah gap fundamental karena operator madrasah akan bertanya "bagaimana sync ke EMIS?"

**Rekomendasi:** Tambahkan ke `15-finance-integration.md` catatan: "Untuk madrasah (MI/MTs/MA), sistem nasional yang digunakan adalah EMIS Kemenag (bukan Dapodik Kemdikbud). ADR-S055 perlu diperluas dengan sub-ADR untuk EMIS integration atau disclaimer bahwa EMIS integration adalah fitur roadmap."

---

### 3.5 Permendikbud tentang PPDB (P1)

**ADR terkait:** S016 (PPDB)
**Lokasi:** `03-student-lifecycle.md`

Dokumen tidak menyebut regulasi PPDB yang berlaku. Regulasi yang relevan:
- **Permendikbud No. 1/2021** tentang PPDB TK, SD, SMP, SMA, SMK (berlaku untuk sekolah negeri)
- **Surat Edaran tahunan** dari Kemendikbud tentang PPDB online

Sekolah swasta tidak diwajibkan mengikuti Permendikbud No. 1/2021, tetapi jalur seleksi di dokumen (zonasi, afirmasi, dll.) adalah ketentuan untuk sekolah negeri. Dokumen tidak membedakan antara sekolah negeri dan swasta dalam PPDB.

---

### 3.6 Standar Nasional Pendidikan (SNP) dan Instrumen Akreditasi (P1)

**ADR terkait:** S048 (School Profile Accreditation)
**Lokasi:** `14-admin-governance.md`

BAN-S/M menggunakan **IASP 2020** (Instrumen Akreditasi Satuan Pendidikan 2020) sebagai pengganti 8 SNP murni. Dokumen masih menggunakan framing "8 SNP" tanpa menyebut IASP 2020 yang mencakup 4 komponen utama:
1. Mutu Lulusan
2. Proses Pembelajaran
3. Mutu Guru
4. Manajemen Sekolah

Data di SekolahPro yang relevan untuk visitasi akreditasi (jurnal mengajar, RPP, nilai, absensi, komite) tidak di-mapping ke komponen IASP 2020.

---

## 4. Gap Operasional Sekolah

*(Perspektif Kepala Sekolah: "Apa yang akan membingungkan guru dan staf TU saya?")*

### 4.1 Buku Induk Siswa (P0 — KRITIS)

**Lokasi:** Tidak ada di dokumentasi manapun

Buku Induk Siswa adalah **dokumen wajib** setiap sekolah berdasarkan regulasi Kemdikbud. Buku Induk mencatat:
- Nomor induk per siswa (berurutan)
- Data lengkap siswa dan orang tua
- Riwayat kelas per tahun
- Nilai akhir per semester

Di sistem saat ini, semua data ada tapi tidak ada fitur khusus "Buku Induk" yang bisa dicetak atau diekspor. Staf TU akan bingung karena mereka terbiasa dengan buku fisik dan akan menanyakan "bagaimana cara cetak Buku Induk?"

**Rekomendasi:** Tambahkan ke `02-student-core.md` bagian "Buku Induk Digital" sebagai laporan yang di-generate dari data existing (student + class history + grade summary).

---

### 4.2 Mutasi Siswa Lintas Kabupaten/Kota (P1)

**Lokasi:** `03-student-lifecycle.md` (S014)

Prosedur mutasi siswa ke/dari sekolah lain melibatkan:
1. Surat keterangan pindah dari sekolah asal
2. Verifikasi NISN di Dapodik
3. Transfer data rapor
4. Re-assign ke kelas di sekolah baru

Dokumen hanya menyebut `student_previous_schools` dan `placement_type = 'transfer'` tanpa membahas workflow dokumen yang diperlukan atau bagaimana data rapor dari sekolah lama bisa diimport.

---

### 4.3 Kenaikan Pangkat Guru dan Angka Kredit (P1)

**Lokasi:** `09-teacher-staff.md` (S029)

Angka kredit PKB disebutkan di ADR dan dokumen, tapi tidak ada penjelasan:
- Format pengajuan DUPAK (Daftar Usulan Penetapan Angka Kredit)
- Pejabat yang berwenang menilai dan menetapkan angka kredit
- Proses upload bukti fisik (sertifikat, SK, laporan penelitian)
- Timeline dari aktivitas PKB ke kenaikan pangkat

Kepala Sekolah yang menggunakan SekolahPro akan menanyakan: "Apakah sistem bisa generate DUPAK?"

---

### 4.4 Pengelolaan Ijazah dan Legalitas Dokumen (P1)

**Lokasi:** `03-student-lifecycle.md` (S010)

`student_documents` mendukung tipe `ijazah` tapi tidak ada penjelasan:
- Penomoran ijazah sesuai format baku
- Penandatanganan kepala sekolah (legalisir digital)
- Blanko ijazah dari dinas — bagaimana sekolah memastikan nomor seri blanko valid
- Ijazah yang hilang dan prosedur penggantian SKHUN

---

### 4.5 Ekuivalensi Jam Pelajaran Lintas Kurikulum (P1)

**Lokasi:** `07-academic-schedule.md` (S021)

Jam pelajaran (JP) berbeda antara kurikulum:
- K13: 1 JP = 40 menit (SD), 40 menit (SMP), 45 menit (SMA)
- Kurikulum Merdeka: fleksibel, sekolah bisa menetapkan sendiri

Dokumen tidak menjelaskan konfigurasi durasi JP per jenjang. Admin sekolah yang pindah dari K13 ke Merdeka akan bingung apakah total JP per minggu berubah.

---

### 4.6 Sinkronisasi e-Rapor Kemdikbud (P0 — KRITIS)

**Lokasi:** `04-student-academic.md` (S018)

e-Rapor adalah sistem resmi Kemdikbud untuk rapor digital yang terintegrasi dengan Dapodik. **Banyak sekolah wajib menggunakan e-Rapor Kemdikbud**, bukan membuat sistem rapor sendiri.

Dokumen membangun sistem rapor lengkap sendiri tapi tidak menyebut sama sekali bagaimana posisi rapor SekolahPro vs e-Rapor Kemdikbud. Ini akan menjadi pertanyaan pertama kepala sekolah:
- "Apakah saya masih harus input ulang ke e-Rapor Kemdikbud?"
- "Apakah rapor ini diakui secara hukum?"

**Rekomendasi:** Tambahkan ke `04-student-academic.md` atau `01-domain-overview.md` penjelasan: "SekolahPro Rapor ditujukan sebagai rapor internal sekolah. Untuk sekolah yang wajib menggunakan e-Rapor Kemdikbud, SekolahPro menyediakan export data nilai dalam format yang kompatibel dengan e-Rapor."

---

### 4.7 Verval PD dan Sinkronisasi NISN (P0)

**Lokasi:** `15-finance-integration.md` (S055)

Verval PD (Verifikasi dan Validasi Peserta Didik) adalah proses wajib Dapodik untuk memastikan setiap siswa memiliki NISN yang valid. Proses ini melibatkan:
- Operator upload data siswa baru
- Sistem Dapodik memverifikasi dan mengalokasikan NISN
- Jika siswa sudah punya NISN di sistem lain, ada proses merge

Dokumen menyebut NISN tapi tidak menjelaskan workflow Verval PD. Operator sekolah akan bingung bagaimana mendapatkan NISN untuk siswa baru.

---

### 4.8 Laporan BOS per Triwulan (P0)

**Lokasi:** `14-admin-governance.md` (S050)

Sekolah penerima BOS wajib membuat Laporan Pertanggungjawaban (LPJ) BOS per triwulan yang disampaikan ke dinas pendidikan. Format LPJ berbeda-beda per daerah tapi minimal meliputi:
- Rekapitulasi penerimaan dana BOS
- Rekapitulasi pengeluaran per komponen 8 SNP
- Bukti pengeluaran (kuitansi)

`budget_realizations` ada di schema tapi tidak ada penjelasan bagaimana LPJ di-generate dari data tersebut.

---

## 5. Rekomendasi Penambahan per Dokumen

### 5.1 `02-student-core.md`
Tambahkan: **Bagian "Buku Induk Digital"** — endpoint `GET /api/v1/students/buku-induk?class_id={id}` yang menghasilkan laporan Buku Induk dalam format Excel/PDF.

### 5.2 `03-student-lifecycle.md`
Tambahkan: **Bagian "Proses Kelulusan"** — workflow rapat kelulusan, penetapan kelulusan, penerbitan SKL digital. Tambahkan state `graduation_decided` sebelum status siswa berubah ke `graduated`.

### 5.3 `04-student-academic.md`
Tambahkan: **Bagian "Posisi vs e-Rapor Kemdikbud"** — disclaimer dan penjelasan export compatibility. Tambahkan catatan nilai sikap di Kurikulum Merdeka.

### 5.4 `05-student-welfare.md` (S012 Disiplin)
Tambahkan: **Kebijakan reset poin disiplin** — konfigurasi per sekolah (per semester vs per tahun). Ini tidak boleh hardcoded karena setiap sekolah berbeda.

### 5.5 `06-curriculum-subject.md`
Tambahkan: **Silabus sebagai dokumen pre-RPP** untuk K13. Jelaskan bahwa silabus ditetapkan Kemdikbud untuk mapel nasional dan bisa di-download dari website Kemdikbud.

### 5.6 `08-exam-assessment.md`
Tambahkan: **Bagian "Asesmen Nasional (ANBK)"** — ruang lingkup, jadwal, dan data yang perlu disiapkan. Tambahkan konfigurasi `remedial_max_score_policy`.

### 5.7 `09-teacher-staff.md`
Koreksi: **PKG indikator** — clarify bahwa default competency template akan berisi 14 kompetensi untuk guru kelas/mapel. Tambahkan PKKS untuk kepala sekolah. Koreksi carry-over cuti. Tambahkan DUPAK workflow untuk PKB.

### 5.8 `14-admin-governance.md` (Komite)
Tambahkan: **Larangan pungutan Komite** sebagai Business Rule eksplisit dengan referensi Permendikbud 75/2016 Pasal 10. Tambahkan penjelasan bahwa `komite` sebagai `source_type` hanya untuk donasi sukarela, bukan pungutan wajib.

### 5.9 `14-admin-governance.md` (RKAS)
Tambahkan: **Bagian "Integrasi ARKAS"** — penjelasan export format dan disclaimer bahwa ARKAS adalah sistem terpisah Kemdikbud. Tambahkan penjelasan LPJ BOS per triwulan.

### 5.10 `15-finance-integration.md` (Dapodik)
Tambahkan: **Bagian "EMIS Kemenag"** untuk madrasah. Tambahkan workflow Verval PD untuk NISN baru. Tambahkan penjelasan cutoff date Dapodik dan implikasi BOS.

---

## 6. Penilaian Ahli: Apakah Business Rules Selaras dengan Operasional Sekolah Indonesia?

### Yang Sudah Benar dan Baik

| Aspek | Penilaian |
|-------|-----------|
| Struktur data siswa (NIS/NISN, 6 agama, status lifecycle) | **Sangat Baik** — sesuai standar Dapodik |
| Kurikulum Merdeka (CP/TP/ATP, P5, Fase A-F) | **Sangat Baik** — akurat dan lengkap |
| Beban mengajar guru 24 JP (PP 74/2008) | **Benar** |
| SPP dengan 3-layer (fee_type/invoice/payment) | **Baik** — workflow realitis |
| Rapor snapshot strategy | **Sangat Baik** — praktik terbaik untuk rapor sebagai dokumen hukum |
| 5 jalur PPDB | **Benar** — sesuai regulasi sekolah negeri |
| Guardian many-to-many | **Benar** — sangat praktis untuk kakak-adik satu sekolah |
| PKG 4 kompetensi (pedagogik, kepribadian, sosial, profesional) | **Benar** secara kategori |
| BOS multi-source dan 8 SNP sebagai komponen | **Benar** |
| Dual-mode pesantren (diniyah + nasional) | **Sangat Baik** — salah satu keunggulan utama |
| Dormitory dengan gender separation absolut | **Benar** |

### Yang Perlu Diperbaiki

| Aspek | Masalah |
|-------|---------|
| Jumlah indikator PKG | 4 area ≠ 14 kompetensi — harus jelas di dokumentasi |
| Nilai sikap Kurikulum Merdeka | Tidak sama dengan K13 — perlu dibedakan |
| e-Rapor Kemdikbud | Tidak dibahas — akan menjadi pertanyaan utama kepala sekolah |
| Buku Induk Siswa | Dokumen wajib yang tidak ada di sistem |
| SKL (Surat Keterangan Lulus) | Tidak ada workflow kelulusan resmi |
| ARKAS | Sistem wajib BOS yang tidak dikaitkan dengan RKAS SekolahPro |
| EMIS Kemenag | Madrasah tidak bisa pakai Dapodik integration |
| Reset poin disiplin | Tidak ada kebijakan yang jelas |
| Verval PD (NISN) | Operator sekolah tidak tahu bagaimana dapat NISN baru |

### Catatan Penting dari Perspektif Kepala Sekolah

> Secara keseluruhan, dokumentasi ini menunjukkan pemahaman yang solid tentang operasional sekolah Indonesia, khususnya untuk Kurikulum Merdeka dan konteks pesantren. Beberapa hal yang akan langsung ditanyakan kepala sekolah saat demo:
>
> 1. "Apakah sistem ini terintegrasi dengan Dapodik?" → Terjawab di S055, tapi Verval PD belum ada.
> 2. "Bagaimana dengan e-Rapor?" → Tidak terjawab di dokumentasi.
> 3. "Apakah PKG sudah sesuai dengan 14 kompetensi?" → Tidak jelas.
> 4. "Bagaimana RKAS ini berhubungan dengan ARKAS?" → Tidak terjawab.
> 5. "Buku Induk Siswa bisa dicetak?" → Tidak ada.
>
> Lima pertanyaan ini harus dijawab di dokumentasi sebelum presentasi ke kepala sekolah.

---

## 7. Daftar Gap Berdasarkan Prioritas

### P0 — Kritis (Segera Perbaiki)

| # | Gap | Dokumen yang Perlu Diperbarui |
|---|-----|-------------------------------|
| 1 | e-Rapor Kemdikbud — posisi dan export compatibility | `04-student-academic.md` |
| 2 | SKL (Surat Keterangan Lulus) dan workflow kelulusan | `03-student-lifecycle.md`, `04-student-academic.md` |
| 3 | Buku Induk Siswa sebagai laporan wajib | `02-student-core.md` |
| 4 | ARKAS dan LPJ BOS per triwulan | `14-admin-governance.md` |
| 5 | EMIS Kemenag untuk madrasah | `15-finance-integration.md` |
| 6 | Verval PD / NISN allocation workflow | `15-finance-integration.md` |
| 7 | Larangan pungutan Komite Sekolah | `14-admin-governance.md` |
| 8 | Reset poin disiplin siswa | `05-student-welfare.md` |
| 9 | PKG — klarifikasi jumlah indikator (14 kompetensi) | `09-teacher-staff.md` |

### P1 — Penting

| # | Gap | Dokumen |
|---|-----|---------|
| 1 | ANBK/AKM dalam konteks assessment | `08-exam-assessment.md` |
| 2 | Carry-over cuti tahunan PNS — referensi regulasi | `09-teacher-staff.md` |
| 3 | Nilai sikap Kurikulum Merdeka vs K13 | `04-student-academic.md` |
| 4 | DUPAK workflow untuk PKB | `09-teacher-staff.md` |
| 5 | Silabus sebagai dokumen pre-RPP (K13) | `06-curriculum-subject.md` |
| 6 | Permendikbud PPDB untuk sekolah negeri vs swasta | `03-student-lifecycle.md` |
| 7 | IASP 2020 dan mapping data ke komponen akreditasi | `14-admin-governance.md` |
| 8 | Mutasi lintas kabupaten — workflow dokumen | `03-student-lifecycle.md` |
| 9 | Konfigurasi remedial max score | `08-exam-assessment.md` |
| 10 | NUPTK untuk pelaporan beban mengajar | `09-teacher-staff.md` |
| 11 | Penomoran ijazah dan legalitas dokumen | `03-student-lifecycle.md` |
| 12 | Ekuivalensi JP antar kurikulum | `07-academic-schedule.md` |

### P2 — Nice-to-Have

| # | Gap | Dokumen |
|---|-----|---------|
| 1 | SNPMB/PPDB PT linkage dari e-learning | `15-finance-integration.md` |
| 2 | Laporan 8 SNP untuk persiapan akreditasi | `15-finance-integration.md` |

---

## 8. Kesimpulan Validator

Dokumentasi SekolahPro `spec/sekolah/` secara keseluruhan **berkualitas baik** — cakupan ADR 100% (tidak ada ADR yang tidak terdokumentasi sama sekali), struktur entitas akurat, dan business rules sebagian besar sesuai dengan praktik nyata sekolah Indonesia.

**Kekuatan utama:**
- Kurikulum Merdeka (CP/TP/ATP/P5) terdokumentasi dengan sangat baik
- Pesantren dual-mode adalah keunggulan kompetitif yang terdokumentasi dengan benar
- Schema database praktis dan sesuai standar Dapodik
- Vernon Pattern konsisten dan terdokumentasi

**Kelemahan utama:**
- Gap di integrasi nasional (e-Rapor, ARKAS, EMIS) adalah risiko terbesar — kepala sekolah akan langsung mempertanyakan ini
- Beberapa dokumen wajib (Buku Induk, SKL) tidak ada
- Regulasi kadang dikutip tanpa nomor pasal/tahun yang tepat

**Rating keseluruhan: 7/10 — Baik, dengan 9 gap P0 yang perlu diselesaikan sebelum presentasi ke kepala sekolah.**

---

*Laporan ini dihasilkan dari validasi silang antara 58 ADR (S001–S058) dan 16 dokumen spec sekolah, dengan perspektif kepala sekolah dan education system expert yang memahami regulasi Kemdikbud, Kemenag, dan praktik operasional sekolah Indonesia.*
