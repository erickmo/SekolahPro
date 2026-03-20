import React from 'react';
import {
  Image,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import type { StudentListItem } from '../types';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getStatusConfig(status: StudentListItem['status']) {
  switch (status) {
    case 'ACTIVE':
      return { label: 'Aktif', color: '#065F46', bgColor: '#D1FAE5' };
    case 'INACTIVE':
      return { label: 'Tidak Aktif', color: '#374151', bgColor: '#F3F4F6' };
    case 'TRANSFERRED':
      return { label: 'Pindah', color: '#92400E', bgColor: '#FEF3C7' };
    case 'GRADUATED':
      return { label: 'Lulus', color: '#1E40AF', bgColor: '#DBEAFE' };
    default:
      return { label: status, color: '#374151', bgColor: '#F3F4F6' };
  }
}

function getRiskColor(score?: number): string {
  if (score === undefined || score === null) return 'transparent';
  if (score >= 70) return '#EF4444';
  if (score >= 40) return '#F59E0B';
  return '#10B981';
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .slice(0, 2)
    .map((n) => n[0])
    .join('')
    .toUpperCase();
}

// ─── Types ────────────────────────────────────────────────────────────────────

interface StudentCardProps {
  student: StudentListItem;
  onPress: (id: string) => void;
}

// ─── Component ────────────────────────────────────────────────────────────────

export default function StudentCard({ student, onPress }: StudentCardProps) {
  const statusConfig = getStatusConfig(student.status);
  const hasRiskScore =
    student.riskScore !== undefined && student.riskScore !== null;

  return (
    <TouchableOpacity
      style={styles.card}
      onPress={() => onPress(student.id)}
      activeOpacity={0.75}
    >
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

        {/* Gender icon overlay */}
        <View
          style={[
            styles.genderBadge,
            {
              backgroundColor:
                student.gender === 'MALE' ? '#3B82F6' : '#EC4899',
            },
          ]}
        >
          <Ionicons
            name={student.gender === 'MALE' ? 'male' : 'female'}
            size={9}
            color="#FFF"
          />
        </View>
      </View>

      {/* Info */}
      <View style={styles.info}>
        <Text style={styles.name} numberOfLines={1}>
          {student.name}
        </Text>
        <Text style={styles.nisn} numberOfLines={1}>
          NISN: {student.nisn}
        </Text>
        {student.className ? (
          <Text style={styles.className} numberOfLines={1}>
            {student.grade ? `Kelas ${student.grade} — ` : ''}
            {student.className}
          </Text>
        ) : null}
      </View>

      {/* Right side */}
      <View style={styles.right}>
        <View
          style={[
            styles.statusBadge,
            { backgroundColor: statusConfig.bgColor },
          ]}
        >
          <Text style={[styles.statusText, { color: statusConfig.color }]}>
            {statusConfig.label}
          </Text>
        </View>

        {hasRiskScore && (
          <View style={styles.riskContainer}>
            <View
              style={[
                styles.riskDot,
                { backgroundColor: getRiskColor(student.riskScore) },
              ]}
            />
            <Text style={styles.riskScore}>{student.riskScore}</Text>
          </View>
        )}

        <Ionicons name="chevron-forward" size={16} color="#D1D5DB" />
      </View>
    </TouchableOpacity>
  );
}

// ─── Styles ───────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 14,
    paddingHorizontal: 14,
    paddingVertical: 12,
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 10,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  avatarContainer: {
    position: 'relative',
    marginRight: 12,
  },
  avatar: {
    width: 48,
    height: 48,
    borderRadius: 24,
  },
  avatarFallback: {
    width: 48,
    height: 48,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarInitials: {
    fontSize: 16,
    fontWeight: '700',
  },
  genderBadge: {
    position: 'absolute',
    bottom: 0,
    right: 0,
    width: 16,
    height: 16,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1.5,
    borderColor: '#FFF',
  },
  info: {
    flex: 1,
    marginRight: 8,
  },
  name: {
    fontSize: 15,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 2,
  },
  nisn: {
    fontSize: 12,
    color: '#6B7280',
    marginBottom: 1,
  },
  className: {
    fontSize: 12,
    color: '#9CA3AF',
  },
  right: {
    alignItems: 'flex-end',
    gap: 4,
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 20,
  },
  statusText: {
    fontSize: 11,
    fontWeight: '600',
  },
  riskContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  riskDot: {
    width: 7,
    height: 7,
    borderRadius: 4,
  },
  riskScore: {
    fontSize: 11,
    color: '#6B7280',
    fontWeight: '500',
  },
});
