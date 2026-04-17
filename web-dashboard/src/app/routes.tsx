import { createBrowserRouter } from 'react-router-dom'
import { lazy, Suspense } from 'react'
import { AppShell } from '@/layouts/AppShell/AppShell'
import { appConfig } from '@/config/app.config'
import {
  RootRedirect,
  AuthRoute,
  GuestRoute,
  SuperuserRoute,
  GroupRoute,
  CompanyRoute,
} from './ProtectedRoute'

// ─── Lazy-loaded pages ────────────────────────────────────────────────────────

const LoginPage          = lazy(() => import('@/pages/Login/LoginPage'))
const DashboardPage      = lazy(() => import('@/pages/Dashboard/DashboardPage'))
const ChooseCompanyPage  = lazy(() => import('@/pages/ChooseCompany/ChooseCompanyPage'))
const AcademicYearsPage  = lazy(() => import('@/pages/AcademicYears/AcademicYearsPage'))
const TeachersPage       = lazy(() => import('@/pages/Teachers/TeachersPage'))
const ClassRoomsPage     = lazy(() => import('@/pages/ClassRooms/ClassRoomsPage'))
const UsersPage          = lazy(() => import('@/pages/Users/UsersPage'))
const ProdukAkadPage     = lazy(() => import('@/pages/Koperasi/ProdukAkadPage'))
const NasabahPage        = lazy(() => import('@/pages/Koperasi/NasabahPage'))
const RekeningPage       = lazy(() => import('@/pages/Koperasi/RekeningPage'))
const NotFoundPage       = lazy(() => import('@/pages/errors/NotFoundPage'))
const ForbiddenPage      = lazy(() => import('@/pages/errors/ForbiddenPage'))

// ─── Phase 4 pages ────────────────────────────────────────────────────────────
const StudentsPage            = lazy(() => import('@/pages/Students/StudentsPage'))
const StudentAdmissionsPage   = lazy(() => import('@/pages/Students/StudentAdmissionsPage'))
const StudentClassPlacementsPage = lazy(() => import('@/pages/Students/StudentClassPlacementsPage'))
const SimpananPokokWajibPage  = lazy(() => import('@/pages/Koperasi/SimpananPokokWajibPage'))
const TabunganPage            = lazy(() => import('@/pages/Koperasi/TabunganPage'))
const DepositoPage            = lazy(() => import('@/pages/Koperasi/DepositoPage'))
const DendaPage               = lazy(() => import('@/pages/Koperasi/DendaPage'))
const JaminanPage             = lazy(() => import('@/pages/Koperasi/JaminanPage'))

// ─── Phase 5 pages ────────────────────────────────────────────────────────────
const CurriculumPage          = lazy(() => import('@/pages/Academic/CurriculumPage'))
const SubjectPage             = lazy(() => import('@/pages/Academic/SubjectPage'))
const AcademicCalendarPage    = lazy(() => import('@/pages/Academic/AcademicCalendarPage'))
const TeachingSchedulePage    = lazy(() => import('@/pages/Academic/TeachingSchedulePage'))
const LessonPlanPage          = lazy(() => import('@/pages/Academic/LessonPlanPage'))
const TeachingJournalPage     = lazy(() => import('@/pages/Academic/TeachingJournalPage'))

function S({ children }: { children: React.ReactNode }) {
  return <Suspense fallback={<div />}>{children}</Suspense>
}

// ─── Single-tenant routes ─────────────────────────────────────────────────────
// Active when VITE_MULTI_TENANT=false (default)

const singleTenantRoutes = [
  {
    path: '/',
    element: <AuthRoute><AppShell /></AuthRoute>,
    children: [
      { path: 'dashboard', element: <S><DashboardPage /></S> },
      { path: 'users', element: <S><UsersPage /></S> },
      { path: 'academic-years', element: <S><AcademicYearsPage /></S> },
      { path: 'teachers', element: <S><TeachersPage /></S> },
      { path: 'class-rooms', element: <S><ClassRoomsPage /></S> },
      { path: 'koperasi/produk-akad', element: <S><ProdukAkadPage /></S> },
      { path: 'koperasi/nasabah', element: <S><NasabahPage /></S> },
      { path: 'koperasi/rekening', element: <S><RekeningPage /></S> },
      { path: 'koperasi/simpanan-pokok-wajib', element: <S><SimpananPokokWajibPage /></S> },
      { path: 'koperasi/tabungan', element: <S><TabunganPage /></S> },
      { path: 'koperasi/deposito', element: <S><DepositoPage /></S> },
      { path: 'koperasi/denda', element: <S><DendaPage /></S> },
      { path: 'koperasi/jaminan', element: <S><JaminanPage /></S> },
      { path: 'akademik/curriculum', element: <S><CurriculumPage /></S> },
      { path: 'akademik/subjects', element: <S><SubjectPage /></S> },
      { path: 'akademik/academic-calendar', element: <S><AcademicCalendarPage /></S> },
      { path: 'akademik/teaching-schedule', element: <S><TeachingSchedulePage /></S> },
      { path: 'akademik/lesson-plan', element: <S><LessonPlanPage /></S> },
      { path: 'akademik/teaching-journal', element: <S><TeachingJournalPage /></S> },
      { path: 'students', element: <S><StudentsPage /></S> },
      { path: 'students/admissions', element: <S><StudentAdmissionsPage /></S> },
      { path: 'students/class-placements', element: <S><StudentClassPlacementsPage /></S> },
    ],
  },
]

// ─── Multi-tenant routes ──────────────────────────────────────────────────────
// Active when VITE_MULTI_TENANT=true

const multiTenantRoutes = [
  // Company/group selection page
  {
    path: '/choose-company',
    element: <AuthRoute><S><ChooseCompanyPage /></S></AuthRoute>,
  },

  // ── Superuser context: /su/* ─────────────────────────────────────────────────
  {
    path: '/su',
    element: <SuperuserRoute><AppShell context="superuser" /></SuperuserRoute>,
    children: [
      { path: 'dashboard', element: <S><DashboardPage /></S> },
      { path: 'users', element: <S><UsersPage /></S> },
      // { path: 'tenants',   element: <S><TenantsListPage /></S> },
      // { path: 'companies', element: <S><CompaniesListPage /></S> },
    ],
  },

  // ── HQ / Group context: /g/* ─────────────────────────────────────────────────
  {
    path: '/g',
    element: <GroupRoute><AppShell context="hq" /></GroupRoute>,
    children: [
      { path: 'dashboard', element: <S><DashboardPage /></S> },
      // { path: 'reports',  element: <S><HQReportsPage /></S> },
    ],
  },

  // ── Company context: /c/:companyCode/* ───────────────────────────────────────
  {
    path: '/c/:companyCode',
    element: <CompanyRoute><AppShell context="company" /></CompanyRoute>,
    children: [
      { path: 'dashboard', element: <S><DashboardPage /></S> },
      { path: 'users', element: <S><UsersPage /></S> },
      { path: 'academic-years', element: <S><AcademicYearsPage /></S> },
      { path: 'teachers', element: <S><TeachersPage /></S> },
      { path: 'class-rooms', element: <S><ClassRoomsPage /></S> },
      { path: 'koperasi/produk-akad', element: <S><ProdukAkadPage /></S> },
      { path: 'koperasi/nasabah', element: <S><NasabahPage /></S> },
      { path: 'koperasi/rekening', element: <S><RekeningPage /></S> },
      { path: 'koperasi/simpanan-pokok-wajib', element: <S><SimpananPokokWajibPage /></S> },
      { path: 'koperasi/tabungan', element: <S><TabunganPage /></S> },
      { path: 'koperasi/deposito', element: <S><DepositoPage /></S> },
      { path: 'koperasi/denda', element: <S><DendaPage /></S> },
      { path: 'koperasi/jaminan', element: <S><JaminanPage /></S> },
      { path: 'akademik/curriculum', element: <S><CurriculumPage /></S> },
      { path: 'akademik/subjects', element: <S><SubjectPage /></S> },
      { path: 'akademik/academic-calendar', element: <S><AcademicCalendarPage /></S> },
      { path: 'akademik/teaching-schedule', element: <S><TeachingSchedulePage /></S> },
      { path: 'akademik/lesson-plan', element: <S><LessonPlanPage /></S> },
      { path: 'akademik/teaching-journal', element: <S><TeachingJournalPage /></S> },
      { path: 'students', element: <S><StudentsPage /></S> },
      { path: 'students/admissions', element: <S><StudentAdmissionsPage /></S> },
      { path: 'students/class-placements', element: <S><StudentClassPlacementsPage /></S> },
    ],
  },
]

// ─── Router ───────────────────────────────────────────────────────────────────

export const router = createBrowserRouter([
  // Root — redirect based on auth state + tenant mode
  { path: '/', element: <RootRedirect /> },

  // Login (guest only)
  { path: '/login', element: <GuestRoute><S><LoginPage /></S></GuestRoute> },

  // Active route set based on tenant mode
  ...(appConfig.isMultiTenant ? multiTenantRoutes : singleTenantRoutes),

  // Error pages
  { path: '/403', element: <S><ForbiddenPage /></S> },
  { path: '*',    element: <S><NotFoundPage /></S> },
])
