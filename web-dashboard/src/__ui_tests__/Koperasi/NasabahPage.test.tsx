import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@/__ui_tests__/test-utils'
import { http, HttpResponse } from 'msw'
import { server } from '@/__ui_tests__/mocks/server'
import NasabahPage from '@/pages/Koperasi/NasabahPage'

const BASE_URL = 'http://localhost:8080'

const mockNasabahList = {
  data: [
    {
      id: 'nsb-1',
      no_nasabah: 'NSB-001',
      nama_lengkap: 'Ahmad Suryadi',
      nik: '3201012345670001',
      type: 'perorangan',
      status: 'active',
      tanggal_daftar: '2026-01-15',
      created_at: '2026-01-15T00:00:00Z',
      updated_at: '2026-01-15T00:00:00Z',
    },
    {
      id: 'nsb-2',
      no_nasabah: 'NSB-002',
      nama_lengkap: 'PT Maju Bersama',
      nik: '3201012345670002',
      type: 'badan_usaha',
      status: 'pending',
      tanggal_daftar: '2026-02-10',
      created_at: '2026-02-10T00:00:00Z',
      updated_at: '2026-02-10T00:00:00Z',
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

describe('NasabahPage', () => {
  beforeEach(() => {
    server.resetHandlers()
  })

  it('renders page title correctly', () => {
    server.use(
      http.get(`${BASE_URL}/nasabah`, () => HttpResponse.json(emptyList)),
    )
    render(<NasabahPage />)
    expect(screen.getByText('Nasabah')).toBeTruthy()
  })

  it('renders table headers', () => {
    server.use(
      http.get(`${BASE_URL}/nasabah`, () => HttpResponse.json(emptyList)),
    )
    render(<NasabahPage />)
    expect(screen.getByText('No. Nasabah')).toBeTruthy()
    expect(screen.getByText('Nama Lengkap')).toBeTruthy()
    expect(screen.getByText('NIK')).toBeTruthy()
    expect(screen.getByText('Type')).toBeTruthy()
    expect(screen.getByText('Status')).toBeTruthy()
    expect(screen.getByText('Tgl Daftar')).toBeTruthy()
    expect(screen.getByText('Aksi')).toBeTruthy()
  })

  it('shows loading skeleton initially', () => {
    server.use(
      http.get(`${BASE_URL}/nasabah`, async () => {
        await new Promise((resolve) => setTimeout(resolve, 5000))
        return HttpResponse.json(mockNasabahList)
      }),
    )
    render(<NasabahPage />)
    const skeletons = document.querySelectorAll('[class*="skeleton"]')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('shows empty state when no data', async () => {
    server.use(
      http.get(`${BASE_URL}/nasabah`, () => HttpResponse.json(emptyList)),
    )
    render(<NasabahPage />)
    await waitFor(() => {
      expect(screen.getByText('Tidak ada data')).toBeTruthy()
    })
    expect(screen.getByText('Belum ada nasabah yang terdaftar.')).toBeTruthy()
  })

  it('renders data rows when data provided', async () => {
    server.use(
      http.get(`${BASE_URL}/nasabah`, () => HttpResponse.json(mockNasabahList)),
    )
    render(<NasabahPage />)
    await waitFor(() => {
      expect(screen.getByText('NSB-001')).toBeTruthy()
    })
    expect(screen.getByText('Ahmad Suryadi')).toBeTruthy()
    // NIK should be masked — only last 4 digits visible
    expect(screen.getByText('**********0001')).toBeTruthy()
    expect(screen.getByText('Perorangan')).toBeTruthy()
    expect(screen.getByText('Aktif')).toBeTruthy()
    expect(screen.getByText('PT Maju Bersama')).toBeTruthy()
  })
})
