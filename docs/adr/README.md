# Architecture Decision Records (ADR)

Direktori ini berisi Architecture Decision Records untuk project **SekolahPro Boilerplate**.

ADR adalah dokumen yang merekam keputusan arsitektur penting beserta konteks, alasan, dan konsekuensinya. Tujuannya adalah memberikan historical context bagi anggota tim yang bergabung di kemudian hari, dan mencegah re-debating keputusan yang sudah dibuat.

## Subprojects

Project ini terdiri dari **dua subproject** yang bisa digunakan secara independen atau bersamaan:

| Subproject | Deskripsi |
|------------|-----------|
| **Management Sekolah** | Pengelolaan akademik sekolah (siswa, guru, kelas, jadwal, nilai, dll) |
| **Management Koperasi Sekolah** | Pengelolaan koperasi sekolah (anggota, simpan pinjam, inventaris, dll) |

## Struktur Direktori

```
docs/adr/
├── core/       ← Shared foundation (berlaku untuk SEMUA subproject)
├── sekolah/    ← Spesifik Management Sekolah
└── koperasi/   ← Spesifik Management Koperasi Sekolah
```

Baca ADR per scope:
- **[Core (Shared)](./core/)** — Stack, pattern, infrastruktur dasar
- **[Management Sekolah](./sekolah/)** — Domain akademik sekolah
- **[Management Koperasi Sekolah](./koperasi/)** — Domain koperasi sekolah

## Format

Setiap ADR menggunakan format standar:

```
# ADR-XXX: [Judul]

Status:     Accepted | Proposed | Deprecated | Superseded by ADR-YYY
Date:       YYYY-MM-DD
Deciders:   [Tim / individu yang membuat keputusan]

## Context
## Decision
## Consequences
## Alternatives Considered
```

## Status Definitions

| Status | Arti |
|--------|------|
| **Proposed** | Sedang dalam diskusi, belum final |
| **Accepted** | Keputusan sudah dibuat dan aktif diterapkan |
| **Deprecated** | Masih berlaku tapi tidak direkomendasikan untuk penggunaan baru |
| **Superseded** | Digantikan oleh ADR lain (selalu sertakan nomor ADR pengganti) |

---

## Penomoran ADR

Setiap subdirektori memiliki penomoran **independen** dengan prefix:

| Scope | Prefix | Contoh |
|-------|--------|--------|
| Core | `ADR-0XX` | ADR-001, ADR-009, ... |
| Sekolah | `ADR-S0XX` | ADR-S001, ADR-S002, ... |
| Koperasi | `ADR-K0XX` | ADR-K001, ADR-K002, ... |

## Cara Membuat ADR Baru

1. Tentukan scope: Core, Sekolah, atau Koperasi
2. Buka README di subdirektori yang sesuai untuk lihat nomor terakhir
3. Beri nomor berikutnya dengan prefix yang benar
4. Isi semua section: Context, Decision, Consequences, Alternatives Considered
5. Update tabel index di README subdirektori
6. Minta review dari minimal satu engineer lain
7. Set status ke `Accepted` setelah disetujui

## Panduan Kapan Membuat ADR

Buat ADR ketika memilih atau mengubah:
- Framework atau library utama
- Database schema pattern
- Authentication / authorization approach
- Communication pattern (sync vs async, REST vs gRPC)
- Deployment architecture
- Testing strategy
- Performance optimization technique yang berdampak pada desain

**Tidak perlu ADR untuk**: keputusan implementasi detail yang bisa diubah tanpa dampak arsitektur (nama variabel, utility function, minor refactor).

---

*ADR template ini mengikuti [Michael Nygard's ADR format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).*
