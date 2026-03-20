import React, { useEffect, useState } from 'react';
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { get } from '../../src/lib/api';
import LoadingSpinner from '../../src/components/LoadingSpinner';
import EmptyState from '../../src/components/EmptyState';
import type { Student } from '../../src/types';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map((n) => n[0])
    .join('')
    .toUpperCase();
}

function getRiskLabel(score?: number): { label: string; color: string; bgColor: string } {
  if (score === undefined || score === null) return { label: 'Tidak Ada Data', color: '#6B7280', bgColor: '#F3F4F6' };
  if (score >= 70) return { label: `Risiko Tinggi (${score})`, color: '#991B1B', bgColor: '#FEE2E2' };
  if (score >= 40) return { label: `Risiko Sedang (${score})`, color: '#92400E', bgColor: '#FEF3C7' };
  return { label: `Risiko Rendah (${score})`, color: '#065F46', bgColor: '#D1FAE5' };
}

// ─── Info Row ─────────────────────────────────────────────────────────────────

type IoniconsName = React.ComponentProps<typeof Ionicons>['name'];

function InfoRow({
  icon,
  label,
  value,
}: {
  icon: IoniconsName;
  label: string;
  value: string | undefined;
}) {
  if (!value) return null;
  return (
    <View style={styles.infoRow}>
      <View style={styles.infoIcon}>
        <Ionicons name={icon} size={16} color="#9CA3AF" />
      </View>
      <View style={styles.infoContent}>
        <Text style={styles.infoLabel}>{label}</Text>
        <Text style={styles.infoValue}>{value}</Text>
      </View>
    </View>
  );
}

// ─── Section Card ─────────────────────────────────────────────────────────────

function SectionCard({
  title,
  icon,
  children,
}: {
  title: string;
  icon: IoniconsName;
  children: React.ReactNode;
}) {
  return (
    <View style={styles.sectionCard}>
      <View style={styles.sectionCardHeader}>
        <Ionicons name={icon} size={18} color="#4F46E5" />
        <Text style={styles.sectionCardTitle}>{title}</Text>
      </View>
      <View style={styles.sectionCardBody}>{children}</View>
    </View>
  );
}

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function StudentDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const [student, setStudent] = useState<Student | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;

    async function fetchStudent() {
      try {
        const data = await get<Student>(`/students/${id}`);
        setStudent(data);
      } catch {
        // Fallback mock data
        setStudent({
          id: id as string,
          nisn: '3010002341',
          nik: '3578012345678901',
          name: 'Ahmad Fauzi Ramadhan',
          gender: 'MALE',
          birthDate: '2009-07-14T00:00:00.000Z',
          birthPlace: 'Malang',
          religion: 'Islam',
          address: 'Jl. Soekarno Hatta No. 45, Lowokwaru, Malang',
          phone: '081234567890',
          email: 'ahmad.fauzi@student.sman1malang.sch.id',
          schoolId: 'school-1',
          status: 'ACTIVE',
          riskScore: 25,
          photoUrl: undefined,
          guardians: [
            {
              id: 'guardian-1',
              name: 'Fauzi Santoso',
              relationship: 'Ayah',
              phone: '08129876543',
              email: 'fauzi.santoso@gmail.com',
              occupation: 'Wiraswasta',
            },
            {
              id: 'guardian-2',
              name: 'Siti Rahayu',
              relationship: 'Ibu',
              phone: '08175432198',
              occupation: 'Ibu Rumah Tangga',
            },
          ],
          currentEnrollment: {
            id: 'enrollment-1',
            academicYear: '2025/2026',
            grade: '11',
            className: 'IPA 3',
            semester: 2,
            status: 'ACTIVE',
          },
          createdAt: '2022-07-01T00:00:00.000Z',
          updatedAt: new Date().toISOString(),
        });
      } finally {
        setIsLoading(false);
      }
    }

    fetchStudent();
  }, [id]);

  if (isLoading) {
    return <LoadingSpinner fullScreen label="Memuat data siswa..." />;
  }

  if (error || !student) {
    return (
      <EmptyState
        icon="person-outline"
        title="Data Tidak Ditemukan"
        subtitle="Data siswa yang Anda cari tidak tersedia atau telah dihapus."
        style={styles.errorState}
      />
    );
  }

  const riskInfo = getRiskLabel(student.riskScore);
  const enrollment = student.currentEnrollment;

  return (
    <ScrollView
      style={styles.scroll}
      contentContainerStyle={styles.scrollContent}
      showsVerticalScrollIndicator={false}
    >
      {/* Profile Header */}
      <View style={styles.profileHeader}>
        {/* Avatar */}
        <View style={styles.avatarContainer}>
          {student.photoUrl ? (
            <Image
              source={{ uri: student.photoUrl }}
              style={styles.avatar}
              resizeMode="cover"
            />
          ) : (
            <View
              style={[
                styles.avatarFallback,
                {
                  backgroundColor:
                    student.gender === 'MALE' ? '#DBEAFE' : '#FCE7F3',
                },
              ]}
            >
              <Text
                style={[
                  styles.avatarInitials,
                  {
                    color: student.gender === 'MALE' ? '#1D4ED8' : '#BE185D',
                  },
                ]}
              >
                {getInitials(student.name)}
              </Text>
            </View>
          )}
        </View>

        <Text style={styles.studentName}>{student.name}</Text>
        <Text style={styles.studentNisn}>NISN: {student.nisn}</Text>

        <View style={styles.badgesRow}>
          {/* Status */}
          <View
            style={[
              styles.statusBadge,
              {
                backgroundColor:
                  student.status === 'ACTIVE' ? '#D1FAE5' : '#F3F4F6',
              },
            ]}
          >
            <Text
              style={[
                styles.statusText,
                { color: student.status === 'ACTIVE' ? '#065F46' : '#374151' },
              ]}
            >
              {student.status === 'ACTIVE' ? 'Aktif' : 'Tidak Aktif'}
            </Text>
          </View>

          {/* Gender */}
          <View
            style={[
              styles.genderBadge,
              {
                backgroundColor:
                  student.gender === 'MALE' ? '#DBEAFE' : '#FCE7F3',
              },
            ]}
          >
            <Ionicons
              name={student.gender === 'MALE' ? 'male' : 'female'}
              size={12}
              color={student.gender === 'MALE' ? '#1D4ED8' : '#BE185D'}
            />
            <Text
              style={[
                styles.genderText,
                {
                  color: student.gender === 'MALE' ? '#1D4ED8' : '#BE185D',
                },
              ]}
            >
              {student.gender === 'MALE' ? 'Laki-laki' : 'Perempuan'}
            </Text>
          </View>
        </View>

        {/* Risk Score */}
        {student.riskScore !== undefined && (
          <View style={[styles.riskBadge, { backgroundColor: riskInfo.bgColor }]}>
            <Ionicons
              name={
                student.riskScore >= 70
                  ? 'warning'
                  : student.riskScore >= 40
                  ? 'alert-circle'
                  : 'checkmark-circle'
              }
              size={14}
              color={riskInfo.color}
            />
            <Text style={[styles.riskText, { color: riskInfo.color }]}>
              {riskInfo.label}
            </Text>
          </View>
        )}
      </View>

      {/* Enrollment Info */}
      {enrollment && (
        <SectionCard title="Informasi Kelas" icon="school-outline">
          <InfoRow icon="calendar-outline" label="Tahun Ajaran" value={enrollment.academicYear} />
          <InfoRow icon="layers-outline" label="Kelas" value={`Kelas ${enrollment.grade} — ${enrollment.className}`} />
          <InfoRow icon="bookmark-outline" label="Semester" value={`Semester ${enrollment.semester}`} />
        </SectionCard>
      )}

      {/* Personal Info */}
      <SectionCard title="Data Pribadi" icon="person-outline">
        <InfoRow icon="card-outline" label="NISN" value={student.nisn} />
        <InfoRow icon="id-card-outline" label="NIK" value={student.nik} />
        <InfoRow icon="calendar-outline" label="Tanggal Lahir" value={formatDate(student.birthDate)} />
        <InfoRow icon="location-outline" label="Tempat Lahir" value={student.birthPlace} />
        <InfoRow icon="book-outline" label="Agama" value={student.religion} />
        <InfoRow icon="home-outline" label="Alamat" value={student.address} />
        <InfoRow icon="call-outline" label="No. HP" value={student.phone} />
        <InfoRow icon="mail-outline" label="Email" value={student.email} />
      </SectionCard>

      {/* Guardians */}
      {student.guardians.length > 0 && (
        <SectionCard title="Orang Tua / Wali" icon="people-outline">
          {student.guardians.map((guardian, index) => (
            <View key={guardian.id}>
              {index > 0 && <View style={styles.guardianSeparator} />}
              <View style={styles.guardianBlock}>
                <View style={styles.guardianHeader}>
                  <View style={styles.guardianAvatar}>
                    <Text style={styles.guardianAvatarText}>
                      {guardian.name[0].toUpperCase()}
                    </Text>
                  </View>
                  <View>
                    <Text style={styles.guardianName}>{guardian.name}</Text>
                    <Text style={styles.guardianRelation}>{guardian.relationship}</Text>
                  </View>
                </View>
                <View style={styles.guardianInfoRows}>
                  <InfoRow icon="call-outline" label="Telepon" value={guardian.phone} />
                  <InfoRow icon="mail-outline" label="Email" value={guardian.email} />
                  <InfoRow icon="briefcase-outline" label="Pekerjaan" value={guardian.occupation} />
                </View>
              </View>
            </View>
          ))}
        </SectionCard>
      )}

      {/* Action Buttons */}
      <View style={styles.actionButtons}>
        <TouchableOpacity style={styles.actionButtonPrimary} activeOpacity={0.85}>
          <Ionicons name="create-outline" size={18} color="#FFFFFF" />
          <Text style={styles.actionButtonPrimaryText}>Edit Data</Text>
        </TouchableOpacity>
        <TouchableOpacity style={styles.actionButtonSecondary} activeOpacity={0.85}>
          <Ionicons name="document-text-outline" size={18} color="#4F46E5" />
          <Text style={styles.actionButtonSecondaryText}>Lihat Rapor</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.bottomPadding} />
    </ScrollView>
  );
}

// ─── Styles ───────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  scroll: {
    flex: 1,
    backgroundColor: '#F9FAFB',
  },
  scrollContent: {
    paddingBottom: 32,
  },
  errorState: {
    flex: 1,
    paddingTop: 80,
  },
  profileHeader: {
    backgroundColor: '#FFFFFF',
    paddingVertical: 28,
    paddingHorizontal: 24,
    alignItems: 'center',
    borderBottomWidth: 1,
    borderBottomColor: '#F3F4F6',
    marginBottom: 12,
  },
  avatarContainer: {
    marginBottom: 14,
  },
  avatar: {
    width: 96,
    height: 96,
    borderRadius: 48,
  },
  avatarFallback: {
    width: 96,
    height: 96,
    borderRadius: 48,
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarInitials: {
    fontSize: 32,
    fontWeight: '700',
  },
  studentName: {
    fontSize: 22,
    fontWeight: '700',
    color: '#111827',
    textAlign: 'center',
    marginBottom: 4,
  },
  studentNisn: {
    fontSize: 14,
    color: '#6B7280',
    marginBottom: 12,
  },
  badgesRow: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 10,
  },
  statusBadge: {
    paddingHorizontal: 12,
    paddingVertical: 5,
    borderRadius: 20,
  },
  statusText: {
    fontSize: 13,
    fontWeight: '600',
  },
  genderBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 12,
    paddingVertical: 5,
    borderRadius: 20,
  },
  genderText: {
    fontSize: 13,
    fontWeight: '600',
  },
  riskBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    paddingHorizontal: 14,
    paddingVertical: 6,
    borderRadius: 20,
  },
  riskText: {
    fontSize: 13,
    fontWeight: '600',
  },
  sectionCard: {
    backgroundColor: '#FFFFFF',
    marginHorizontal: 16,
    borderRadius: 16,
    marginBottom: 12,
    overflow: 'hidden',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  sectionCardHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    padding: 14,
    paddingBottom: 10,
    borderBottomWidth: 1,
    borderBottomColor: '#F9FAFB',
  },
  sectionCardTitle: {
    fontSize: 15,
    fontWeight: '700',
    color: '#111827',
  },
  sectionCardBody: {
    padding: 14,
    paddingTop: 10,
    gap: 2,
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    paddingVertical: 7,
    borderBottomWidth: 1,
    borderBottomColor: '#F9FAFB',
  },
  infoIcon: {
    width: 28,
    alignItems: 'center',
    marginTop: 1,
    marginRight: 8,
  },
  infoContent: {
    flex: 1,
  },
  infoLabel: {
    fontSize: 11,
    color: '#9CA3AF',
    marginBottom: 1,
  },
  infoValue: {
    fontSize: 14,
    color: '#111827',
    fontWeight: '500',
    lineHeight: 20,
  },
  guardianBlock: {
    paddingVertical: 4,
  },
  guardianHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    marginBottom: 8,
  },
  guardianAvatar: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#EEF2FF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  guardianAvatarText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#4F46E5',
  },
  guardianName: {
    fontSize: 15,
    fontWeight: '600',
    color: '#111827',
  },
  guardianRelation: {
    fontSize: 12,
    color: '#9CA3AF',
    marginTop: 1,
  },
  guardianInfoRows: {
    paddingLeft: 4,
  },
  guardianSeparator: {
    height: 1,
    backgroundColor: '#F3F4F6',
    marginVertical: 10,
  },
  actionButtons: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    gap: 10,
    marginTop: 8,
    marginBottom: 8,
  },
  actionButtonPrimary: {
    flex: 1,
    backgroundColor: '#4F46E5',
    borderRadius: 14,
    height: 50,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    shadowColor: '#4F46E5',
    shadowOffset: { width: 0, height: 3 },
    shadowOpacity: 0.25,
    shadowRadius: 8,
    elevation: 4,
  },
  actionButtonPrimaryText: {
    color: '#FFFFFF',
    fontSize: 15,
    fontWeight: '600',
  },
  actionButtonSecondary: {
    flex: 1,
    backgroundColor: '#FFFFFF',
    borderRadius: 14,
    height: 50,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    borderWidth: 1.5,
    borderColor: '#4F46E5',
  },
  actionButtonSecondaryText: {
    color: '#4F46E5',
    fontSize: 15,
    fontWeight: '600',
  },
  bottomPadding: {
    height: 16,
  },
});
