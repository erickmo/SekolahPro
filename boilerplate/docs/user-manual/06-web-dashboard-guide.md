# 06 — Web Dashboard Guide

Panduan ini ditujukan untuk **end-user** yang menggunakan web dashboard sehari-hari,
dan **operator** yang mengkonfigurasi atau menyesuaikan dashboard untuk tim mereka.

---

## Akses Dashboard

Buka browser dan navigasikan ke URL yang diberikan oleh tim IT/operator Anda.

Contoh:
- Development: `http://localhost:5173`
- Production: `https://dashboard.perusahaan.com`

Browser yang didukung: Chrome 90+, Firefox 88+, Edge 90+, Safari 14+.

---

## Login dan Autentikasi

### Single-Tenant — Login Langsung

Jika organisasi Anda menggunakan deployment khusus (bukan SaaS):

1. Buka URL dashboard
2. Anda akan diarahkan otomatis ke halaman `/login`
3. Masukkan **username** (email) dan **password**
4. Klik **Login**
5. Anda langsung diarahkan ke `/dashboard`

```
Buka URL → /login → masukkan kredensial → /dashboard
```

### Multi-Tenant — Login Dua Langkah

Jika dashboard digunakan oleh banyak organisasi (platform SaaS):

**Langkah 1 — Login**

1. Buka URL dashboard
2. Masukkan username dan password
3. Klik **Login**

**Langkah 2 — Pilih Company**

Setelah login berhasil, Anda akan diarahkan ke halaman **Pilih Company** (`/choose-company`):

1. Daftar company/organisasi yang tersedia akan ditampilkan
2. Klik company yang ingin Anda akses
3. Anda diarahkan ke dashboard company tersebut: `/c/{kode-company}/dashboard`

```
Buka URL → /login → /choose-company → pilih company → /c/:code/dashboard
```

> **Tip:** Jika Anda hanya memiliki akses ke satu company, pilihan tetap
> ditampilkan untuk mengkonfirmasi konteks kerja Anda.

### Logout

Klik avatar/nama pengguna di pojok kanan atas navbar, lalu pilih **Logout**.
Semua session token akan dihapus dari browser.

---

## Navigasi

### Single-Tenant

Navigasi tersedia di sidebar/navbar kiri. Semua menu dapat diakses langsung
tanpa harus memilih konteks company terlebih dahulu.

```
[Logo]
[Dashboard]
[Produk]          ← menu yang didaftarkan developer
[Laporan]
[Pengaturan]
─────────────────
[Avatar: Nama User]
[Logout]
```

### Multi-Tenant

Navigasi berubah tergantung konteks (role dan company yang dipilih):

**Konteks Superuser (`/su/*`)**
```
[Logo] [Mode: SUPERUSER]
[Dashboard Global]
[Manajemen Tenant]
[Manajemen Company]
─────────────────
[Avatar: Nama User]
```

**Konteks HQ/Group (`/g/*`)**
```
[Logo] [Nama Group]
[Dashboard HQ]
[Laporan Konsolidasi]
─────────────────
[Switcher: Ganti Company ▾]
[Avatar: Nama User]
```

**Konteks Company (`/c/:code/*`)**
```
[Logo] [Nama Company]
[Dashboard]
[Menu-menu yang relevan]
─────────────────
[Switcher: Ganti Company ▾]
[Avatar: Nama User]
```

---

## Company Switcher (Multi-Tenant)

Di mode multi-tenant, Anda bisa berpindah antara company yang berbeda tanpa
harus logout dan login ulang.

**Cara menggunakan:**

1. Klik tombol **Switcher** di navbar (biasanya menampilkan nama company aktif)
2. Daftar company yang tersedia akan muncul
3. Pilih company yang ingin Anda akses
4. Halaman akan reload dengan konteks company baru

> **Penting:** Berpindah company akan memuat ulang semua data. Pastikan Anda
> sudah menyimpan pekerjaan yang sedang berlangsung sebelum berpindah.

---

## Role-Based Access

Dashboard menerapkan kontrol akses berbasis role. Halaman atau menu tertentu
hanya terlihat dan dapat diakses oleh pengguna dengan role yang sesuai.

### Daftar Role

| Role | Deskripsi | Akses |
|------|-----------|-------|
| `superuser` | Administrator platform SaaS | Semua fitur, semua tenant |
| `tenant_owner` | Pemilik/admin organisasi | Semua company dalam tenant mereka |
| `employee` | Karyawan biasa | Company tertentu sesuai assignment |

### Alur Redirect Berdasarkan Role (Multi-Tenant)

Setelah login, sistem akan mengarahkan Anda secara otomatis:

| Role | Diarahkan ke |
|------|-------------|
| `superuser` | `/su/dashboard` |
| `tenant_owner` | `/choose-company` (pilih company untuk dikelola) |
| `employee` dengan company aktif | `/c/:code/dashboard` |
| `employee` tanpa company aktif | `/choose-company` |

### Jika Anda Mengakses Halaman Terlarang

Jika mencoba mengakses URL yang tidak sesuai dengan role Anda, sistem akan
menampilkan halaman **403 Forbidden** dengan pesan yang menjelaskan alasannya.

Solusi: kembali ke dashboard utama atau hubungi administrator untuk memeriksa
hak akses Anda.

---

## Toast Notifications

Sistem menampilkan notifikasi singkat (toast) di pojok layar untuk memberikan
umpan balik terhadap aksi yang Anda lakukan.

### Jenis Notifikasi

| Warna | Jenis | Contoh Pesan |
|-------|-------|-------------|
| Hijau | Sukses | "Data berhasil disimpan" |
| Merah | Error | "Terjadi kesalahan, coba lagi" |
| Kuning | Peringatan | "Perubahan belum disimpan" |
| Biru | Informasi | "Memuat data..." |

### Perilaku Toast

- Toast muncul di pojok kanan bawah layar
- Hilang otomatis setelah beberapa detik
- Bisa ditutup manual dengan klik tombol ×
- Beberapa toast bisa muncul bersamaan (ditumpuk)

---

## Common UI Patterns

### Tabel Data

Semua halaman daftar (list) menggunakan pola yang konsisten:

```
┌──────────────────────────────────────────────────────┐
│ [Judul Halaman]                    [+ Tambah Baru]   │
│──────────────────────────────────────────────────────│
│ [🔍 Cari...]          [Filter ▾]  [Sort ▾]           │
│──────────────────────────────────────────────────────│
│  Nama          │  Status  │  Tanggal   │  Aksi       │
│────────────────┼──────────┼────────────┼─────────────│
│  Item Pertama  │  Aktif   │  14/04/26  │  [Edit][×]  │
│  Item Kedua    │  Nonaktif│  13/04/26  │  [Edit][×]  │
│──────────────────────────────────────────────────────│
│  ← Sebelumnya     Hal 1 dari 8     Selanjutnya →    │
└──────────────────────────────────────────────────────┘
```

**Navigasi halaman:** Gunakan tombol **Sebelumnya** / **Selanjutnya** di bawah tabel,
atau klik nomor halaman langsung.

### Formulir

Formulir tambah/edit mengikuti pola:

1. Field wajib ditandai dengan ***** merah
2. Error validasi muncul di bawah field yang bermasalah
3. Tombol **Simpan** disabled selama proses penyimpanan
4. Berhasil disimpan → toast hijau + kembali ke halaman daftar
5. Gagal → toast merah + form tetap terbuka dengan error

### Konfirmasi Hapus

Sebelum menghapus data, sistem akan menampilkan dialog konfirmasi:

```
┌─────────────────────────────────────────┐
│  ⚠ Konfirmasi Hapus                     │
│                                         │
│  Apakah Anda yakin ingin menghapus     │
│  "Nama Item"? Tindakan ini tidak       │
│  dapat dibatalkan.                      │
│                                         │
│          [Batal]  [Hapus]              │
└─────────────────────────────────────────┘
```

Klik **Hapus** untuk konfirmasi, atau **Batal** untuk membatalkan.

### Loading State

Ketika data sedang dimuat dari server, tabel/halaman akan menampilkan
indikator loading (skeleton atau spinner). Tunggu hingga data selesai dimuat.

---

## Troubleshooting

### Tidak bisa login

1. Pastikan username dan password sudah benar (case-sensitive)
2. Coba reset password jika tersedia
3. Cek apakah API server berjalan: hubungi tim IT
4. Cek apakah URL dashboard benar

### Halaman blank / putih setelah login

1. Hard refresh browser: `Ctrl+Shift+R` (Windows/Linux) atau `Cmd+Shift+R` (Mac)
2. Clear cache browser
3. Coba browser yang berbeda
4. Jika masalah berlanjut, hubungi tim IT

### Data tidak muncul / tabel kosong

1. Pastikan Anda berada di company yang benar (multi-tenant)
2. Cek apakah ada filter yang aktif — klik **Reset Filter**
3. Refresh halaman
4. Kemungkinan belum ada data — coba tambahkan data pertama

### "403 Forbidden" muncul

Anda tidak memiliki izin untuk mengakses halaman tersebut. Hubungi administrator
untuk memastikan role Anda sudah benar.

### Perubahan tidak tersimpan

1. Pastikan koneksi internet stabil
2. Periksa pesan error yang muncul (toast merah)
3. Coba simpan ulang

### Session expired / keluar otomatis

Demi keamanan, session akan expired setelah periode tidak aktif. Login kembali
menggunakan kredensial Anda.

---

## Shortcut Keyboard

| Shortcut | Fungsi |
|----------|--------|
| `Esc` | Tutup modal / dialog |
| `Enter` | Submit form (jika fokus di dalam form) |
| `Tab` | Pindah ke field berikutnya |
| `Shift+Tab` | Pindah ke field sebelumnya |

---

## Pengaturan Profil

Akses pengaturan profil melalui menu di pojok kanan atas (klik avatar atau nama Anda):

- **Profil:** Lihat dan edit informasi profil (nama, email)
- **Ganti Password:** Ubah password akun Anda
- **Audit Log:** Riwayat aktivitas akun (jika tersedia)
- **Logout:** Keluar dari sistem

---

## Tips Penggunaan

1. **Gunakan keyboard shortcuts** untuk navigasi yang lebih cepat di form
2. **Bookmark URL halaman yang sering dikunjungi** — URL sudah mencerminkan
   halaman aktif dan company context (multi-tenant)
3. **Refresh halaman** jika data terlihat tidak sinkron (mungkin ada update
   dari user lain yang belum muncul)
4. **Gunakan filter dan sort** untuk menemukan data lebih cepat di tabel
   dengan banyak entri
5. **Di mode multi-tenant**, selalu perhatikan nama company di navbar untuk
   memastikan Anda bekerja di konteks yang benar
