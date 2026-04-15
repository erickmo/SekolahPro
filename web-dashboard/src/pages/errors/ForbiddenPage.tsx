import { Link } from 'react-router-dom'
import styles from './ErrorPage.module.css'

export default function ForbiddenPage() {
  return (
    <div className={styles.root}>
      <div className={styles.code}>403</div>
      <h1 className={styles.title}>Akses Ditolak</h1>
      <p className={styles.message}>
        Anda tidak memiliki izin untuk mengakses halaman ini.
      </p>
      <Link to="/dashboard" className={styles.backBtn}>
        Kembali ke Dashboard
      </Link>
    </div>
  )
}
