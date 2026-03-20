'use client';

import React, { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import {
  ArrowLeft,
  User,
  Phone,
  MapPin,
  Calendar,
  BookOpen,
  CreditCard,
  Heart,
  AlertTriangle,
  TrendingUp,
} from 'lucide-react';
import api from '@/lib/api';
import type { Student, StudentAttendance, Grade } from '@/types';
import { Card, CardHeader } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge, RiskScoreBadge } from '@/components/ui/Badge';
import { formatDate, genderLabel } from '@/lib/utils';

type TabKey = 'overview' | 'attendance' | 'grades' | 'payments';

export default function StudentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [student, setStudent] = useState<Student | null>(null);
  const [attendance, setAttendance] = useState<StudentAttendance[]>([]);
  const [grades, setGrades] = useState<Grade[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<TabKey>('overview');

  useEffect(() => {
    const fetchStudent = async () => {
      setIsLoading(true);
      try {
        const [sRes, aRes] = await Promise.all([
          api.get(`/students/${id}`),
          api.get(`/students/${id}/attendance`),
        ]);
        setStudent(sRes.data.data);
        setAttendance(aRes.data.data || []);

        const gRes = await api.get(`/academic/grades/${id}`);
        setGrades(gRes.data.data || []);
      } catch {
        // handle error
      } finally {
        setIsLoading(false);
      }
    };
    fetchStudent();
  }, [id]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="w-7 h-7 border-2 border-primary-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!student) {
    return (
      <div className="text-center py-24">
        <p className="text-gray-500">Siswa tidak ditemukan</p>
        <Button variant="ghost" className="mt-4" onClick={() => router.back()}>
          Kembali
        </Button>
      </div>
    );
  }

  const tabs: { key: TabKey; label: string; icon: React.ReactNode }[] = [
    { key: 'overview', label: 'Profil', icon: <User className="w-4 h-4" /> },
    { key: 'attendance', label: 'Kehadiran', icon: <Calendar className="w-4 h-4" /> },
    { key: 'grades', label: 'Nilai', icon: <BookOpen className="w-4 h-4" /> },
    { key: 'payments', label: 'Pembayaran', icon: <CreditCard className="w-4 h-4" /> },
  ];

  const presentCount = attendance.filter((a) => a.status === 'PRESENT').length;
  const absentCount = attendance.filter((a) => a.status === 'ABSENT').length;
  const attendanceRate = attendance.length > 0 ? (presentCount / attendance.length) * 100 : 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" leftIcon={<ArrowLeft className="w-4 h-4" />} onClick={() => router.back()}>
          Kembali
        </Button>
      </div>

      {/* Profile card */}
      <Card>
        <div className="flex items-start gap-6">
          <div className="w-20 h-20 rounded-2xl bg-primary-100 flex items-center justify-center flex-shrink-0">
            {student.photoUrl ? (
              <img src={student.photoUrl} alt={student.name} className="w-20 h-20 rounded-2xl object-cover" />
            ) : (
              <User className="w-10 h-10 text-primary-400" />
            )}
          </div>

          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h1 className="text-xl font-bold text-gray-900">{student.name}</h1>
                <p className="text-sm text-gray-500 mt-0.5">NISN: {student.nisn}</p>
              </div>
              {student.riskScore !== undefined && (
                <RiskScoreBadge score={student.riskScore} />
              )}
            </div>

            <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4">
              <InfoItem icon={<Calendar className="w-4 h-4" />} label="Tanggal Lahir" value={formatDate(student.birthDate)} />
              <InfoItem icon={<User className="w-4 h-4" />} label="Jenis Kelamin" value={genderLabel(student.gender)} />
              <InfoItem icon={<BookOpen className="w-4 h-4" />} label="Kelas" value={student.currentClass || '—'} />
              {student.address && (
                <InfoItem icon={<MapPin className="w-4 h-4" />} label="Alamat" value={student.address} />
              )}
            </div>
          </div>
        </div>
      </Card>

      {/* Quick stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="text-center">
            <p className="text-2xl font-bold text-primary-600">{attendanceRate.toFixed(0)}%</p>
            <p className="text-xs text-gray-500 mt-1">Tingkat Kehadiran</p>
          </div>
        </Card>
        <Card padding="sm">
          <div className="text-center">
            <p className="text-2xl font-bold text-green-600">{presentCount}</p>
            <p className="text-xs text-gray-500 mt-1">Hari Hadir</p>
          </div>
        </Card>
        <Card padding="sm">
          <div className="text-center">
            <p className="text-2xl font-bold text-red-600">{absentCount}</p>
            <p className="text-xs text-gray-500 mt-1">Hari Absen</p>
          </div>
        </Card>
      </div>

      {/* Tabs */}
      <div>
        <div className="flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                activeTab === tab.key
                  ? 'bg-white text-primary-700 shadow-sm'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>

        <div className="mt-4">
          {activeTab === 'overview' && (
            <Card>
              <CardHeader title="Data Wali Siswa" />
              {student.guardians && student.guardians.length > 0 ? (
                <div className="space-y-3">
                  {student.guardians.map((g) => (
                    <div key={g.id} className="flex items-center gap-4 p-3 bg-gray-50 rounded-lg">
                      <div className="w-10 h-10 rounded-full bg-primary-100 flex items-center justify-center">
                        <User className="w-5 h-5 text-primary-500" />
                      </div>
                      <div>
                        <p className="font-medium text-sm text-gray-900">{g.name}</p>
                        <p className="text-xs text-gray-500">{g.relationship}</p>
                      </div>
                      <div className="ml-auto flex items-center gap-1 text-sm text-gray-600">
                        <Phone className="w-3.5 h-3.5" />
                        {g.phone}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-gray-400">Belum ada data wali</p>
              )}
            </Card>
          )}

          {activeTab === 'attendance' && (
            <Card padding="none">
              <div className="p-4 border-b border-gray-100">
                <h3 className="font-semibold text-sm text-gray-900">Riwayat Kehadiran</h3>
              </div>
              <div className="divide-y divide-gray-50">
                {attendance.length === 0 ? (
                  <p className="p-6 text-center text-sm text-gray-400">Belum ada data kehadiran</p>
                ) : (
                  attendance.slice(0, 30).map((a) => (
                    <div key={a.id} className="flex items-center justify-between px-4 py-3">
                      <span className="text-sm text-gray-700">{formatDate(a.date)}</span>
                      <AttBadge status={a.status} />
                    </div>
                  ))
                )}
              </div>
            </Card>
          )}

          {activeTab === 'grades' && (
            <Card padding="none">
              <div className="p-4 border-b border-gray-100">
                <h3 className="font-semibold text-sm text-gray-900">Nilai Akademik</h3>
              </div>
              {grades.length === 0 ? (
                <p className="p-6 text-center text-sm text-gray-400">Belum ada data nilai</p>
              ) : (
                <table className="w-full text-sm">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Mata Pelajaran</th>
                      <th className="text-center px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Harian</th>
                      <th className="text-center px-4 py-3 text-xs font-semibold text-gray-500 uppercase">UTS</th>
                      <th className="text-center px-4 py-3 text-xs font-semibold text-gray-500 uppercase">UAS</th>
                      <th className="text-center px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Akhir</th>
                      <th className="text-center px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Predikat</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-50">
                    {grades.map((g) => (
                      <tr key={g.id} className="hover:bg-gray-50">
                        <td className="px-4 py-3">{g.subjectId}</td>
                        <td className="px-4 py-3 text-center">{g.dailyScore ?? '—'}</td>
                        <td className="px-4 py-3 text-center">{g.midtermScore ?? '—'}</td>
                        <td className="px-4 py-3 text-center">{g.finalScore ?? '—'}</td>
                        <td className="px-4 py-3 text-center font-medium text-primary-700">
                          {g.finalGrade ?? '—'}
                        </td>
                        <td className="px-4 py-3 text-center">
                          {g.predicate ? (
                            <Badge variant="primary" size="sm">{g.predicate}</Badge>
                          ) : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </Card>
          )}

          {activeTab === 'payments' && (
            <Card>
              <CardHeader title="Riwayat Pembayaran" />
              <div className="flex flex-col items-center py-8 text-center text-gray-400">
                <CreditCard className="w-10 h-10 mb-2 text-gray-200" />
                <p className="text-sm">Lihat detail pembayaran di halaman SPP</p>
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-3"
                  onClick={() => router.push('/dashboard/payments')}
                >
                  Ke Halaman Pembayaran
                </Button>
              </div>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}

function InfoItem({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="flex items-start gap-2">
      <span className="text-gray-400 mt-0.5">{icon}</span>
      <div>
        <p className="text-xs text-gray-400">{label}</p>
        <p className="text-sm font-medium text-gray-800">{value}</p>
      </div>
    </div>
  );
}

function AttBadge({ status }: { status: string }) {
  const map: Record<string, { label: string; color: string }> = {
    PRESENT: { label: 'Hadir', color: 'text-green-600 bg-green-50' },
    ABSENT: { label: 'Tidak Hadir', color: 'text-red-600 bg-red-50' },
    LATE: { label: 'Terlambat', color: 'text-amber-600 bg-amber-50' },
    SICK: { label: 'Sakit', color: 'text-blue-600 bg-blue-50' },
    PERMITTED: { label: 'Izin', color: 'text-purple-600 bg-purple-50' },
  };
  const { label, color } = map[status] || { label: status, color: 'text-gray-600 bg-gray-100' };
  return (
    <span className={`text-xs font-medium px-2.5 py-1 rounded-full ${color}`}>
      {label}
    </span>
  );
}
