import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@/__ui_tests__/test-utils'
import { http, HttpResponse } from 'msw'
import { server } from '@/__ui_tests__/mocks/server'
import UsersPage from '@/pages/Users/UsersPage'

const BASE_URL = 'http://localhost:8080'

const mockUserList = {
  data: [
    {
      id: 'usr-1',
      email: 'admin@sekolah.id',
      full_name: 'Administrator Sekolah',
      is_active: true,
      is_superadmin: false,
      phone: '081234567890',
      roles: ['admin'],
      created_at: '2026-01-15T08:00:00Z',
      updated_at: '2026-01-15T08:00:00Z',
    },
    {
      id: 'usr-2',
      email: 'guru@sekolah.id',
      full_name: 'Guru Matematika',
      is_active: true,
      is_superadmin: false,
      phone: '081298765432',
      roles: ['user'],
      created_at: '2026-02-01T10:30:00Z',
      updated_at: '2026-02-01T10:30:00Z',
    },
    {
      id: 'usr-3',
      email: 'viewer@sekolah.id',
      full_name: 'Peninjau Data',
      is_active: false,
      is_superadmin: false,
      phone: '',
      roles: ['viewer'],
      created_at: '2026-03-01T12:00:00Z',
      updated_at: '2026-03-01T12:00:00Z',
    },
  ],
  meta: {
    total: 3,
    page: 1,
    per_page: 20,
    total_pages: 1,
  },
}

const emptyList = {
  data: [],
  meta: { total: 0, page: 1, per_page: 20, total_pages: 0 },
}

describe('UsersPage', () => {
  beforeEach(() => {
    server.resetHandlers()
  })

  it('renders page title correctly', () => {
    server.use(
      http.get(`${BASE_URL}/api/users`, () => HttpResponse.json(emptyList)),
    )
    render(<UsersPage />)
    expect(screen.getByText('Pengguna')).toBeTruthy()
  })

  it('renders table headers', () => {
    server.use(
      http.get(`${BASE_URL}/api/users`, () => HttpResponse.json(emptyList)),
    )
    render(<UsersPage />)
    expect(screen.getByText('Email')).toBeTruthy()
    expect(screen.getByText('Nama Lengkap')).toBeTruthy()
    expect(screen.getByText('Role')).toBeTruthy()
    expect(screen.getByText('Status')).toBeTruthy()
    expect(screen.getByText('Telepon')).toBeTruthy()
    expect(screen.getByText('Tanggal Daftar')).toBeTruthy()
    expect(screen.getByText('Aksi')).toBeTruthy()
  })

  it('shows loading skeleton initially', () => {
    server.use(
      http.get(`${BASE_URL}/api/users`, async () => {
        await new Promise((resolve) => setTimeout(resolve, 5000))
        return HttpResponse.json(mockUserList)
      }),
    )
    render(<UsersPage />)
    const skeletons = document.querySelectorAll('[class*="skeleton"]')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('shows empty state when no data', async () => {
    server.use(
      http.get(`${BASE_URL}/api/users`, () => HttpResponse.json(emptyList)),
    )
    render(<UsersPage />)
    await waitFor(() => {
      expect(screen.getByText('Tidak ada data')).toBeTruthy()
    })
    expect(screen.getByText('Belum ada pengguna yang terdaftar.')).toBeTruthy()
  })

  it('renders data rows when data provided', async () => {
    server.use(
      http.get(`${BASE_URL}/api/users`, () => HttpResponse.json(mockUserList)),
    )
    render(<UsersPage />)
    await waitFor(() => {
      expect(screen.getByText('admin@sekolah.id')).toBeTruthy()
    })
    expect(screen.getByText('Administrator Sekolah')).toBeTruthy()
    expect(screen.getByText('Admin')).toBeTruthy()
    expect(screen.getByText('Aktif')).toBeTruthy()
    expect(screen.getByText('guru@sekolah.id')).toBeTruthy()
    expect(screen.getByText('Guru Matematika')).toBeTruthy()
    expect(screen.getByText('Peninjau')).toBeTruthy()
    expect(screen.getByText('viewer@sekolah.id')).toBeTruthy()
    expect(screen.getByText('Nonaktif')).toBeTruthy()
  })
})
