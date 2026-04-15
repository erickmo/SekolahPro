import { PageHeader } from '@/layouts/PageHeader/PageHeader'
import { useAuthStore } from '@/stores/auth.store'
import styles from './DashboardPage.module.css'

// ─── Stat card component ──────────────────────────────────────────────────────

interface StatCardProps {
  label: string
  value: string | number
  change?: string
  trend?: 'up' | 'down' | 'neutral'
}

function StatCard({ label, value, change, trend }: StatCardProps) {
  return (
    <div className={styles.statCard}>
      <p className={styles.statLabel}>{label}</p>
      <p className={styles.statValue}>{value}</p>
      {change && (
        <p className={styles[`statTrend_${trend ?? 'neutral'}`]}>
          {change}
        </p>
      )}
    </div>
  )
}

// ─── Dashboard page ───────────────────────────────────────────────────────────

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user)

  return (
    <div className="animate-page-in">
      <PageHeader
        title={`Selamat Datang, ${user?.name ?? 'User'}`}
        subtitle="Ini adalah halaman dashboard utama aplikasi Anda."
      />

      {/* Stats grid */}
      <div className={styles.statsGrid}>
        <StatCard label="Total Pengguna" value="1,240" change="+12% dari bulan lalu" trend="up" />
        <StatCard label="Aktif Hari Ini" value="84" change="+5% dari kemarin" trend="up" />
        <StatCard label="Tugas Selesai" value="320" change="-3% dari minggu lalu" trend="down" />
        <StatCard label="Pendapatan" value="Rp 48,2jt" change="+8% dari bulan lalu" trend="up" />
      </div>

      {/* Placeholder content */}
      <div className={styles.contentGrid}>
        <div className={styles.card}>
          <h2 className={styles.cardTitle}>Aktivitas Terbaru</h2>
          <p className={styles.cardPlaceholder}>
            Ganti bagian ini dengan komponen list, chart, atau table yang relevan.
          </p>
        </div>
        <div className={styles.card}>
          <h2 className={styles.cardTitle}>Ringkasan</h2>
          <p className={styles.cardPlaceholder}>
            Ganti bagian ini dengan chart atau statistik ringkasan.
          </p>
        </div>
      </div>
    </div>
  )
}
