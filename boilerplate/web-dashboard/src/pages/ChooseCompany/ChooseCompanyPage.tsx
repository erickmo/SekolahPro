import { useNavigate } from 'react-router-dom'
import { Building2, ChevronRight, LogOut } from 'lucide-react'
import { useAuthStore } from '@/stores/auth.store'
import type { Company, CompanyGroup } from '@/types/auth.types'
import styles from './ChooseCompanyPage.module.css'

export default function ChooseCompanyPage() {
  const navigate = useNavigate()
  const { user, availableGroups, selectGroup, selectCompany, logout } = useAuthStore()

  function handleSelectCompany(group: CompanyGroup, company: Company) {
    selectGroup(group)
    selectCompany(company)
    navigate(`/c/${company.code}/dashboard`, { replace: true })
  }

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <div className={styles.root}>
      <div className={styles.card}>
        {/* Header */}
        <div className={styles.header}>
          <Building2 size={28} className={styles.headerIcon} />
          <div>
            <h1 className={styles.title}>Pilih Perusahaan</h1>
            <p className={styles.subtitle}>
              Selamat datang, <strong>{user?.name}</strong>. Pilih perusahaan untuk melanjutkan.
            </p>
          </div>
        </div>

        {/* Company groups list */}
        <div className={styles.groups}>
          {availableGroups.length === 0 && (
            <p className={styles.empty}>
              Tidak ada perusahaan yang tersedia. Hubungi administrator.
            </p>
          )}
          {availableGroups.map((group) => (
            <div key={group.id} className={styles.group}>
              {availableGroups.length > 1 && (
                <p className={styles.groupLabel}>{group.name}</p>
              )}
              <ul className={styles.companyList}>
                {group.companies.map((company) => (
                  <li key={company.id}>
                    <button
                      className={styles.companyBtn}
                      onClick={() => handleSelectCompany(group, company)}
                    >
                      <span className={styles.companyAvatar}>
                        {company.name[0]?.toUpperCase() ?? 'C'}
                      </span>
                      <div className={styles.companyInfo}>
                        <span className={styles.companyName}>{company.name}</span>
                        <span className={styles.companyCode}>{company.code}</span>
                      </div>
                      <ChevronRight size={16} className={styles.companyChevron} />
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Logout link */}
        <button className={styles.logoutBtn} onClick={handleLogout}>
          <LogOut size={14} />
          Keluar dari akun
        </button>
      </div>
    </div>
  )
}
