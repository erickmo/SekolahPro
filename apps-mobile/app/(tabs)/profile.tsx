import React, { useState } from 'react';
import {
  Alert,
  Image,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../../src/lib/auth';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getRoleLabel(role: string): string {
  const labels: Record<string, string> = {
    SUPERADMIN: 'Super Admin',
    ADMIN_YAYASAN: 'Admin Yayasan',
    ADMIN_SEKOLAH: 'Admin Sekolah',
    KEPALA_SEKOLAH: 'Kepala Sekolah',
    KEPALA_KURIKULUM: 'Kepala Kurikulum',
    BENDAHARA: 'Bendahara',
    GURU: 'Guru',
    WALI_KELAS: 'Wali Kelas',
    GURU_BK: 'Guru BK',
    OPERATOR_SIMS: 'Operator SIMS',
    KASIR_KOPERASI: 'Kasir Koperasi',
    PUSTAKAWAN: 'Pustakawan',
    PETUGAS_UKS: 'Petugas UKS',
    TATA_USAHA: 'Tata Usaha',
    SISWA: 'Siswa',
    ORANG_TUA: 'Orang Tua',
    ALUMNI: 'Alumni',
  };
  return labels[role] ?? role;
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map((n) => n[0])
    .join('')
    .toUpperCase();
}

// ─── Menu Item ────────────────────────────────────────────────────────────────

type IoniconsName = React.ComponentProps<typeof Ionicons>['name'];

interface MenuItemProps {
  icon: IoniconsName;
  label: string;
  subtitle?: string;
  onPress: () => void;
  color?: string;
  showArrow?: boolean;
  badge?: string;
}

function MenuItem({
  icon,
  label,
  subtitle,
  onPress,
  color = '#374151',
  showArrow = true,
  badge,
}: MenuItemProps) {
  return (
    <TouchableOpacity style={styles.menuItem} onPress={onPress} activeOpacity={0.7}>
      <View style={[styles.menuItemIcon, { backgroundColor: color + '15' }]}>
        <Ionicons name={icon} size={20} color={color} />
      </View>
      <View style={styles.menuItemContent}>
        <Text style={[styles.menuItemLabel, { color: color === '#EF4444' ? '#EF4444' : '#111827' }]}>
          {label}
        </Text>
        {subtitle ? (
          <Text style={styles.menuItemSubtitle}>{subtitle}</Text>
        ) : null}
      </View>
      {badge ? (
        <View style={styles.menuItemBadge}>
          <Text style={styles.menuItemBadgeText}>{badge}</Text>
        </View>
      ) : null}
      {showArrow && (
        <Ionicons name="chevron-forward" size={16} color="#D1D5DB" />
      )}
    </TouchableOpacity>
  );
}

// ─── Section ──────────────────────────────────────────────────────────────────

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <View style={styles.section}>
      <Text style={styles.sectionTitle}>{title}</Text>
      <View style={styles.sectionCard}>{children}</View>
    </View>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function ProfileScreen() {
  const { user, school, logout } = useAuth();
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const handleLogout = () => {
    Alert.alert(
      'Keluar',
      'Apakah Anda yakin ingin keluar dari akun ini?',
      [
        { text: 'Batal', style: 'cancel' },
        {
          text: 'Ya, Keluar',
          style: 'destructive',
          onPress: async () => {
            setIsLoggingOut(true);
            try {
              await logout();
              // AuthGate in _layout.tsx handles redirect to login
            } catch {
              setIsLoggingOut(false);
            }
          },
        },
      ],
      { cancelable: true },
    );
  };

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Header */}
        <View style={styles.pageHeader}>
          <Text style={styles.pageTitle}>Profil</Text>
          <TouchableOpacity style={styles.settingsButton} activeOpacity={0.7}>
            <Ionicons name="settings-outline" size={22} color="#374151" />
          </TouchableOpacity>
        </View>

        {/* Profile Card */}
        <View style={styles.profileCard}>
          {/* Avatar */}
          <View style={styles.avatarContainer}>
            {user?.photoUrl ? (
              <Image
                source={{ uri: user.photoUrl }}
                style={styles.avatar}
                resizeMode="cover"
              />
            ) : (
              <View style={styles.avatarFallback}>
                <Text style={styles.avatarInitials}>
                  {getInitials(user?.name ?? 'U')}
                </Text>
              </View>
            )}
            <TouchableOpacity style={styles.avatarEditButton} activeOpacity={0.8}>
              <Ionicons name="camera" size={14} color="#FFFFFF" />
            </TouchableOpacity>
          </View>

          {/* User Info */}
          <Text style={styles.userName}>{user?.name ?? 'Pengguna'}</Text>
          <Text style={styles.userEmail}>{user?.email ?? ''}</Text>

          <View style={styles.roleBadge}>
            <Ionicons name="shield-checkmark" size={13} color="#4F46E5" />
            <Text style={styles.roleText}>
              {getRoleLabel(user?.role ?? 'GURU')}
            </Text>
          </View>

          {/* School Info */}
          {school ? (
            <View style={styles.schoolInfo}>
              <Ionicons name="school-outline" size={14} color="#9CA3AF" />
              <Text style={styles.schoolName} numberOfLines={1}>
                {school.name}
              </Text>
              <View style={styles.planBadge}>
                <Text style={styles.planText}>{school.tenantPlan}</Text>
              </View>
            </View>
          ) : null}
        </View>

        {/* Account Section */}
        <Section title="Akun">
          <MenuItem
            icon="person-outline"
            label="Edit Profil"
            subtitle="Ubah nama, foto, dan info pribadi"
            onPress={() => {}}
          />
          <View style={styles.menuSeparator} />
          <MenuItem
            icon="lock-closed-outline"
            label="Ubah Password"
            subtitle="Perbarui password akun Anda"
            onPress={() => {}}
          />
          <View style={styles.menuSeparator} />
          <MenuItem
            icon="notifications-outline"
            label="Pengaturan Notifikasi"
            subtitle="Atur preferensi pemberitahuan"
            onPress={() => {}}
            badge="3 baru"
          />
        </Section>

        {/* App Section */}
        <Section title="Aplikasi">
          <MenuItem
            icon="moon-outline"
            label="Mode Gelap"
            subtitle="Tampilan gelap untuk kenyamanan mata"
            onPress={() => {}}
          />
          <View style={styles.menuSeparator} />
          <MenuItem
            icon="language-outline"
            label="Bahasa"
            subtitle="Bahasa Indonesia"
            onPress={() => {}}
          />
          <View style={styles.menuSeparator} />
          <MenuItem
            icon="download-outline"
            label="Unduh Data Offline"
            subtitle="Simpan data untuk akses tanpa internet"
            onPress={() => {}}
          />
        </Section>

        {/* Support Section */}
        <Section title="Bantuan">
          <MenuItem
            icon="help-circle-outline"
            label="Pusat Bantuan"
            subtitle="FAQ dan panduan penggunaan"
            onPress={() => {}}
          />
          <View style={styles.menuSeparator} />
          <MenuItem
            icon="chatbubble-outline"
            label="Hubungi Dukungan"
            subtitle="Chat dengan tim dukungan EDS"
            onPress={() => {}}
          />
          <View style={styles.menuSeparator} />
          <MenuItem
            icon="information-circle-outline"
            label="Tentang Aplikasi"
            subtitle="EDS Mobile v1.0.0"
            onPress={() => {}}
          />
        </Section>

        {/* Logout */}
        <View style={styles.section}>
          <View style={styles.sectionCard}>
            <MenuItem
              icon="log-out-outline"
              label={isLoggingOut ? 'Sedang keluar...' : 'Keluar'}
              color="#EF4444"
              showArrow={false}
              onPress={isLoggingOut ? () => {} : handleLogout}
            />
          </View>
        </View>

        {/* Footer */}
        <View style={styles.footer}>
          <Text style={styles.footerText}>EDS — Ekosistem Digital Sekolah</Text>
          <Text style={styles.footerVersion}>Versi 1.0.0 · Build 2026.03.20</Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

// ─── Styles ───────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#F9FAFB',
  },
  scroll: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 32,
  },
  pageHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingTop: 20,
    paddingBottom: 16,
  },
  pageTitle: {
    fontSize: 24,
    fontWeight: '700',
    color: '#111827',
  },
  settingsButton: {
    width: 44,
    height: 44,
    borderRadius: 12,
    backgroundColor: '#FFFFFF',
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  profileCard: {
    backgroundColor: '#FFFFFF',
    marginHorizontal: 16,
    borderRadius: 20,
    padding: 24,
    alignItems: 'center',
    marginBottom: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.06,
    shadowRadius: 10,
    elevation: 3,
  },
  avatarContainer: {
    position: 'relative',
    marginBottom: 14,
  },
  avatar: {
    width: 88,
    height: 88,
    borderRadius: 44,
  },
  avatarFallback: {
    width: 88,
    height: 88,
    borderRadius: 44,
    backgroundColor: '#EEF2FF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarInitials: {
    fontSize: 28,
    fontWeight: '700',
    color: '#4F46E5',
  },
  avatarEditButton: {
    position: 'absolute',
    bottom: 2,
    right: 2,
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: '#4F46E5',
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 2,
    borderColor: '#FFFFFF',
  },
  userName: {
    fontSize: 20,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 2,
    textAlign: 'center',
  },
  userEmail: {
    fontSize: 14,
    color: '#6B7280',
    marginBottom: 10,
    textAlign: 'center',
  },
  roleBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    backgroundColor: '#EEF2FF',
    paddingHorizontal: 12,
    paddingVertical: 5,
    borderRadius: 20,
    marginBottom: 12,
  },
  roleText: {
    fontSize: 13,
    fontWeight: '600',
    color: '#4F46E5',
  },
  schoolInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: '#F9FAFB',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 10,
    width: '100%',
    justifyContent: 'center',
  },
  schoolName: {
    fontSize: 13,
    color: '#6B7280',
    flex: 1,
  },
  planBadge: {
    backgroundColor: '#D1FAE5',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 8,
  },
  planText: {
    fontSize: 10,
    color: '#065F46',
    fontWeight: '700',
    letterSpacing: 0.5,
  },
  section: {
    paddingHorizontal: 16,
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 13,
    fontWeight: '600',
    color: '#9CA3AF',
    textTransform: 'uppercase',
    letterSpacing: 0.8,
    marginBottom: 8,
    marginLeft: 4,
  },
  sectionCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    overflow: 'hidden',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.04,
    shadowRadius: 4,
    elevation: 1,
  },
  menuItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 14,
  },
  menuItemIcon: {
    width: 40,
    height: 40,
    borderRadius: 12,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
  },
  menuItemContent: {
    flex: 1,
  },
  menuItemLabel: {
    fontSize: 15,
    fontWeight: '500',
    color: '#111827',
    marginBottom: 1,
  },
  menuItemSubtitle: {
    fontSize: 12,
    color: '#9CA3AF',
  },
  menuItemBadge: {
    backgroundColor: '#EEF2FF',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 10,
    marginRight: 6,
  },
  menuItemBadgeText: {
    fontSize: 11,
    color: '#4F46E5',
    fontWeight: '600',
  },
  menuSeparator: {
    height: 1,
    backgroundColor: '#F9FAFB',
    marginLeft: 66,
  },
  footer: {
    alignItems: 'center',
    paddingTop: 8,
    paddingBottom: 8,
    gap: 4,
  },
  footerText: {
    fontSize: 12,
    color: '#9CA3AF',
    fontWeight: '500',
  },
  footerVersion: {
    fontSize: 11,
    color: '#D1D5DB',
  },
});
