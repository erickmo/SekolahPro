import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@/__ui_tests__/test-utils'
import { http, HttpResponse } from 'msw'
import { server } from '@/__ui_tests__/mocks/server'
import ProdukAkadPage from '@/pages/Koperasi/ProdukAkadPage'

const BASE_URL = 'http://localhost:8080'

const mockProdukAkadList = {
  data: [
    {
      id: 'pa-1',
      kode: 'PA-001',
      nama_produk: 'Murabahah Mobil',
      type: 'pembiayaan',
      akad_type: 'murabahah',
      deskripsi: 'Pembiayaan mobil',
      status: 'active',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    },
    {
      id: 'pa-2',
      kode: 'PA-002',
      nama_produk: 'Tabungan Pendidikan',
      type: 'tabungan',
      akad_type: 'wadiyah',
      deskripsi: 'Tabungan pendidikan',
      status: 'draft',
      created_at: '2026-01-02T00:00:00Z',
      updated_at: '2026-01-02T00:00:00Z',
    },
  ],
  meta: {
    total: 2,
    page: 1,
    per_page: 20,
    total_pages: 1,
  },
}

const emptyList = {
  data: [],
  meta: { total: 0, page: 1, per_page: 20, total_pages: 0 },
}

describe('ProdukAkadPage', () => {
  beforeEach(() => {
    server.resetHandlers()
  })

  it('renders page title correctly', () => {
    server.use(
      http.get(`${BASE_URL}/produk_akad`, () => HttpResponse.json(emptyList)),
    )
    render(<ProdukAkadPage />)
    expect(screen.getByText('Produk Akad')).toBeTruthy()
  })

  it('renders table headers', () => {
    server.use(
      http.get(`${BASE_URL}/produk_akad`, () => HttpResponse.json(emptyList)),
    )
    render(<ProdukAkadPage />)
    expect(screen.getByText('Kode')).toBeTruthy()
    expect(screen.getByText('Nama Produk')).toBeTruthy()
    expect(screen.getByText('Type')).toBeTruthy()
    expect(screen.getByText('Akad Type')).toBeTruthy()
    expect(screen.getByText('Status')).toBeTruthy()
    expect(screen.getByText('Aksi')).toBeTruthy()
  })

  it('shows loading skeleton initially', () => {
    server.use(
      http.get(`${BASE_URL}/produk_akad`, async () => {
        await new Promise((resolve) => setTimeout(resolve, 5000))
        return HttpResponse.json(mockProdukAkadList)
      }),
    )
    render(<ProdukAkadPage />)
    // Skeleton rows should be present (pulse animation class)
    const skeletons = document.querySelectorAll('[class*="skeleton"]')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('shows empty state when no data', async () => {
    server.use(
      http.get(`${BASE_URL}/produk_akad`, () => HttpResponse.json(emptyList)),
    )
    render(<ProdukAkadPage />)
    await waitFor(() => {
      expect(screen.getByText('Tidak ada data')).toBeTruthy()
    })
    expect(screen.getByText('Belum ada produk akad yang terdaftar.')).toBeTruthy()
  })

  it('renders data rows when data provided', async () => {
    server.use(
      http.get(`${BASE_URL}/produk_akad`, () => HttpResponse.json(mockProdukAkadList)),
    )
    render(<ProdukAkadPage />)
    await waitFor(() => {
      expect(screen.getByText('PA-001')).toBeTruthy()
    })
    expect(screen.getByText('Murabahah Mobil')).toBeTruthy()
    expect(screen.getByText('Pembiayaan')).toBeTruthy()
    expect(screen.getByText('Murabahah')).toBeTruthy()
    expect(screen.getByText('Aktif')).toBeTruthy()
    expect(screen.getByText('PA-002')).toBeTruthy()
    expect(screen.getByText('Tabungan Pendidikan')).toBeTruthy()
  })
})
