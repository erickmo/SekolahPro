import { Link } from 'react-router-dom'
import styles from './ErrorPage.module.css'

export default function NotFoundPage() {
  return (
    <div className={styles.root}>
      <div className={styles.code}>404</div>
      <h1 className={styles.title}>Halaman Tidak Ditemukan</h1>
      <p className={styles.message}>
        Halaman yang Anda cari tidak ada atau telah dipindahkan.
      </p>
      <Link to="/dashboard" className={styles.backBtn}>
        Kembali ke Dashboard
      </Link>
    </div>
  )
}
