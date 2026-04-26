# ADR-015: Frontend Architecture — Multi-App Strategy

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: Erick Mo, CTO Review Panel
**Supersedes**: Memperluas ADR-008 (React 18 + Vite + CSS Modules)

## Context

ADR-008 menetapkan React 18 + Vite + CSS Modules sebagai stack frontend. Keputusan tersebut solid untuk satu dashboard, namun SekolahPro melayani **16 role stakeholder** (ADR-013) dengan kebutuhan UX yang sangat berbeda:

| Aplikasi | Pengguna | Karakteristik UX |
|----------|----------|-------------------|
| **Admin Dashboard** | Admin Sekolah, Kepala Sekolah, TU, Wakasek | Data-heavy: tabel besar, bulk operations, form kompleks. Desktop-first. |
| **Portal Orang Tua & Guru** | Orang Tua, Guru, Wali Kelas, Guru BK | Mobile-first, read-heavy, notification-driven. Akses via HP di jaringan 3G. |
| **Teller / POS** | Bendahara, Kasir Kantin | Transaction-critical, fast input, minimal UI. Zero tolerance untuk lag. |

### Masalah dengan Single SPA

Satu React SPA untuk semua role mengakibatkan:

1. **Bundle bloat** — Admin butuh data table, chart library, rich text editor. Parent portal tidak butuh semua itu. Tapi semua masuk ke initial bundle atau lazy-loaded chunks yang tetap di-download.
2. **Performance mismatch** — POS butuh Time-to-Interactive < 2 detik di hardware sederhana. Admin dashboard bisa toleransi 3-4 detik di desktop. Satu app tidak bisa dioptimasi untuk keduanya.
3. **Offline strategy berbeda** — Portal orang tua butuh PWA dengan offline cache untuk rapor/absensi. POS butuh offline queue untuk transaksi. Admin tidak butuh offline.
4. **Deployment coupling** — Bug fix di POS tidak seharusnya memerlukan redeploy seluruh admin dashboard.
5. **UX focus terpecah** — Mobile-first dan desktop-first dalam satu codebase menghasilkan kompromi di kedua sisi.

## Decision

Mengadopsi **3 aplikasi React terpisah** dalam satu **pnpm monorepo**, sharing common packages.

### Aplikasi

```
apps/
  app-admin/       # Dashboard admin, Kepala Sekolah, TU, Wakasek
  app-portal/      # Portal Orang Tua + Guru (PWA, mobile-first)
  app-pos/         # Teller session + POS kantin
```

| App | Target Device | Bundle Target (gzip) | Offline | Auth Flow |
|-----|--------------|----------------------|---------|-----------|
| `app-admin` | Desktop browser | < 300 KB initial | Tidak | JWT + scope selection (ADR-004) |
| `app-portal` | Mobile browser (3G) | < 120 KB initial | PWA + service worker | JWT + parent/teacher scope |
| `app-pos` | Tablet/desktop dedicated | < 80 KB initial | Offline queue (transactions) | JWT + teller session lock |

### Shared Packages

```
packages/
  ui/              # Shared component library (Button, Table, Form, Layout)
  api-client/      # Generated TypeScript client dari OpenAPI spec
  auth/            # JWT handling, token refresh, scope management
  types/           # Shared TypeScript types & domain interfaces
```

- **`packages/ui`** — Design token system (CSS variables dari ADR-008) tetap berlaku. Komponen headless yang di-style per app sesuai konteks (compact untuk POS, spacious untuk admin).
- **`packages/api-client`** — Auto-generated dari OpenAPI spec backend. Satu sumber kebenaran untuk semua API calls.
- **`packages/auth`** — JWT decode, refresh logic, scope store (Zustand). Shared agar auth behavior konsisten di semua app.

### Monorepo Structure

```
sekolahpro-frontend/
  pnpm-workspace.yaml
  package.json              # Root scripts: build:all, lint:all, test:all
  apps/
    app-admin/
      vite.config.ts
      src/
      package.json          # deps: @sekolahpro/ui, @sekolahpro/api-client
    app-portal/
      vite.config.ts        # PWA plugin, compression, 3G budget
      src/
      package.json
    app-pos/
      vite.config.ts        # Minimal deps, aggressive tree-shaking
      src/
      package.json
  packages/
    ui/
    api-client/
    auth/
    types/
  tsconfig.base.json        # Shared TypeScript config
```

Tool: **pnpm workspaces** + **Turborepo** untuk build orchestration dan caching.

## Deployment

Setiap app di-build dan di-deploy secara independen:

| App | URL | CDN Path | CI Trigger |
|-----|-----|----------|------------|
| `app-admin` | `admin.sekolahpro.id` | `/admin/` | changes di `apps/app-admin/` atau `packages/` |
| `app-portal` | `portal.sekolahpro.id` | `/portal/` | changes di `apps/app-portal/` atau `packages/` |
| `app-pos` | `pos.sekolahpro.id` | `/pos/` | changes di `apps/app-pos/` atau `packages/` |

- Nginx/reverse proxy mengarahkan subdomain ke static files masing-masing.
- Shared packages change memicu rebuild **semua** app yang depend padanya (Turborepo dependency graph).
- Setiap app punya Dockerfile sendiri untuk containerized deployment.

## Consequences

### Positif

- **Bundle optimal per audiens** — POS < 80 KB, tidak membawa chart library yang hanya admin butuhkan.
- **Independent deployment** — Hotfix POS bisa deploy dalam 2 menit tanpa menyentuh admin.
- **Focused UX** — Tim bisa desain mobile-first untuk portal tanpa kompromi desktop admin.
- **Offline strategy tepat sasaran** — PWA hanya di portal, offline queue hanya di POS.
- **Scalable team structure** — Tim bisa split: Tim A fokus admin, Tim B fokus portal, shared package dijaga bersama.
- **Reuse via packages** — Komponen UI, API client, dan auth logic ditulis sekali, dipakai tiga app.

### Negatif / Trade-offs

- **Build complexity naik** — Monorepo tooling (pnpm workspaces + Turborepo) perlu setup dan maintenance.
- **Shared package versioning** — Breaking change di `packages/ui` berdampak ke semua app. Perlu semver internal dan CI checks.
- **Duplikasi konfigurasi** — Setiap app punya `vite.config.ts`, `tsconfig.json`, test setup sendiri. Mitigasi: extend dari base config.
- **Developer onboarding** — Developer baru perlu memahami monorepo structure. Mitigasi: dokumentasi dan `pnpm dev:admin` / `pnpm dev:portal` scripts.
- **Cross-app feature** — Fitur yang muncul di lebih dari satu app (e.g., notifikasi) perlu koordinasi. Mitigasi: shared package atau event-driven via backend.

## Alternatives Considered

### 1. Single SPA dengan Code Splitting (Lazy Routes)

- Satu app, `React.lazy()` per route group (admin routes, portal routes, POS routes).
- **Ditolak** karena: shared vendor chunks tetap besar, tidak bisa optimize offline strategy per audiens, deployment tetap coupled, dan PWA scope sulit dibatasi per route group.

### 2. Micro-frontends (Module Federation)

- Webpack Module Federation atau Vite federation plugin. Setiap "app" jadi remote module yang di-load runtime.
- **Ditolak** karena: runtime overhead untuk federation bootstrap, shared dependency versioning lebih kompleks dari monorepo, debugging cross-module issues sulit, dan overkill untuk 3 app yang tidak perlu runtime composition.

### 3. Framework Berbeda per App

- Admin: React. Portal: Svelte (lighter). POS: Preact (minimal).
- **Ditolak** karena: tidak bisa share component library, developer perlu menguasai 3 framework, hiring lebih sulit, dan maintenance cost berlipat.

### 4. Status Quo (Single React App — ADR-008 as-is)

- Tetap satu SPA, optimasi dengan lazy loading.
- **Ditolak** karena masalah yang diuraikan di Context section. Bisa bekerja untuk MVP, tapi tidak scalable untuk 16 role dengan kebutuhan UX berbeda.

## References

- ADR-008: React 18 + Vite + CSS Modules
- ADR-004: Multi-Tenant Architecture
- ADR-013: Users & Roles
- [pnpm Workspaces](https://pnpm.io/workspaces)
- [Turborepo](https://turbo.build/repo)
