'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { Users, Bell, MessageCircle, Pin } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import api from '@/lib/api';
import { Card, CardHeader } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Pagination } from '@/components/ui/Pagination';
import { Badge } from '@/components/ui/Badge';
import { formatDate, formatDateTime } from '@/lib/utils';

interface Announcement {
  id: string;
  title: string;
  type: string;
  publishedAt: string;
  priority: 'LOW' | 'MEDIUM' | 'HIGH';
  isPublished: boolean;
}

interface ParentMessage {
  id: string;
  senderName: string;
  senderRole: string;
  subject: string;
  preview: string;
  sentAt: string;
  isRead: boolean;
}

type TabKey = 'announcements' | 'messages';

const TYPE_LABELS: Record<string, string> = {
  ACADEMIC: 'Akademik',
  ADMINISTRATIVE: 'Administrasi',
  EVENT: 'Acara',
  FINANCIAL: 'Keuangan',
  HEALTH: 'Kesehatan',
  GENERAL: 'Umum',
};

const PRIORITY_CONFIG: Record<string, { label: string; variant: 'danger' | 'warning' | 'gray' }> = {
  HIGH: { label: 'Tinggi', variant: 'danger' },
  MEDIUM: { label: 'Sedang', variant: 'warning' },
  LOW: { label: 'Rendah', variant: 'gray' },
};

export default function ParentPortalPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabKey>('announcements');
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [messages, setMessages] = useState<ParentMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [announcePage, setAnnouncePage] = useState(1);
  const [messagePage, setMessagePage] = useState(1);
  const [announceMeta, setAnnounceMeta] = useState({ total: 0, totalPages: 1 });
  const [messageMeta, setMessageMeta] = useState({ total: 0, totalPages: 1 });

  const fetchAnnouncements = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(announcePage), limit: '20' });
      const res = await api.get(`/parent-portal/announcements?${params}`);
      setAnnouncements(res.data.data || []);
      setAnnounceMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setAnnouncements([]);
    } finally {
      setIsLoading(false);
    }
  }, [announcePage]);

  const fetchMessages = useCallback(async () => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams({ page: String(messagePage), limit: '20' });
      const res = await api.get(`/parent-portal/messages?${params}`);
      setMessages(res.data.data || []);
      setMessageMeta(res.data.meta || { total: 0, totalPages: 1 });
    } catch {
      setMessages([]);
    } finally {
      setIsLoading(false);
    }
  }, [messagePage]);

  useEffect(() => { fetchAnnouncements(); }, [fetchAnnouncements]);
  useEffect(() => { fetchMessages(); }, [fetchMessages]);

  const unreadMessages = messages.filter((m) => !m.isRead).length;
  const highPriorityCount = announcements.filter((a) => a.priority === 'HIGH').length;

  const announcementColumns = [
    {
      key: 'title',
      header: 'Judul',
      render: (a: Announcement) => (
        <div className="flex items-center gap-2">
          {a.priority === 'HIGH' && <Pin className="w-3.5 h-3.5 text-red-500 flex-shrink-0" />}
          <span className="font-medium text-gray-900">{a.title}</span>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Tipe',
      render: (a: Announcement) => (
        <Badge variant="primary" size="sm">{TYPE_LABELS[a.type] || a.type}</Badge>
      ),
    },
    {
      key: 'publishedAt',
      header: 'Diterbitkan',
      render: (a: Announcement) => formatDate(a.publishedAt),
    },
    {
      key: 'priority',
      header: 'Prioritas',
      render: (a: Announcement) => {
        const cfg = PRIORITY_CONFIG[a.priority] || { label: a.priority, variant: 'gray' as const };
        return <Badge variant={cfg.variant} size="sm">{cfg.label}</Badge>;
      },
    },
  ];

  const messageColumns = [
    {
      key: 'senderName',
      header: 'Pengirim',
      render: (m: ParentMessage) => (
        <div className="flex items-center gap-2">
          {!m.isRead && <span className="w-2 h-2 bg-primary-500 rounded-full flex-shrink-0" />}
          <div>
            <p className={`text-sm ${!m.isRead ? 'font-semibold text-gray-900' : 'text-gray-700'}`}>
              {m.senderName}
            </p>
            <p className="text-xs text-gray-400">{m.senderRole}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'subject',
      header: 'Subjek',
      render: (m: ParentMessage) => (
        <span className={m.isRead ? 'text-gray-700' : 'font-medium text-gray-900'}>{m.subject}</span>
      ),
    },
    {
      key: 'preview',
      header: 'Pratinjau',
      render: (m: ParentMessage) => (
        <p className="max-w-xs text-sm text-gray-500 truncate">{m.preview}</p>
      ),
    },
    {
      key: 'sentAt',
      header: 'Waktu',
      render: (m: ParentMessage) => formatDateTime(m.sentAt),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Portal Orang Tua</h1>
          <p className="text-sm text-gray-500 mt-0.5">Informasi dan komunikasi dengan orang tua siswa</p>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-primary-50 rounded-lg">
              <Users className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{announceMeta.total}</p>
              <p className="text-xs text-gray-500">Total Pengumuman</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-red-50 rounded-lg">
              <Bell className="w-5 h-5 text-red-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{highPriorityCount}</p>
              <p className="text-xs text-gray-500">Prioritas Tinggi</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-blue-50 rounded-lg">
              <MessageCircle className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-xl font-bold text-gray-900">{unreadMessages}</p>
              <p className="text-xs text-gray-500">Pesan Belum Dibaca</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
        {([
          { key: 'announcements', label: 'Pengumuman' },
          { key: 'messages', label: 'Pesan' },
        ] as { key: TabKey; label: string }[]).map((t) => (
          <button
            key={t.key}
            onClick={() => setActiveTab(t.key)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === t.key ? 'bg-white text-primary-700 shadow-sm' : 'text-gray-500'
            }`}
          >
            {t.label}
            {t.key === 'messages' && unreadMessages > 0 && (
              <span className="ml-1.5 inline-flex items-center justify-center px-1.5 py-0.5 rounded-full text-xs font-bold bg-primary-600 text-white">
                {unreadMessages}
              </span>
            )}
          </button>
        ))}
      </div>

      {activeTab === 'announcements' && (
        <Card padding="none">
          <Table
            columns={announcementColumns}
            data={announcements}
            keyExtractor={(a) => a.id}
            isLoading={isLoading}
            emptyMessage="Belum ada pengumuman untuk orang tua"
          />
          {announceMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={announcePage}
                totalPages={announceMeta.totalPages}
                total={announceMeta.total}
                limit={20}
                onPageChange={setAnnouncePage}
              />
            </div>
          )}
        </Card>
      )}

      {activeTab === 'messages' && (
        <Card padding="none">
          <Table
            columns={messageColumns}
            data={messages}
            keyExtractor={(m) => m.id}
            isLoading={isLoading}
            emptyMessage="Belum ada pesan dari orang tua"
          />
          {messageMeta.total > 0 && (
            <div className="px-4 py-3 border-t border-gray-100">
              <Pagination
                currentPage={messagePage}
                totalPages={messageMeta.totalPages}
                total={messageMeta.total}
                limit={20}
                onPageChange={setMessagePage}
              />
            </div>
          )}
        </Card>
      )}
    </div>
  );
}
