# ADR-008: React 18 + Vite + CSS Modules (Tanpa Tailwind / UI Library)

**Status**: Accepted
**Date**: 2026-04-14
**Deciders**: Vernon Corp Engineering Team

## Context

Untuk frontend dashboard, tim perlu memilih:

1. **Build tool**: Webpack (Create React App), Vite, atau lainnya.
2. **State management**: Redux Toolkit, Zustand, Jotai, atau React Query saja.
3. **Styling approach**: Tailwind CSS, CSS-in-JS (Emotion/Styled Components), CSS Modules, atau UI library (MUI, Ant Design, Chakra).
4. **Data fetching**: SWR, TanStack Query, RTK Query, atau custom hooks.
5. **Testing**: Jest, Vitest, Playwright, Cypress.

### Business Requirements untuk Dashboard

- Dashboard admin multi-tenant: data tables, forms, charts, sidebar navigation
- Harus mendukung custom branding per tenant (warna, logo)
- Performa: First Contentful Paint < 1.5s, Bundle size < 300KB gzipped (initial load)
- Strict TypeScript — tidak ada `any`, semua props di-type
- Maintainability: developer baru harus bisa berkontribusi dalam < 1 hari orientasi

### Masalah yang Dihindari

**UI Library (MUI, Ant Design)**:
- Bundle besar: MUI core ~70KB gzipped, banyak yang di-tree-shake tapi ada overhead
- Opinionated design yang sulit di-override untuk custom branding
- Version lock: upgrade major version sering breaking
- Over-abstraction: simple button jadi component dengan 30+ props

**Tailwind CSS**:
- HTML menjadi verbose: `className="flex items-center justify-between px-4 py-2 rounded-lg bg-blue-500 hover:bg-blue-600 text-white font-semibold text-sm shadow-md"`
- Tidak ada component encapsulation — styling tersebar di markup
- PurgeCSS config perlu careful setup
- Colocation styling dengan logic ada, tapi readability turun

**CSS-in-JS (Styled Components, Emotion)**:
- Runtime overhead: setiap render bisa menghasilkan class names baru
- SSR complexity (jika dibutuhkan di masa depan)
- Bundle size lebih besar
- Debugging di DevTools lebih sulit (generated class names)

## Decision

Mengadopsi stack:
- **React 18** dengan Concurrent Features (Suspense, startTransition)
- **Vite** sebagai build tool + dev server
- **CSS Modules** sebagai styling approach
- **Zustand** untuk global state (light-weight, minimal boilerplate)
- **TanStack Query v5** untuk server state management
- **Vitest + MSW + Playwright** untuk testing

### Kenapa React 18 (bukan Next.js)

Dashboard ini adalah SPA (Single Page Application) murni — tidak membutuhkan SSR/SSG. Menambahkan Next.js overhead (routing conventions, server components complexity) tidak justified untuk use case ini. Jika kebutuhan berubah (SEO, landing page), bisa migrate ke Next.js di kemudian hari.

### Kenapa Vite (bukan Webpack/CRA)

```
Development server start time:
- Create React App (webpack): ~8-15 detik
- Vite:                       ~300-500ms

Hot Module Replacement (HMR):
- Webpack:  100-500ms per change
- Vite:     10-50ms per change (ESM native)

Production build time:
- Webpack: ~60-120 detik (100+ components)
- Vite:    ~15-30 detik (Rollup-based)
```

Vite menggunakan native ESM di development — tidak ada bundling, browser langsung load modules. Production menggunakan Rollup dengan code splitting otomatis.

### Kenapa CSS Modules

```tsx
// ProductCard.module.css
.card {
    background: var(--surface-primary);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--space-4);
    transition: box-shadow 0.2s ease;
}

.card:hover {
    box-shadow: var(--shadow-md);
}

.title {
    font-size: var(--text-lg);
    font-weight: var(--font-semibold);
    color: var(--text-primary);
    margin: 0 0 var(--space-2);
}

.price {
    font-size: var(--text-2xl);
    font-weight: var(--font-bold);
    color: var(--color-brand-primary);
}
```

```tsx
// ProductCard.tsx
import styles from './ProductCard.module.css';

interface ProductCardProps {
    name: string;
    sku: string;
    price: number;
    category: string;
}

export function ProductCard({ name, sku, price, category }: ProductCardProps) {
    return (
        <article className={styles.card}>
            <h3 className={styles.title}>{name}</h3>
            <span className={styles.sku}>{sku}</span>
            <div className={styles.price}>{formatCurrency(price)}</div>
            <span className={styles.category}>{category}</span>
        </article>
    );
}
```

CSS Modules memberikan:
- **Scoped styles**: `.card` di `ProductCard.module.css` tidak konflik dengan `.card` di `OrderCard.module.css`
- **Zero runtime overhead**: Semua class name hashing dilakukan saat build
- **TypeScript support**: `import styles from './X.module.css'` memberikan type `{ [key: string]: string }`
- **Full CSS power**: Media queries, pseudo-classes, CSS variables, nesting — semua tersedia
- **Colocated dengan component**: File `.module.css` di samping `.tsx` yang sama

### Design Token System via CSS Variables

```css
/* src/styles/tokens.css */
:root {
    /* Color Palette */
    --color-brand-primary:    #2563EB;
    --color-brand-secondary:  #0EA5E9;
    
    /* Semantic Colors */
    --text-primary:           #111827;
    --text-secondary:         #6B7280;
    --surface-primary:        #FFFFFF;
    --surface-secondary:      #F9FAFB;
    --border-subtle:          #E5E7EB;
    
    /* Typography */
    --text-sm:     0.875rem;
    --text-base:   1rem;
    --text-lg:     1.125rem;
    --text-xl:     1.25rem;
    --text-2xl:    1.5rem;
    --font-normal:    400;
    --font-semibold:  600;
    --font-bold:      700;
    
    /* Spacing */
    --space-1:    0.25rem;
    --space-2:    0.5rem;
    --space-3:    0.75rem;
    --space-4:    1rem;
    --space-6:    1.5rem;
    --space-8:    2rem;
    
    /* Border Radius */
    --radius-sm:  4px;
    --radius-md:  8px;
    --radius-lg:  12px;
    --radius-xl:  16px;
    
    /* Shadows */
    --shadow-sm:  0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow-md:  0 4px 6px -1px rgb(0 0 0 / 0.1);
    --shadow-lg:  0 10px 15px -3px rgb(0 0 0 / 0.1);
}

/* Per-tenant branding override */
[data-tenant="vernoncorp"] {
    --color-brand-primary:   #1E40AF;
    --color-brand-secondary: #3B82F6;
}
```

### State Management Architecture

```
State Layer:
├── Server State    → TanStack Query v5
│   ├── caching, background refetch, pagination
│   └── optimistic updates
├── Global UI State → Zustand
│   ├── sidebar open/close
│   ├── selected tenant/branch scope
│   └── notification queue
└── Local State     → useState / useReducer
    ├── form state
    └── component-specific UI state
```

```typescript
// store/scope.store.ts — Zustand
interface ScopeState {
    tenantID:    string | null;
    companyID:   string | null;
    branchID:    string | null;
    warehouseID: string | null;
    setScope: (scope: Partial<ScopeState>) => void;
    clearScope: () => void;
}

export const useScopeStore = create<ScopeState>((set) => ({
    tenantID:    null,
    companyID:   null,
    branchID:    null,
    warehouseID: null,
    setScope: (scope) => set((state) => ({ ...state, ...scope })),
    clearScope: () => set({ tenantID: null, companyID: null, branchID: null, warehouseID: null }),
}));
```

```typescript
// hooks/useProducts.ts — TanStack Query
export function useProducts(filters: ProductFilters) {
    const scope = useScopeStore();
    
    return useQuery({
        queryKey: ['products', scope.tenantID, scope.warehouseID, filters],
        queryFn: () => productApi.list({ ...filters, ...scope }),
        staleTime: 30_000,  // 30 detik — tidak refetch terus-menerus
        enabled: !!scope.tenantID,  // hanya fetch jika ada scope
    });
}
```

### Testing Stack

```
Unit Tests (Vitest):
├── Pure functions, utilities, formatters
├── Custom hooks (renderHook)
└── Component render + interaction

API Mocking (MSW v2):
├── Mock service worker intercepts fetch di test
├── Handler definitions colocated dengan feature
└── Same handlers bisa digunakan di Storybook

E2E Tests (Playwright):
├── Critical user journeys (login, create product, place order)
├── Multi-browser (Chrome, Firefox, Safari)
└── Visual regression (screenshots)
```

```typescript
// tests/ProductCard.test.tsx
import { render, screen } from '@testing-library/react';
import { ProductCard } from './ProductCard';

describe('ProductCard', () => {
    it('displays product name and price', () => {
        render(<ProductCard name="Laptop" sku="LPT-001" price={15000000} category="Electronics" />);
        
        expect(screen.getByText('Laptop')).toBeInTheDocument();
        expect(screen.getByText('Rp 15.000.000')).toBeInTheDocument();
    });
});
```

### Vite Config

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
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
                    'vendor-charts': ['recharts'],
                },
            },
        },
    },
});
```

## Consequences

### Positive

- **Bundle size terkontrol**: Tanpa MUI atau Ant Design, initial bundle ~120KB gzipped. Jauh di bawah 300KB target.
- **Custom branding mudah**: CSS variables di-override per tenant di `:root` — tidak ada !important wars atau JS theming complexity.
- **Developer velocity tinggi**: Vite HMR < 50ms — feedback loop sangat cepat saat development.
- **Type safety end-to-end**: Strict TypeScript dari API response (generated dari OpenAPI spec) hingga ke React props.
- **CSS Modules isolation**: Tidak ada class name conflict meskipun ada ratusan components.
- **Test yang cepat**: Vitest menggunakan Vite — test suite ratusan test selesai dalam < 10 detik.

### Negative / Trade-offs

- **Lebih banyak CSS ditulis manual**: Tanpa Tailwind atau UI library, developer perlu menulis CSS untuk setiap component. Estimasi +30-40% waktu untuk component baru vs Tailwind.
- **Tidak ada pre-built components**: Table, Modal, DatePicker, Select harus dibangun atau dipilih library spesifik (headless: Radix UI, React Aria).
- **Design consistency perlu enforced**: Tanpa Tailwind's constrained scale, developer bisa pakai nilai arbitrary (`padding: 13px`). Perlu design review process.
- **CSS specificity issues**: Meskipun CSS Modules meng-scope class, global styles dari `index.css` bisa masih konflik.
- **Learning curve CSS Modules**: Developer yang terbiasa Tailwind perlu adjustment ke cara berpikir colocated CSS files.

## Alternatives Considered

### 1. Next.js + MUI
- Full-featured, great DX, SSR support
- Ditolak karena SSR tidak dibutuhkan untuk dashboard SPA, MUI bundle terlalu besar, dan design customization lebih sulit.

### 2. Vite + Tailwind CSS
- DX sangat baik, utility-first cepat untuk prototyping
- Ditolak karena HTML readability turun drastis dengan banyak utility classes, sulit enforce design consistency, dan tidak colocated dengan component logic.

### 3. Vite + Styled Components (CSS-in-JS)
- Component-scoped styling, dynamic styling mudah
- Ditolak karena runtime overhead di setiap render, SSR complexity jika dibutuhkan nanti, dan debugging class names yang di-generate lebih sulit.

### 4. Redux Toolkit (bukan Zustand)
- Industry standard, Redux DevTools excellent
- Ditolak karena boilerplate lebih banyak (slice, action, selector) untuk use case yang relatif sederhana (scope switcher, UI state). Zustand memberikan 80% manfaat dengan 20% boilerplate.

### 5. SWR (bukan TanStack Query)
- Lebih ringan dari TanStack Query
- Ditolak karena TanStack Query v5 memberikan lebih banyak fitur yang dibutuhkan: infinite queries untuk data tables, optimistic updates yang lebih robust, dan better TypeScript inference.
