# Member (Nasabah/Anggota) & Rekening

Modul member mengelola seluruh siklus keanggotaan koperasi dari pendaftaran hingga keluar. Setiap nasabah memiliki satu atau lebih rekening yang menjadi pusat seluruh aktivitas keuangan. Rekening adalah **central nexus** yang menghubungkan nasabah dengan semua produk, transaksi, dan laporan.

---

## ADR References

- **ADR-K001** — Nasabah (Anggota/Member)
- **ADR-K002** — Rekening (Akun Nasabah)

---

## Domain Entities

### Tabel `nasabah`

| Field | Tipe | Keterangan |
|-------|------|------------|
| id | UUID v7 (PK) | |
| tenant_id | UUID (FK) | |
| branch_id | UUID (FK) | Cabang domisili |
| member_number | VARCHAR UNIQUE/tenant | Format: `KOP-YYYY-BRANCH-NNNNNN` (auto-generate) |
| full_name | VARCHAR NOT NULL | |
| nickname | VARCHAR | |
| birth_place / birth_date | VARCHAR / DATE | |
| gender | ENUM (male, female) | |
| religion / marital_status | VARCHAR / ENUM | |
| education_level / occupation | VARCHAR | |
| identity_type | ENUM (ktp, sim, passport, kartu_pelajar) | |
| identity_number | VARCHAR UNIQUE/tenant | |
| identity_expiry | DATE nullable | null = seumur hidup (KTP) |
| identity_photo_url | VARCHAR | |
| photo_url | VARCHAR | Foto wajah terbaru |
| signature_specimen_url | VARCHAR | Tanda tangan untuk verifikasi teller |
| phone / email | VARCHAR | |
| address / province / city / district / village / postal_code | TEXT / VARCHAR | |
| school_relation_type | ENUM | student, teacher, staff, parent, external |
| school_entity_id | UUID nullable | FK → entitas sekolah (siswa/guru/staf) |
| sponsor_id | UUID nullable | FK → nasabah lain sebagai referral |
| kyc_level | ENUM | basic (default), full |
| kyc_verified_at / kyc_verified_by | TIMESTAMPTZ / UUID | Kapan dan siapa yang upgrade KYC |
| status | ENUM | active, inactive, suspended |
| join_date / exit_date / exit_reason | DATE / DATE / TEXT | |
| application_id | UUID (FK) | Sumber pendaftaran |
| template_version | INTEGER | Versi form saat mendaftar |
| custom_fields | JSONB | Data dari custom sections form |
| _rels / _data | JSONB | Vernon pattern |

### Tabel `nasabah_application`

| Field | Tipe | Keterangan |
|-------|------|------------|
| id | UUID v7 (PK) | |
| form_data | JSONB | Snapshot isian form |
| sponsor_id | UUID nullable | |
| template_version | INTEGER | |
| terms_accepted | BOOLEAN NOT NULL | |
| status | ENUM | draft, pending, approved, rejected |
| submitted_at | TIMESTAMPTZ nullable | |
| reviewed_by / reviewed_at | UUID / TIMESTAMPTZ | |
| rejection_reason | TEXT nullable | Wajib jika rejected |
| nasabah_id | UUID nullable | Diisi saat approved |

### Tabel `nasabah_beneficiary` (Ahli Waris)

| Field | Tipe | Keterangan |
|-------|------|------------|
| nasabah_id | UUID (FK) | |
| full_name | VARCHAR NOT NULL | |
| relationship | ENUM | spouse, child, parent, sibling, other |
| identity_number | VARCHAR | KTP ahli waris |
| share_percentage | NUMERIC(5,2) NOT NULL | Total semua ahli waris = 100% |
| is_primary | BOOLEAN | Contact person utama saat klaim |
| claim_status | ENUM | none, claimed, settled |

### Tabel `rekening`

| Field | Tipe | Keterangan |
|-------|------|------------|
| id | UUID v7 (PK) | |
| nasabah_id | UUID (FK) | |
| product_id | UUID (FK) | |
| account_number | VARCHAR UNIQUE/tenant | Format: `TB-YYYY-BRANCH-NNNNNNNN` |
| category | ENUM | simpanan_pokok, simpanan_wajib, tabungan, deposito, pinjaman |
| account_name | VARCHAR | Default: nama nasabah + jenis produk |
| balance | NUMERIC(15,2) | Cache saldo — source of truth adalah SUM transaksi |
| available_balance | NUMERIC(15,2) | = balance - hold_amount |
| hold_amount | NUMERIC(15,2) | Saldo yang di-hold (jaminan, proses kliring) |
| status | ENUM | active, frozen, closed |
| freeze_reason | TEXT nullable | Wajib jika frozen |
| minimum_balance | NUMERIC(15,2) | Override dari produk |
| allow_overdraft | BOOLEAN | Default: false |
| is_dormant | BOOLEAN | Flag dormant (bukan status terpisah) |
| dormant_since | DATE nullable | |
| goal_config | JSONB nullable | Konfigurasi tabungan berencana |
| deposito_config | JSONB nullable | Konfigurasi deposito (tenor, maturity, rollover) |
| last_transaction_at | TIMESTAMPTZ nullable | |
| opened_at | TIMESTAMPTZ NOT NULL | |
| _rels / _data | JSONB | Vernon pattern |

### Tabel `rekening_application`

| Field | Tipe | Keterangan |
|-------|------|------------|
| nasabah_id | UUID (FK) | |
| product_id | UUID (FK) | |
| category | ENUM | tabungan, deposito, pinjaman |
| form_data | JSONB | Snapshot isian form |
| initial_deposit | NUMERIC nullable | Setoran awal yang direncanakan |
| deposit_amount / tenor_months / maturity_date | NUMERIC / INTEGER / DATE | Khusus deposito |
| rollover_instruction | ENUM | Khusus deposito |
| loan_amount / loan_tenor_months / loan_purpose | NUMERIC / INTEGER / TEXT | Khusus pinjaman |
| status | ENUM | draft, pending, approved, rejected |
| rekening_id | UUID nullable | Diisi saat approved |

---

## Business Rules

### Pendaftaran Anggota (K001)

1. Pendaftaran **wajib** melalui formulir aplikasi formal dengan status DRAFT → PENDING → APPROVED/REJECTED
2. Teller bisa buat aplikasi tapi **tidak bisa approve sendiri** — minimal Supervisor untuk approve
3. Rejection **wajib** menyertakan `rejection_reason`
4. Setiap anggota **wajib** mencantumkan minimal 1 ahli waris — aplikasi tidak bisa di-submit tanpa ahli waris
5. `SUM(share_percentage)` semua ahli waris satu nasabah **harus = 100%**
6. Tepat 1 ahli waris harus ber-flag `is_primary = true`
7. Saat nasabah di-approve (status → ACTIVE), sistem **otomatis** membuat rekening SIMPANAN_POKOK dan SIMPANAN_WAJIB
8. Nomor anggota **immutable** setelah assigned
9. Semua nasabah baru dimulai dari KYC level **basic** (limit transaksi bulanan, default Rp 10.000.000)
10. Upgrade KYC ke `full` hanya oleh **Supervisor+** — langsung update + audit log
11. Data nasabah **tidak dihapus** saat deactivation — soft deactivation

### Deactivation

12. Nasabah hanya bisa di-deactivate jika **semua saldo = 0**: tabungan, deposito, simpanan pokok/wajib, sisa pokok pinjaman, tunggakan, denda
13. Response menampilkan daftar rekening/pinjaman yang masih memiliki saldo (blocking items)
14. Deactivation meng-set `status = inactive`, `exit_date = NOW()`, wajib isi `exit_reason`
15. Reactivation dimungkinkan oleh Admin — menghasilkan application baru (proses approval ulang)

### Transfer Antar Cabang

16. Approval hanya dari **Manager di branch asal** — branch tujuan cukup notifikasi
17. Transfer dilakukan dalam **satu database transaction** (atomik) — semua rekening ikut pindah
18. Nasabah tidak bisa di-transfer jika ada rekening dalam status **FROZEN**

### Merge Duplikat

19. Hanya **Admin** yang bisa execute merge
20. **Preview report** wajib ditampilkan sebelum eksekusi
21. Auto-reconciliation: SUM saldo sebelum merge **harus = SUM sesudah merge** — jika tidak match, merge dibatalkan
22. Reversible selama **30 hari** — Admin bisa undo
23. Duplicate di-set `status = merged`, `merged_into_id = primary.id`

### Specimen Tanda Tangan

24. Teller **wajib verifikasi visual** specimen tanda tangan untuk penarikan di atas threshold (default: Rp 1.000.000)
25. Penarikan di atas `manager_threshold` (default: Rp 5.000.000) perlu approval Manager real-time
26. Update specimen membutuhkan minimal role **Supervisor**

### Rekening (K002)

27. Rekening SIMPANAN_POKOK dan SIMPANAN_WAJIB **tidak perlu pengajuan** — dibuat otomatis
28. Tabungan, deposito, pinjaman: harus melalui pengajuan (rekening_application) + approval Supervisor+
29. Pinjaman: approval minimum **Manager+**
30. Saldo rekening adalah **cache** — di-update atomik bersamaan dengan insert transaksi
31. Reconciliation harian via scheduled job membandingkan cached balance vs SUM transaksi
32. Penarikan menggunakan `available_balance`, bukan `balance`
33. Hold dilakukan oleh **Supervisor+** dengan alasan — auto-release saat proses selesai
34. `ACTIVE → FROZEN`: Supervisor+ dengan alasan wajib
35. `FROZEN → ACTIVE`: Manager+
36. `ACTIVE → CLOSED` atau `FROZEN → CLOSED`: hanya jika saldo = 0, oleh Manager+
37. `CLOSED → *`: tidak bisa direaktivasi — buka rekening baru
38. FROZEN membolehkan transaksi **masuk** (setoran, angsuran) tapi memblokir transaksi **keluar**
39. Rekening SIMPANAN_POKOK dan SIMPANAN_WAJIB **tidak bisa ditutup individual** — hanya saat nasabah di-deactivate

### Dormant

40. Dormant adalah **flag** (`is_dormant`, `dormant_since`), bukan status terpisah
41. Rekening yang tidak bertransaksi selama `dormant_period` (default: 12 bulan) di-flag dormant
42. Nasabah bisa mengaktifkan kembali kapan saja dengan melakukan transaksi — flag dihapus
43. Admin fee dormant hanya berlaku jika dikonfigurasi di tenant config

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Application form + approval untuk pendaftaran | Regulasi koperasi mewajibkan persetujuan pengurus; mencegah data tidak valid |
| Ahli waris wajib saat pendaftaran | Kewajiban regulasi koperasi dan OJK untuk LKM |
| Soft deactivation (bukan hard delete) | Data historis diperlukan untuk audit, laporan regulasi, dan SHU |
| Deactivation diblokir jika ada saldo > 0 | Mencegah settlement otomatis yang membutuhkan keputusan manusia |
| Cached balance + reconciliation harian | Performa optimal; query SUM semua transaksi tidak acceptable untuk ribuan transaksi |
| Dormant sebagai flag bukan status | Dormant bersifat informasional dan auto-reversible; rekening dormant masih bisa terima setoran |
| FROZEN membolehkan transaksi masuk | Nasabah tetap bisa memenuhi kewajiban (bayar angsuran) tanpa bisa menarik dana |
| available_balance = balance - hold_amount | Mencegah penarikan dana yang sedang dalam proses |
| Nomor anggota/rekening immutable | Integritas referensi data historis |
| Vernon _data tidak boleh data sensitif | Keamanan — identity_number, saldo, dan data finansial hanya diakses via query detail |

---

## RBAC Summary

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Buat aplikasi anggota | v | v | v | v |
| Approve/Reject aplikasi | - | v | v | v |
| Edit form template | - | - | v | v |
| Deactivate nasabah | - | - | v | v |
| Reactivate nasabah | - | - | - | v |
| Transfer antar cabang | - | - | v | v |
| Merge duplikat | - | - | - | v |
| Update specimen tanda tangan | - | v | v | v |
| Upgrade KYC level | - | v | v | v |
| Buat pengajuan rekening (tabungan/deposito) | v | v | v | v |
| Approve pengajuan rekening (tabungan/deposito) | - | v | v | v |
| Approve pengajuan pinjaman | - | - | v | v |
| Freeze rekening | - | v | v | v |
| Unfreeze rekening | - | - | v | v |
| Tutup rekening | - | - | v | v |
| Hold/release saldo | - | v | v | v |

---

## Integration Points

| Integrasi | Arah | Keterangan |
|-----------|------|------------|
| `produk` (K003) | Rekening → Produk | Setiap rekening terikat pada satu produk |
| `transaksi` (K011) | Rekening ← Transaksi | Setiap transaksi rekening mempengaruhi saldo |
| `pinjaman` (K007) | Rekening → Pinjaman | Rekening PINJAMAN adalah representasi saldo outstanding |
| `simpanan_wajib_billing` (K004) | Nasabah ← Billing | Tagihan bulanan untuk setiap nasabah aktif |
| `jaminan` (K010) | Rekening ← Jaminan | Rekening tabungan/deposito bisa dijaminkan (hold_amount) |
| `shu_anggota` (K016) | Nasabah → SHU | Simpanan dan transaksi menjadi basis perhitungan SHU |
| `SekolahPro Core` | Nasabah ← Entitas Sekolah | `school_entity_id` merujuk ke siswa/guru/staf di modul sekolah |
| `toko_penjualan` (K019) | Nasabah → POS | Pembelian toko/kantin dengan debit tabungan |
| `payroll_authorization` (K020) | Nasabah → Payroll | Otorisasi potongan gaji dari rekening |

---

## Vernon _data Structure

**`_data` di nasabah:**
```json
{
  "branch": { "id": "...", "name": "Cabang Jakarta Pusat", "code": "JKT" },
  "sponsor": { "id": "...", "full_name": "Haji Ahmad", "member_number": "KOP-..." }
}
```

**`_data` di rekening:**
```json
{
  "nasabah": { "id": "...", "full_name": "Ahmad Fauzi", "member_number": "KOP-..." },
  "product": { "id": "...", "name": "Tabungan Berkah", "code": "TB-BERKAH", "category": "tabungan" },
  "branch": { "id": "...", "name": "Cabang Jakarta Pusat", "code": "JKT" }
}
```

**SyncEngine triggers:**
- `BranchUpdatedEvent` → update `_data.branch` di nasabah dan rekening
- `NasabahUpdatedEvent` → update `_data.nasabah` di rekening dan transaksi
- `ProductUpdatedEvent` → update `_data.product` di rekening
