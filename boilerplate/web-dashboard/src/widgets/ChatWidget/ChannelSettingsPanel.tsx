import { useState, useEffect } from 'react'
import { ArrowLeft, RefreshCw, Trash2 } from 'lucide-react'
import { chatService } from '@/services/chat.service'
import { useChatStore } from '@/stores/chat.store'
import { toast } from '@/widgets/Toast/Toast'
import type { ChatMemberAPI, ChatRuleAPI } from '@/services/chat.service'
import type { ChatChannel } from './chat.types'
import styles from './ChannelSettingsPanel.module.css'

// ─── Types ──────────────────────────────────────────────────────────────────────

type Tab = 'members' | 'rules' | 'info'

interface ChannelSettingsPanelProps {
  channel: ChatChannel
}

// ─── Role Badge ─────────────────────────────────────────────────────────────────

function RoleBadge({ role }: { role: string }) {
  const cls =
    role === 'moderator'
      ? styles.roleModerator
      : role === 'readonly'
        ? styles.roleReadonly
        : styles.roleMember
  return (
    <span className={`${styles.roleBadge} ${cls}`}>
      {role === 'moderator' ? 'Mod' : role === 'readonly' ? 'View' : 'Member'}
    </span>
  )
}

// ─── Members Tab ─────────────────────────────────────────────────────────────────

function MembersTab({ channelId }: { channelId: string }) {
  const [members, setMembers] = useState<ChatMemberAPI[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [changingPerm, setChangingPerm] = useState<string | null>(null)
  const currentMemberRole = useChatStore((s) => s.currentMemberRole)
  const isModerator = currentMemberRole === 'moderator'

  useEffect(() => {
    loadMembers()
  }, [channelId])

  async function loadMembers() {
    setIsLoading(true)
    try {
      const result = await chatService.listMembers(channelId, { limit: 100 })
      setMembers(result.items)
    } catch {
      // Silent
    } finally {
      setIsLoading(false)
    }
  }

  async function handleChangePermission(userId: string, permission: string) {
    setChangingPerm(userId)
    try {
      await chatService.updateMemberPermission(channelId, userId, permission)
      toast.success('Permission diperbarui')
      await loadMembers()
    } catch {
      toast.error('Gagal mengubah permission')
    } finally {
      setChangingPerm(null)
    }
  }

  async function handleRemoveMember(userId: string, userName: string) {
    try {
      await chatService.removeMember(channelId, userId)
      toast.success(`${userName} dikeluarkan dari channel`)
      await loadMembers()
    } catch {
      toast.error('Gagal mengeluarkan member')
    }
  }

  if (isLoading) return <div className={styles.empty}>Memuat anggota...</div>
  if (members.length === 0) return <div className={styles.empty}>Belum ada anggota</div>

  return (
    <div className={styles.section}>
      <p className={styles.sectionTitle}>Anggota ({members.length})</p>
      {members.map((m) => (
        <div key={m.user_id} className={styles.memberItem}>
          <div className={styles.memberAvatar}>{m.inisial}</div>
          <div className={styles.memberInfo}>
            <span className={styles.memberName}>{m.nama}</span>
          </div>
          <RoleBadge role={m.peran} />
          {isModerator && (
            <div className={styles.memberActions}>
              <select
                className={styles.permSelect}
                value={m.peran}
                disabled={changingPerm === m.user_id}
                onChange={(e) => handleChangePermission(m.user_id, e.target.value)}
                title="Ubah permission"
              >
                <option value="moderator">Moderator</option>
                <option value="member">Member</option>
                <option value="readonly">Read-only</option>
              </select>
              <button
                className={styles.removeBtn}
                onClick={() => handleRemoveMember(m.user_id, m.nama)}
                title="Keluarkan dari channel"
              >
                <Trash2 size={12} />
              </button>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

// ─── Rules Tab ───────────────────────────────────────────────────────────────────

function RulesTab({ channelId }: { channelId: string }) {
  const [rules, setRules] = useState<ChatRuleAPI[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [newRuleType, setNewRuleType] = useState<'user' | 'role'>('user')
  const [newRuleValue, setNewRuleValue] = useState('')
  const [newRulePerm, setNewRulePerm] = useState('member')
  const [isAdding, setIsAdding] = useState(false)
  const currentMemberRole = useChatStore((s) => s.currentMemberRole)
  const isModerator = currentMemberRole === 'moderator'

  useEffect(() => {
    loadRules()
  }, [channelId])

  async function loadRules() {
    setIsLoading(true)
    try {
      const result = await chatService.listRules(channelId)
      setRules(result)
    } catch {
      // Silent
    } finally {
      setIsLoading(false)
    }
  }

  async function handleDeleteRule(ruleId: string) {
    try {
      await chatService.deleteRule(channelId, ruleId)
      toast.success('Aturan dihapus')
      await loadRules()
    } catch {
      toast.error('Gagal menghapus aturan')
    }
  }

  async function handleAddRule() {
    if (!newRuleValue.trim()) return
    setIsAdding(true)
    try {
      await chatService.addRule(channelId, {
        tipe_aturan: newRuleType,
        ...(newRuleType === 'user' ? { user_id: newRuleValue.trim() } : { role: newRuleValue.trim() }),
        permission: newRulePerm,
      })
      toast.success('Aturan ditambahkan')
      setShowAddForm(false)
      setNewRuleValue('')
      await loadRules()
    } catch {
      toast.error('Gagal menambahkan aturan')
    } finally {
      setIsAdding(false)
    }
  }

  return (
    <div className={styles.section}>
      <p className={styles.sectionTitle}>Aturan Akses ({rules.length})</p>

      {isLoading ? (
        <div className={styles.empty}>Memuat aturan...</div>
      ) : rules.length === 0 ? (
        <div className={styles.empty}>Belum ada aturan akses</div>
      ) : (
        rules.map((r) => (
          <div key={r.id} className={styles.ruleItem}>
            <div className={styles.ruleInfo}>
              <span className={styles.ruleType}>
                {r.tipe_aturan === 'user' ? 'User' : 'Role'}
              </span>
              <span className={styles.ruleDetail}>
                {r.tipe_aturan === 'user' ? r.user_id ?? '-' : r.role ?? '-'}
              </span>
            </div>
            <span className={styles.rulePerm}>{r.peran}</span>
            {isModerator && (
              <button className={styles.removeBtn} onClick={() => handleDeleteRule(r.id)} title="Hapus aturan">
                <Trash2 size={12} />
              </button>
            )}
          </div>
        ))
      )}

      {isModerator && (
        <>
          {showAddForm ? (
            <div className={styles.addRuleForm}>
              <div className={styles.addRuleRow}>
                <div className={styles.addRuleField}>
                  <span className={styles.addRuleLabel}>Tipe</span>
                  <select className={styles.addRuleSelect} value={newRuleType} onChange={(e) => setNewRuleType(e.target.value as 'user' | 'role')}>
                    <option value="user">User</option>
                    <option value="role">Role</option>
                  </select>
                </div>
                <div className={styles.addRuleField}>
                  <span className={styles.addRuleLabel}>{newRuleType === 'user' ? 'User ID' : 'Role'}</span>
                  <input className={styles.addRuleInput} value={newRuleValue} onChange={(e) => setNewRuleValue(e.target.value)} placeholder={newRuleType === 'user' ? 'User ID' : 'Nama role'} />
                </div>
                <div className={styles.addRuleField}>
                  <span className={styles.addRuleLabel}>Permission</span>
                  <select className={styles.addRuleSelect} value={newRulePerm} onChange={(e) => setNewRulePerm(e.target.value)}>
                    <option value="moderator">Moderator</option>
                    <option value="member">Member</option>
                    <option value="readonly">Read-only</option>
                  </select>
                </div>
              </div>
              <div className={styles.addRuleRow}>
                <button className={styles.addRuleBtn} onClick={handleAddRule} disabled={isAdding}>
                  {isAdding ? 'Menambahkan...' : 'Tambah'}
                </button>
                <button className={styles.addRuleBtn} style={{ background: 'var(--color-surface)', color: 'var(--color-text)', border: '1px solid var(--color-border)' }} onClick={() => setShowAddForm(false)}>
                  Batal
                </button>
              </div>
            </div>
          ) : (
            <button className={styles.syncBtn} onClick={() => setShowAddForm(true)}>
              + Tambah Aturan
            </button>
          )}
          <button className={styles.syncBtn} onClick={async () => {
            try {
              const count = await chatService.syncMembers(channelId)
              toast.success(`${count} anggota disinkronkan`)
              await loadRules()
            } catch {
              toast.error('Gagal sinkronisasi')
            }
          }}>
            <RefreshCw size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} />
            Sinkronisasi Anggota
          </button>
        </>
      )}
    </div>
  )
}

// ─── Info Tab ────────────────────────────────────────────────────────────────────

function InfoTab({ channel }: { channel: ChatChannel }) {
  return (
    <div className={styles.section}>
      <p className={styles.sectionTitle}>Informasi Channel</p>
      <div className={styles.infoRow}>
        <span className={styles.infoLabel}>Nama</span>
        <span className={styles.infoValue}>{channel.name}</span>
      </div>
      {channel.description && (
        <div className={styles.infoRow}>
          <span className={styles.infoLabel}>Deskripsi</span>
          <span className={styles.infoValue}>{channel.description}</span>
        </div>
      )}
      <div className={styles.infoRow}>
        <span className={styles.infoLabel}>Kategori</span>
        <span className={styles.infoValue}>{channel.categoryId}</span>
      </div>
    </div>
  )
}

// ─── Main Panel ──────────────────────────────────────────────────────────────────

export function ChannelSettingsPanel({ channel }: ChannelSettingsPanelProps) {
  const [activeTab, setActiveTab] = useState<Tab>('members')
  const setShowSettings = useChatStore((s) => s.setShowSettings)

  return (
    <div className={styles.settingsPanel}>
      {/* Header */}
      <div className={styles.settingsHeader}>
        <button className={styles.backBtn} onClick={() => setShowSettings(false)} title="Kembali">
          <ArrowLeft size={14} />
        </button>
        <span className={styles.settingsTitle}>#{channel.name}</span>
      </div>

      {/* Tabs */}
      <div className={styles.tabs}>
        <button className={`${styles.tab} ${activeTab === 'members' ? styles.tabActive : ''}`} onClick={() => setActiveTab('members')}>
          Members
        </button>
        <button className={`${styles.tab} ${activeTab === 'rules' ? styles.tabActive : ''}`} onClick={() => setActiveTab('rules')}>
          Aturan
        </button>
        <button className={`${styles.tab} ${activeTab === 'info' ? styles.tabActive : ''}`} onClick={() => setActiveTab('info')}>
          Info
        </button>
      </div>

      {/* Content */}
      <div className={styles.settingsContent}>
        {activeTab === 'members' && <MembersTab channelId={channel.id} />}
        {activeTab === 'rules' && <RulesTab channelId={channel.id} />}
        {activeTab === 'info' && <InfoTab channel={channel} />}
      </div>
    </div>
  )
}
