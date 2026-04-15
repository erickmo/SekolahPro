# ADR-009: Dual-Mode Institution Type (General / Islamic)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

SekolahPro melayani dua segmen institusi pendidikan di Indonesia:

1. **Sekolah Umum** — SD, SMP, SMA, SMK dengan koperasi sekolah konvensional
2. **Sekolah Islam** — Pondok Pesantren dengan BMT (Baitul Maal wat Tamwil)

Kedua segmen memiliki:
- **Terminologi berbeda** — contoh: Siswa vs Santri, Kelas vs Halaqah, Bunga vs Bagi Hasil
- **Fitur tambahan** — contoh: modul hafalan Quran hanya untuk pesantren, akad syariah hanya untuk BMT
- **Business rule berbeda** — contoh: perhitungan bunga (konvensional) vs bagi hasil/margin (syariah)

Project ini terdiri dari dua subproject independen:
- **Management Sekolah** — pengelolaan akademik
- **Management Koperasi Sekolah** — pengelolaan koperasi/BMT

Kedua subproject bisa digunakan secara independen atau bersamaan, dan masing-masing memiliki mode sendiri. Sebuah tenant bisa menjalankan kombinasi apapun (misal: Pesantren + Koperasi konvensional).

## Decision

### 1. Dua field type terpisah di tenant config

Setiap tenant memiliki dua field independen:

```
school_type: "general" | "islamic"
coop_type:   "general" | "islamic"
```

Kedua field ini **tidak terikat satu sama lain** — menghasilkan 4 kemungkinan kombinasi:

| school_type | coop_type | Konteks Penggunaan |
|-------------|-----------|-------------------|
| `general` | `general` | Sekolah umum + Koperasi konvensional |
| `general` | `islamic` | Sekolah umum + BMT |
| `islamic` | `general` | Pondok Pesantren + Koperasi konvensional |
| `islamic` | `islamic` | Pondok Pesantren + BMT |

Jika tenant hanya menggunakan satu subproject, field yang tidak digunakan bisa bernilai `null`.

### 2. Immutable setelah tenant aktif

`school_type` dan `coop_type` **tidak dapat diubah** setelah tenant di-setup dan aktif. Alasan:

- Perubahan type berdampak pada business rule, kalkulasi finansial, dan data historis
- Migrasi data antar mode berisiko tinggi (misal: konversi bunga ke bagi hasil pada transaksi lama)
- Mencegah inkonsistensi data yang sulit di-debug

Field ini di-set **satu kali saat onboarding tenant** dan di-enforce di application layer.

### 3. Terminologi sebagai konfigurasi, bukan hardcode

Perbedaan terminologi antar mode di-handle melalui **label mapping** yang dikonfigurasi berdasarkan type, bukan hardcode di UI:

```
school_type = "general"  → { student: "Siswa",  class: "Kelas",    teacher: "Guru" }
school_type = "islamic"  → { student: "Santri", class: "Halaqah",  teacher: "Ustadz" }

coop_type = "general"    → { interest: "Bunga",      loan: "Pinjaman" }
coop_type = "islamic"    → { interest: "Bagi Hasil", loan: "Pembiayaan" }
```

### 4. Fitur tambahan menggunakan feature flag per type

Fitur yang hanya tersedia di mode tertentu diaktifkan berdasarkan type:

```
school_type = "islamic" → enable: [tahfidz_module, diniyah_schedule, kitab_tracking]
coop_type   = "islamic" → enable: [akad_syariah, bagi_hasil_calculator, zakat_module]
```

Feature flag ini **derived dari type** (bukan konfigurasi terpisah), sehingga tidak bisa diaktifkan secara manual di mode yang salah.

### 5. Business rule dispatch berdasarkan type

Logika bisnis yang berbeda antar mode di-handle melalui **strategy pattern**:

```
// Pseudocode
interface ProfitCalculator {
    Calculate(principal, rate, period) → result
}

// general: bunga konvensional
ConventionalCalculator implements ProfitCalculator

// islamic: bagi hasil / margin murabahah
IslamicCalculator implements ProfitCalculator
```

Service layer memilih strategy berdasarkan `coop_type` dari tenant context.

## C-Suite Strategic Review (2026-04-15)

> **CEO + CMO Review:**
> ADR ini memperlakukan `general` dan `islamic` sebagai pilihan yang equal. Secara teknis
> ini benar — kedua mode memiliki bobot implementasi yang sama. Namun secara **strategi bisnis**,
> **pesantren/Islamic adalah primary market** SekolahPro, bukan equal alternative:
>
> 1. **Pesantren membutuhkan ALL-IN-ONE** (akademik + koperasi/BMT + asrama + kantin) —
>    sekolah umum biasanya hanya butuh sebagian. Pesantren = higher ARPU.
> 2. **Fitur Islamic bukan "tambahan"** — tahfidz, diniyah, kitab tracking, akad syariah
>    adalah **core differentiator** yang tidak dimiliki kompetitor manapun.
> 3. **Sekolah umum sudah punya banyak opsi** (Jibas, AdminSekolah, Pijar) — pesantren underserved.
>
> **Implikasi implementasi:**
> - Default onboarding flow harus **optimize untuk pesantren** (`islamic` sebagai preset pertama)
> - Demo/marketing materials harus lead dengan pesantren use case
> - Feature flag untuk Islamic mode harus dikembangkan **di Phase 1**, bukan ditunda
> - `general` mode tetap tersedia tetapi sebagai **simplified variant**, bukan primary focus
>
> **Ini TIDAK mengubah arsitektur teknis ADR ini** — keputusan dual-mode dan strategy pattern
> tetap benar. Yang berubah adalah **prioritas implementasi dan go-to-market strategy**.

## Consequences

### Positif

- **Fleksibel** — 4 kombinasi mode tanpa code branching yang kompleks
- **Aman** — immutability mencegah inkonsistensi data
- **Scalable** — menambah mode baru (misal: `international`) hanya perlu extend enum dan implementasi strategy
- **Clean separation** — business rule per-mode terisolasi di strategy masing-masing
- **Market reach** — satu codebase bisa melayani sekolah umum dan pesantren

### Negatif

- **Tidak bisa migrasi mode** — jika tenant salah pilih, harus buat tenant baru
- **Duplikasi strategy** — setiap business rule yang berbeda antar mode butuh implementasi terpisah
- **Testing matrix** — 4 kombinasi mode memperbesar test surface

### Mitigasi

- Onboarding flow harus memiliki **konfirmasi eksplisit** sebelum finalize type
- Strategy implementations harus di-unit test secara independen per mode
- Integration test minimal harus cover 4 kombinasi mode

## Alternatives Considered

### A. Single `institution_type` field

```
institution_type: "general" | "islamic"
```

Satu field menentukan mode untuk sekolah dan koperasi sekaligus.

**Ditolak** karena: tidak mengakomodasi kasus pesantren yang koperasinya konvensional (atau sebaliknya). Memaksa pasangan mode yang kaku.

### B. Mutable type dengan migration tool

Mengizinkan perubahan type dengan tool migrasi data otomatis.

**Ditolak** karena: risiko inkonsistensi data finansial terlalu tinggi. Konversi kalkulasi bunga ke bagi hasil pada data historis tidak reliabel dan bisa menimbulkan masalah audit.

### C. Dua codebase terpisah (fork)

Memisahkan codebase untuk versi umum dan Islam.

**Ditolak** karena: duplikasi maintenance yang tinggi. Bug fix dan fitur baru harus diterapkan di dua tempat. Tidak scalable jika ada mode ketiga di masa depan.
