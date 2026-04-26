# 05 — Arsitektur Frontend SekolahPro

Dokumen ini menjabarkan arsitektur frontend SekolahPro: multi-app strategy di atas pnpm monorepo, stack teknis (React 18 + Vite + CSS Modules), sistem state management, koneksi ke backend Vernon read-cache, dan strategi testing.

---

## ADR References

| ADR | Judul | Relevansi |
|-----|-------|-----------|
| ADR-008 | React 18 + Vite + CSS Modules | Stack teknologi frontend dasar |
| ADR-015 | Frontend Architecture — Multi-App Strategy | Tiga aplikasi terpisah dalam satu monorepo |

---

## 1. Multi-App Strategy

SekolahPro melayani **16 role stakeholder** dengan kebutuhan UX yang sangat berbeda. Satu SPA tidak bisa dioptimasi untuk semua audiens secara bersamaan.

### Tiga Aplikasi Terpisah

```
apps/
  app-admin/    # Dashboard — Admin Sekolah, Kepala Sekolah, TU, Wakasek
  app-portal/   # Portal — Orang Tua, Guru, Wali Kelas, Guru BK (PWA)
  app-pos/      # Teller — Bendahara, Kasir Koperasi (transaction-critical)
```

| App | Target Device | Bundle Target (gzip) | Offline | Auth |
|-----|--------------|---------------------|---------|------|
| `app-admin` | Desktop browser | < 300 KB initial | Tidak | JWT + scope selection |
| `app-portal` | Mobile (3G) | < 120 KB initial | PWA + service worker | JWT + parent/teacher scope |
| `app-pos` | Tablet/desktop | < 80 KB initial | Offline queue transaksi | JWT + teller session lock |

### Shared Packages

```
packages/
  ui/           # Shared component library (Button, Table, Form, Layout)
  api-client/   # TypeScript client dari OpenAPI spec (auto-generated)
  auth/         # JWT handling, token refresh, scope management (Zustand)
  types/        # Shared TypeScript types & domain interfaces
```

- **`packages/ui`**: Komponen headless yang di-style per app. Design token (CSS variables) berlaku di semua app
- **`packages/api-client`**: Auto-generated dari OpenAPI spec backend — satu sumber kebenaran untuk semua API calls
- **`packages/auth`**: Shared JWT decode, refresh logic, scope store (Zustand) — auth behavior konsisten di semua app

### Monorepo Structure Lengkap

```
sekolahpro-frontend/
  pnpm-workspace.yaml
  package.json              # Root scripts: build:all, lint:all, test:all
  turbo.json                # Turborepo build pipeline & caching
  tsconfig.base.json        # Shared TypeScript config

  apps/
    app-admin/
      vite.config.ts        # Desktop-optimized, chart library chunks
      src/
        pages/              # Route-level components
        features/           # Feature-scoped components + hooks
        components/         # App-specific reusable components
      package.json

    app-portal/
      vite.config.ts        # PWA plugin, aggressive compression
      src/
      package.json

    app-pos/
      vite.config.ts        # Minimal deps, aggressive tree-shaking
      src/
      package.json

  packages/
    ui/src/                 # Button, Table, Modal, Form, Layout, etc.
    api-client/src/         # Generated API client per domain
    auth/src/               # JWT + scope store
    types/src/              # Domain interfaces, DTOs
```

Tool: **pnpm workspaces** + **Turborepo** untuk build orchestration dan caching.

---

## 2. Stack Teknologi

### Build Tool: Vite

```
Development server start time:
  Webpack (CRA): 8-15 detik
  Vite:          300-500ms

HMR (Hot Module Replacement):
  Webpack:  100-500ms per change
  Vite:     10-50ms per change (native ESM)

Production build:
  Webpack: 60-120 detik
  Vite:    15-30 detik (Rollup-based)
```

Vite menggunakan native ESM di development — tidak ada bundling. Production menggunakan Rollup dengan code splitting otomatis.

### Framework: React 18 (bukan Next.js)

SekolahPro adalah SPA murni — tidak membutuhkan SSR/SSG. Next.js overhead tidak justified:
- Routing conventions Next.js terlalu opinionated untuk multi-app setup
- Server components complexity tidak diperlukan untuk dashboard internal
- Jika SEO/landing page dibutuhkan di masa depan, bisa ditambah sebagai app keempat

React 18 Concurrent Features yang digunakan: `Suspense` untuk lazy loading, `startTransition` untuk non-urgent state updates.

### Styling: CSS Modules (tanpa Tailwind / UI Library)

Keputusan ini menghindari:
- **Tailwind**: HTML verbose dengan banyak utility class, sulit enforce design consistency, tidak encapsulated
- **MUI / Ant Design**: Bundle besar (~70KB gzipped), opinionated design sulit di-override untuk custom branding, version lock
- **CSS-in-JS**: Runtime overhead per render, SSR complexity, debugging class names sulit

CSS Modules memberikan:
- Scoped styles — tidak ada class name conflict antar komponen
- Zero runtime overhead — class name hashing dilakukan saat build
- TypeScript support via `import styles from './X.module.css'`
- Full CSS power: media queries, pseudo-classes, nesting

### Design Token System

Semua nilai visual didefinisikan sebagai CSS variables di `src/styles/tokens.css`:

```css
:root {
    /* Brand */
    --color-brand-primary:   #2563EB;
    --color-brand-secondary: #0EA5E9;

    /* Semantic colors */
    --text-primary:          #111827;
    --surface-primary:       #FFFFFF;
    --border-subtle:         #E5E7EB;

    /* Typography scale */
    --text-sm:    0.875rem;
    --text-base:  1rem;
    --text-lg:    1.125rem;
    --text-xl:    1.25rem;
    --text-2xl:   1.5rem;

    /* Spacing scale */
    --space-1: 0.25rem;
    --space-2: 0.5rem;
    --space-4: 1rem;
    --space-8: 2rem;

    /* Shadows, radius, etc. */
}

/* Per-tenant branding override */
[data-tenant="yayasan-al-hikmah"] {
    --color-brand-primary:   #1E40AF;
    --color-brand-secondary: #3B82F6;
}
```

Custom branding per tenant hanya perlu override CSS variables — tidak ada JS theming complexity.

---

## 3. State Management

```
State Layer:
├── Server State    → TanStack Query v5
│   ├── caching, background refetch, pagination
│   └── optimistic updates, infinite queries
├── Global UI State → Zustand
│   ├── sidebar open/close
│   ├── selected tenant/company scope (dari Phase 2 JWT)
│   └── notification queue
└── Local State     → useState / useReducer
    ├── form state
    └── component-specific UI state
```

### Scope Store (Zustand)

Menyimpan active tenant/company scope yang dipilih setelah Phase 2 JWT:

```typescript
// packages/auth/src/scope.store.ts
interface ScopeState {
    tenantID:    string | null;
    companyID:   string | null;
    branchID:    string | null;
    warehouseID: string | null;
    roles:       string[];
    permissions: string[];
    teacherID:   string | null;    // untuk guru
    classRoomIDs: string[];        // untuk wali kelas/guru
}
```

### API Integration (TanStack Query)

```typescript
// Contoh hook untuk Vernon domain
export function useStudents(filters: StudentFilters) {
    const scope = useScopeStore();

    return useQuery({
        queryKey: ['students', scope.tenantID, scope.companyID, filters],
        queryFn: () => studentApi.list({
            ...filters,
            tenantID:  scope.tenantID,
            companyID: scope.companyID,
        }),
        staleTime: 30_000,     // 30 detik — Vernon _data bersifat eventually consistent
        enabled: !!scope.tenantID,
    });
}
```

Catatan: `staleTime: 30_000` selaras dengan SLA SyncEngine "normal" priority (< 30 detik). Data Vernon boleh stale selama itu.

---

## 4. Koneksi Frontend ke Backend Vernon Read-Cache

### Alur Data Read

```
React Component
    │
    ▼
TanStack Query hook (useStudents, useClassRooms, ...)
    │
    ▼
packages/api-client (auto-generated dari OpenAPI spec)
    │
    ▼
HTTP GET /api/v1/students?tenant_id=...&company_id=...
    │
    ▼
Go Backend: usecase/query/list_students/handler.go
    │
    ▼
Vernon Reader: SELECT id, ..., _data FROM students WHERE tenant_id = $1
               (ZERO JOIN — semua dari _data JSONB)
    │
    ▼
Response DTO (JSON) — termasuk data dari _data field
    │
    ▼
TanStack Query cache
    │
    ▼
React component renders
```

### Handling Eventually Consistent Data

Vernon `_data` bersifat eventually consistent (stale < 30 detik untuk "normal" priority). Frontend harus:

1. **`staleTime: 30_000`** di TanStack Query — tidak re-fetch terlalu agresif
2. **Tidak menampilkan loading spinner** untuk background refetch
3. **Optimistic updates** untuk operasi write yang sering: update cache lokal segera, konfirmasi dari server di background
4. **Indikasi visual minimal** jika data mungkin stale — umumnya tidak perlu, kecuali untuk data finansial kritikal

### API Client Generation

```
Backend OpenAPI spec
    │
    ▼ (auto-generate via openapi-typescript-codegen atau orval)
packages/api-client/src/
    ├── students.ts
    ├── class-rooms.ts
    ├── academic-years.ts
    └── ...
```

Satu sumber kebenaran untuk semua API calls. Jika backend mengubah response schema, client code di-regenerate dan TypeScript akan mendeteksi breaking changes.

---

## 5. Testing Stack

```
Unit Tests (Vitest):
├── Pure functions, utilities, formatters
├── Custom hooks (renderHook)
└── Component render + interaction

API Mocking (MSW v2):
├── Service Worker intercepts fetch di test
├── Handler definitions per feature
└── Same handlers bisa digunakan di Storybook

E2E Tests (Playwright):
├── Critical user journeys (login, input nilai, generate rapor)
├── Multi-browser (Chrome, Firefox, Safari)
└── Visual regression (screenshots)
```

### Vite Config

```typescript
// vite.config.ts
export default defineConfig({
    plugins: [react()],
    resolve: { alias: { '@': path.resolve(__dirname, './src') } },
    css: {
        modules: {
            localsConvention: 'camelCase',
            generateScopedName: '[name]__[local]__[hash:6]',
        },
    },
    build: {
        rollupOptions: {
            output: {
                manualChunks: {
                    'vendor-react':  ['react', 'react-dom'],
                    'vendor-query':  ['@tanstack/react-query'],
                    'vendor-charts': ['recharts'],   // app-admin only
                },
            },
        },
    },
});
```

---

## 6. Deployment

Setiap app di-build dan di-deploy secara independen:

| App | URL | CDN Cache |
|-----|-----|-----------|
| `app-admin` | `admin.sekolahpro.id` | Static files, CDN-backed |
| `app-portal` | `portal.sekolahpro.id` | Static files + PWA cache |
| `app-pos` | `pos.sekolahpro.id` | Static files + offline queue |

- Setiap app memiliki Dockerfile sendiri (`nginx:alpine` serving static files)
- Perubahan di `packages/` memicu rebuild **semua** app yang depend padanya (Turborepo dependency graph)
- Hotfix di `app-pos` dapat di-deploy dalam ~2 menit tanpa menyentuh `app-admin`

---

## Key Decisions

1. **3 aplikasi terpisah, bukan 1 SPA** — memungkinkan optimasi bundle, UX, dan deployment strategy yang berbeda per audiens

2. **pnpm monorepo + Turborepo** — code sharing via `packages/` tanpa deployment coupling

3. **CSS Modules bukan Tailwind/MUI** — custom branding mudah via CSS variables, bundle size terkontrol (< 120-300 KB)

4. **TanStack Query untuk server state** — caching, background refetch, optimistic updates selaras dengan Vernon eventual consistency

5. **Zustand untuk global UI state** — minimal boilerplate untuk scope store dan UI state dibandingkan Redux Toolkit

6. **Auto-generated API client dari OpenAPI spec** — satu sumber kebenaran; TypeScript mendeteksi breaking changes dari backend

7. **`staleTime: 30_000`** — selaras dengan SLA SyncEngine "normal" priority; tidak over-fetch untuk data yang boleh stale

---

## Constraints & Implications

### Constraints

- Setiap app harus memenuhi bundle target masing-masing (admin < 300KB, portal < 120KB, pos < 80KB)
- Tidak ada business logic di React component — semua di custom hooks atau `packages/`
- CSS Modules hanya boleh menggunakan CSS variables dari `tokens.css` untuk nilai visual (tidak boleh arbitrary values)
- API client wajib digunakan untuk semua HTTP calls — tidak boleh fetch langsung ke URL string
- Permission check harus menggunakan `scope.permissions` dari auth package, bukan hardcoded role check

### Implications untuk Feature Baru

- Komponen yang dibutuhkan di > 1 app harus masuk ke `packages/ui`
- Hook yang mengakses API harus menggunakan scope dari `useScopeStore()` — tidak boleh hardcode tenant ID
- Feature yang memerlukan offline (misal: attendance di portal) harus diimplementasi dengan service worker strategy yang sesuai
- Perubahan di `packages/api-client/` (karena backend schema berubah) akan memerlukan rebuild semua apps yang menggunakannya — koordinasikan dengan backend

### Implications untuk Performance

- `app-portal` targeting jaringan 3G: setiap dependency baru harus dievaluasi ukurannya
- `app-pos` harus Time-to-Interactive < 2 detik di hardware sederhana — tidak boleh mengimpor chart library atau rich text editor
- Service Worker PWA di `app-portal` harus mengcache rapor dan absensi untuk akses offline
