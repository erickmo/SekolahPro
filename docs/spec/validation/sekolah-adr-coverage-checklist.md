# Sekolah ADR Coverage Checklist
**Tanggal Validasi:** 2026-04-26
**Validator:** Kepala Sekolah & Education System Expert
**Total ADR:** 58 (S001–S058)

---

## Legend

- **Coverage Level:** FULL = semua keputusan utama ADR tercakup; PARTIAL = sebagian tercakup dengan gap; MISSING = tidak ada representasi sama sekali
- **Priority Gap:** P0 = kritis (perlu segera); P1 = penting (perlu dilengkapi); P2 = nice-to-have

---

## Student Domain (S001–S018)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S001 | Student Core Data Model | 02-student-core.md | FULL | — | — |
| S002 | Student Relationships & Autoload Strategy | 02-student-core.md | FULL | Autoload strategy lengkap, Vernon descriptor terdokumentasi | — |
| S003 | Student Guardian (Orang Tua/Wali) | 02-student-core.md | FULL | — | — |
| S004 | Student Academic Record | 03-student-lifecycle.md, 04-student-academic.md | FULL | — | — |
| S005 | Student Health Record | 03-student-lifecycle.md | FULL | — | — |
| S006 | Student Dashboard Menu Structure | 02-student-core.md | FULL | — | — |
| S007 | Student Address & Previous School | 02-student-core.md | FULL | — | — |
| S008 | Daily Attendance Transaction | 03-student-lifecycle.md | FULL | — | — |
| S009 | Student Finance / SPP | 05-student-welfare.md | PARTIAL | (1) Tidak ada penjelasan mekanisme waive (keringanan) untuk anak yatim/kurang mampu sesuai juknis BOS; (2) Integrasi Payment Gateway (S051) untuk pembayaran online SPP tidak dijelaskan; (3) Tidak ada mention laporan keuangan SPP untuk BOS accountability | P1 |
| S010 | Student Document Management | 03-student-lifecycle.md | FULL | — | — |
| S011 | Subject Grade Detail | 04-student-academic.md | FULL | — | — |
| S012 | Student Discipline / Tata Tertib | 05-student-welfare.md | PARTIAL | (1) ADR menyebut skorsing `1 minggu` dalam field `action_taken`, namun tidak ada penjelasan apakah skorsing di atas 3 hari memerlukan Surat Keputusan Kepala Sekolah sesuai prosedur (aturan disiplin). (2) Kebijakan reset poin (per semester vs per tahun) disebutkan "belum di-define" di ADR — ini gap kritis untuk konsistensi rapor | P0 |
| S013 | Student Achievement / Prestasi | 05-student-welfare.md | FULL | — | — |
| S014 | Student Class Placement / Mutasi Kelas | 03-student-lifecycle.md | PARTIAL | (1) Tidak ada penjelasan mekanisme `double promotion` (siswa lompat kelas); (2) Tidak ada penjelasan prosedur mutasi keluar kota (diperlukan `surat keterangan pindah` dari sekolah asal dan NISN matching) | P1 |
| S015 | Student Extracurricular | 05-student-welfare.md | FULL | — | — |
| S016 | PPDB / Student Admission | 03-student-lifecycle.md | PARTIAL | (1) Tidak menyebut Permendikbud tentang PPDB yang berlaku (Permendikbud No. 1/2021 atau revisinya); (2) Mekanisme jalur afirmasi (KIP, DTKS, disabilitas) tidak dirinci; (3) Pengumuman via website publik belum dibahas; (4) Tidak ada penjelasan pengelolaan `daftar tunggu` (waitlist) saat ada yang mundur | P1 |
| S017 | BK / Student Counseling Record | 05-student-welfare.md | FULL | — | — |
| S018 | Rapor Generation | 04-student-academic.md | PARTIAL | (1) Tidak ada penjelasan `e-Rapor` Kemendikbud — sistem nasional yang berbeda dari rapor cetak; (2) Format rapor K13 berbeda signifikan (ada kolom KI-1, KI-2) vs Merdeka — perbedaan template ini tidak dibahas; (3) Tidak ada penjelasan SKL (Surat Keterangan Lulus) sebagai dokumen terpisah dari rapor akhir; (4) Rapor untuk kelas akhir (kelas 9/12) seharusnya memuat informasi lulus/tidak — link ke graduation/kelulusan tidak ada | P0 |

---

## Academic Domain (S019–S025)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S019 | Curriculum Management | 06-curriculum-subject.md | FULL | — | — |
| S020 | Subject Management | 06-curriculum-subject.md | FULL | — | — |
| S021 | Teaching Schedule / Timetable | 07-academic-schedule.md | FULL | — | — |
| S022 | Exam & Assessment Management | 08-exam-assessment.md | PARTIAL | (1) ANBK (Asesmen Nasional Berbasis Komputer) sebagai pengganti UN tidak dibahas — sekolah perlu scheduling dan persiapan khusus; (2) AKM (Asesmen Kompetensi Minimum) dan Survei Karakter tidak masuk scope — perlu penjelasan bahwa data ini tidak dikelola sistem; (3) Tidak ada penjelasan untuk kelas akhir: bagaimana penentuan kelulusan selain dari nilai akhir | P1 |
| S023 | Academic Calendar | 07-academic-schedule.md | FULL | — | — |
| S024 | Lesson Plan / RPP | 07-academic-schedule.md | PARTIAL | (1) Silabus sebagai dokumen yang mendahului RPP di K13 tidak dibahas; (2) Tidak ada mention kewajiban RPP sebagai komponen Standar Proses SNP (Permendikbud No. 22/2016 untuk K13 masih berlaku di banyak sekolah) | P1 |
| S025 | Teaching Journal | 07-academic-schedule.md | FULL | — | — |

---

## Teacher Domain (S026–S032)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S026 | Teacher Attendance | 09-teacher-staff.md | FULL | — | — |
| S027 | Teacher Workload | 09-teacher-staff.md | PARTIAL | (1) Tidak ada penjelasan NUPTK (Nomor Unik Pendidik dan Tenaga Kependidikan) sebagai identifier wajib untuk pelaporan beban mengajar ke Dapodik/BKN; (2) Tidak ada penjelasan ekuivalen JP untuk kepala perpustakaan, laboran, dan konselor (BK) yang berbeda dari guru kelas; (3) Guru non-PNS (honorer/yayasan) — aturan minimal JP tidak sama dengan PNS, tidak dibahas | P1 |
| S028 | Teacher Performance Evaluation (PKG) | 09-teacher-staff.md | PARTIAL | (1) Jumlah indikator PKG: ADR menyebut "4 kompetensi" tapi standar Permenneg PAN & RB 16/2009 memiliki 14 kompetensi dengan 78 indikator untuk guru kelas dan 45 indikator untuk guru mata pelajaran — perbedaan ini kritis; (2) PKG untuk Kepala Sekolah menggunakan instrumen berbeda (PKKS — Penilaian Kinerja Kepala Sekolah) yang tidak dibahas; (3) Hubungan PKG dengan kenaikan pangkat PNS (angka kredit) tidak dijelaskan; (4) Dokumen PKG sebagai bukti akreditasi belum disebutkan | P0 |
| S029 | Professional Development (PKB) | 09-teacher-staff.md | PARTIAL | (1) Struktur 3 komponen PKB (Pengembangan Diri / Publikasi Ilmiah / Karya Inovatif) dari Permenneg PAN & RB 16/2009 ada di ADR tapi tidak dijelaskan angka kredit minimum tiap komponen; (2) Sertifikasi guru melalui PLPG/PPG tidak dibahas; (3) Tidak ada penjelasan e-PKG / e-PKB Dapodik integration | P1 |
| S030 | Leave Management | 09-teacher-staff.md | FULL | — | — |
| S031 | Staff Payroll | 09-teacher-staff.md | PARTIAL | (1) Gaji pokok PNS "di luar scope" tapi tidak ada penjelasan bagaimana slip gaji tetap bisa di-generate untuk rekap (gaji pokok perlu masuk slip meski transfer dari APBN); (2) Tunjangan fungsional (Tunjangan Jabatan Fungsional) berbeda dari TPG — tidak dibahas; (3) Tidak ada penjelasan kewajiban laporan pajak PPh 21 SPT tahunan yang dihasilkan sistem | P1 |
| S032 | Teacher Substitution | 09-teacher-staff.md | FULL | — | — |

---

## Dormitory Domain (S033–S035)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S033 | Dormitory Management | 10-dormitory.md | FULL | — | — |
| S034 | Dormitory Activity & Attendance | 10-dormitory.md | FULL | — | — |
| S035 | Dormitory Discipline & Health | 10-dormitory.md | PARTIAL | (1) UKS (Unit Kesehatan Sekolah) terintegrasi dengan asrama di pesantren — tidak ada penjelasan bagaimana data kesehatan asrama berinteraksi dengan data UKS akademik (S005); (2) Tidak ada penjelasan prosedur rujukan ke rumah sakit dan siapa yang berwenang | P1 |

---

## Canteen Domain (S036–S037)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S036 | Canteen Management | 11-canteen.md | FULL | — | — |
| S037 | Canteen Transaction & Billing | 11-canteen.md | PARTIAL | (1) Bagaimana integrasi tagihan makan ke SPP (satu tagihan gabungan) tidak dijelaskan secara eksplisit meski ADR menyebutnya; (2) Aturan khusus untuk santri yang tidak mampu (beasiswa makan) tidak ada | P1 |

---

## Facility Domain (S038–S041)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S038 | Library Management | 12-facility.md | FULL | — | — |
| S039 | Laboratory Management | 12-facility.md | FULL | — | — |
| S040 | Asset & Inventory Management | 12-facility.md | FULL | — | — |
| S041 | Room & Facility Booking | 12-facility.md | PARTIAL | (1) Tidak ada penjelasan bagaimana booking ruangan yang disetujui ter-propagasi ke jadwal pelajaran (mencegah konflik guru mengajar di ruang yang sudah dibooking); (2) Tidak ada penjelasan untuk booking yang dibatalkan akibat kondisi darurat (bencana, dll.) | P1 |

---

## Communication Domain (S042–S046)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S042 | Parent Portal & Dashboard | 13-communication.md | FULL | — | — |
| S043 | Communication & Messaging | 13-communication.md | FULL | — | — |
| S044 | Notification System | 13-communication.md | FULL | — | — |
| S045 | Announcement & News | 13-communication.md | FULL | — | — |
| S046 | Correspondence Management | 13-communication.md | PARTIAL | (1) Format nomor surat keluar (surat dinas) mengikuti standar Permenpan — tidak ada penjelasan bahwa sekolah negeri wajib menggunakan format Permenpan No. 55/2011 atau SE terbaru; (2) Surat masuk dari Dinas Pendidikan memerlukan disposisi ke bidang yang relevan — tidak dibahas apakah ada tracking follow-up deadline | P1 |

---

## Admin Domain (S047–S050)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S047 | Approval Workflow Engine | 14-admin-governance.md | FULL | — | — |
| S048 | School Profile & Accreditation | 14-admin-governance.md | PARTIAL | (1) Akreditasi BAN-S/M menggunakan 8 Standar Nasional Pendidikan (SNP) sebagai instrument visitasi — tidak ada penjelasan bagaimana data di sistem di-mapping ke instrumen akreditasi; (2) Perpanjangan akreditasi otomatis (kebijakan BAN-SM saat COVID yang berlanjut) tidak dibahas; (3) EMIS (Education Management Information System) untuk madrasah (Kemenag) berbeda dari Dapodik — tidak ada bedanya untuk madrasah yang pakai SekolahPro | P1 |
| S049 | School Committee (Komite Sekolah) | 14-admin-governance.md | PARTIAL | (1) Larangan Komite memungut dana dari orang tua (Permendikbud 75/2016 Pasal 10) tidak dibahas — ini kritis untuk compliance karena sering jadi masalah; (2) Pemilihan Komite Sekolah (Musyawarah Orang Tua) tidak ada prosesnya dalam dokumen; (3) Laporan kegiatan Komite ke dinas pendidikan tidak dibahas | P0 |
| S050 | School Budget / RKAS | 14-admin-governance.md | PARTIAL | (1) Petunjuk teknis BOS terbaru (Permendikbud No. 2/2022 dan perubahannya) tidak dikutip secara eksplisit — komponen penggunaan BOS bisa berubah tiap tahun; (2) ARKAS (Aplikasi Rencana Kegiatan dan Anggaran Sekolah) adalah sistem nasional Kemdikbud untuk RKAS — tidak ada penjelasan apakah SekolahPro RKAS terintegrasi atau terpisah dari ARKAS; (3) Format laporan pertanggungjawaban (LPJ) BOS tidak dibahas; (4) `quarter` (triwulan) ada di schema tapi mekanisme pelaporan BOS per triwulan ke dinas tidak dijelaskan | P0 |

---

## Finance Domain (S051)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S051 | School Payment Gateway | 15-finance-integration.md | FULL | — | — |

---

## Integration Domain (S052–S055)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S052 | Transportation Management | 15-finance-integration.md | FULL | — | — |
| S053 | E-Learning Integration | 15-finance-integration.md | PARTIAL | (1) PMB (Penerimaan Mahasiswa Baru) linkage untuk siswa kelas XII — e-learning link ke SNPMB/PPDB Perguruan Tinggi tidak dibahas; (2) Verifikasi lisensi Google Workspace for Education yang banyak dipakai sekolah tidak ada | P2 |
| S054 | Reporting & Analytics Dashboard | 15-finance-integration.md | PARTIAL | (1) Laporan 8 SNP (Standar Nasional Pendidikan) sebagai output untuk persiapan akreditasi tidak ada di daftar pre-built reports; (2) Laporan Mutu Sekolah (EDS — Evaluasi Diri Sekolah) tidak dibahas; (3) Dashboard analitik untuk kepala dinas/pengawas (multi-sekolah) tidak ada | P1 |
| S055 | Dapodik Integration | 15-finance-integration.md | PARTIAL | (1) EMIS Kemenag (untuk madrasah) adalah sistem berbeda dari Dapodik — tidak ada penjelasan; (2) Verval PD (Verifikasi dan Validasi Peserta Didik) sebagai proses wajib di Dapodik tidak dibahas; (3) Cutoff date Dapodik per semester (biasanya akhir September dan Maret) dan implikasi terhadap dana BOS tidak ada; (4) NISN allocation workflow (request NISN baru untuk siswa yang belum punya) tidak dibahas | P0 |

---

## Alumni Domain (S056–S058)

| ADR ID | ADR Judul | Tercakup Di Dok | Coverage Level | Gap / Kekurangan | Prioritas |
|--------|-----------|----------------|----------------|------------------|-----------|
| S056 | Alumni Management | 16-alumni.md | FULL | — | — |
| S057 | Scholarship Management | 16-alumni.md | PARTIAL | (1) Beasiswa pemerintah (KIP, LPDP, Bidikmisi) memiliki persyaratan dokumen yang berbeda — tidak ada penjelasan pengelolaan dokumen per jenis beasiswa pemerintah vs swasta; (2) Tidak ada penjelasan verifikasi DTKS (Data Terpadu Kesejahteraan Sosial) untuk penerima KIP — integrasi atau manual? | P1 |
| S058 | School Event Management | 16-alumni.md | FULL | — | — |

---

## Ringkasan Coverage

| Kategori | Total ADR | FULL | PARTIAL | MISSING |
|----------|-----------|------|---------|---------|
| Student (S001–S018) | 18 | 11 | 7 | 0 |
| Academic (S019–S025) | 7 | 4 | 3 | 0 |
| Teacher (S026–S032) | 7 | 3 | 4 | 0 |
| Dormitory (S033–S035) | 3 | 2 | 1 | 0 |
| Canteen (S036–S037) | 2 | 1 | 1 | 0 |
| Facility (S038–S041) | 4 | 3 | 1 | 0 |
| Communication (S042–S046) | 5 | 4 | 1 | 0 |
| Admin (S047–S050) | 4 | 1 | 3 | 0 |
| Finance (S051) | 1 | 1 | 0 | 0 |
| Integration (S052–S055) | 4 | 1 | 3 | 0 |
| Alumni (S056–S058) | 3 | 2 | 1 | 0 |
| **Total** | **58** | **33** | **25** | **0** |

**Coverage Rate:** 33/58 FULL (57%), 25/58 PARTIAL (43%), 0/58 MISSING (0%)

---

*Dokumen ini dihasilkan dari validasi manual ADR S001–S058 terhadap dokumentasi di `/docs/spec/sekolah/`.*
