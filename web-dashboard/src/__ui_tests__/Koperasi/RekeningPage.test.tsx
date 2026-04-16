import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@/__ui_tests__/test-utils'
import { http, HttpResponse } from 'msw'
import { server } from '@/__ui_tests__/mocks/server'
import RekeningPage from '@/pages/Koperasi/RekeningPage'

const BASE_URL = 'http://localhost:8080'

const mockRekeningList = {
  data: [
    {
      id: 'rek-1',
      no_rekening: 'REK-001',
      nasabah_id: 'nsb-1',
      produk_akad_id: 'pa-1',
      saldo: 15000000,
      status: 'active',
      _data: {
        nasabah_nama: 'Ahmad Suryadi',
        produk_akad_nama: 'Murabahah Mobil',
      },
      created_at: '2026-01-20T00:00:00Z',
      updated_at: '2026-01-20T00:00:00Z',
    },
    {
      id: 'rek-2',
      no_rekening: 'REK-002',
      nasabah_id: 'nsb-2',
      produk_akad_id: 'pa-2',
      saldo: 500000,
      status: 'pending_open',
      _data: {
        nasabah_nama: 'PT Maju Bersama',
        produk_akad_nama: 'Tabungan Pendidikan',
      },
      created_at: '2026-02-01T00:00:00Z',
      updated_at: '2026-02-01T00:00:00Z',
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

describe('RekeningPage', () => {
  beforeEach(() => {
    server.resetHandlers()
  })

  it('renders page title correctly', () => {
    server.use(
      http.get(`${BASE_URL}/rekening`, () => HttpResponse.json(emptyList)),
    )
    render(<RekeningPage />)
    expect(screen.getByText('Rekening')).toBeTruthy()
  })

  it('renders table headers', () => {
    server.use(
      http.get(`${BASE_URL}/rekening`, () => HttpResponse.json(emptyList)),
    )
    render(<RekeningPage />)
    expect(screen.getByText('No. Rekening')).toBeTruthy()
    expect(screen.getByText('Nasabah')).toBeTruthy()
    expect(screen.getByText('Produk')).toBeTruthy()
    expect(screen.getByText('Saldo')).toBeTruthy()
    expect(screen.getByText('Status')).toBeTruthy()
    expect(screen.getByText('Aksi')).toBeTruthy()
  })

  it('shows loading skeleton initially', () => {
    server.use(
      http.get(`${BASE_URL}/rekening`, async () => {
        await new Promise((resolve) => setTimeout(resolve, 5000))
        return HttpResponse.json(mockRekeningList)
      }),
    )
    render(<RekeningPage />)
    const skeletons = document.querySelectorAll('[class*="skeleton"]')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('shows empty state when no data', async () => {
    server.use(
      http.get(`${BASE_URL}/rekening`, () => HttpResponse.json(emptyList)),
    )
    render(<RekeningPage />)
    await waitFor(() => {
      expect(screen.getByText('Tidak ada data')).toBeTruthy()
    })
    expect(screen.getByText('Belum ada rekening yang terdaftar.')).toBeTruthy()
  })

  it('renders data rows when data provided', async () => {
    server.use(
      http.get(`${BASE_URL}/rekening`, () => HttpResponse.json(mockRekeningList)),
    )
    render(<RekeningPage />)
    await waitFor(() => {
      expect(screen.getByText('REK-001')).toBeTruthy()
    })
    // Nasabah name from _data
    expect(screen.getByText('Ahmad Suryadi')).toBeTruthy()
    // Produk name from _data
    expect(screen.getByText('Murabahah Mobil')).toBeTruthy()
    // Saldo formatted as Rupiah
    expect(screen.getByText('Rp15.000.000')).toBeTruthy()
    // Status badge
    expect(screen.getByText('Aktif')).toBeTruthy()
    // Second row
    expect(screen.getByText('REK-002')).toBeTruthy()
    expect(screen.getByText('PT Maju Bersama')).toBeTruthy()
  })
})
