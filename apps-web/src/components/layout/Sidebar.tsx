'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { useAuth } from '@/context/AuthContext';
import { canAccess } from '@/lib/auth';
import {
  LayoutDashboard, Users, GraduationCap, BookOpen, CreditCard, Library,
  PiggyBank, Heart, MessageSquare, ChevronDown, ChevronRight, School,
  Bell, AlertTriangle, Building2, Shield, Database, Globe, Bot,
  BarChart3, Wifi, Trophy, ShoppingBag, Briefcase, FileText, ClipboardList,
  Bus, Wrench, Shirt, Calendar, Leaf, UserCheck, BookMarked, Stethoscope,
  Mic, Radio, DollarSign, Gift, Users2, HandHeart, Target, Banknote,
  Map, Star, Package, Archive, Activity, Lock, HardDrive,
} from 'lucide-react';

interface NavItem {
  label: string;
  href?: string;
  icon: React.ReactNode;
  module?: string;
  roles?: string[];
  children?: NavItem[];
}

interface NavGroup {
  label: string;
  items: NavItem[];
}

const NAV_GROUPS: NavGroup[] = [
  {
    label: '',
    items: [
      { label: 'Dashboard', href: '/dashboard', icon: <LayoutDashboard className="w-4 h-4" />, module: 'dashboard' },
    ],
  },
  {
    label: 'Siswa',
    items: [
      { label: 'Data Siswa', href: '/dashboard/students', icon: <Users className="w-4 h-4" />, module: 'students' },
      { label: 'Kartu Siswa', href: '/dashboard/student-cards', icon: <CreditCard className="w-4 h-4" />, module: 'student-cards' },
      { label: 'PPDB Online', href: '/dashboard/ppdb', icon: <ClipboardList className="w-4 h-4" />, module: 'ppdb' },
      { label: 'Portofolio', href: '/dashboard/portfolio', icon: <Star className="w-4 h-4" />, module: 'portfolio' },
    ],
  },
  {
    label: 'Akademik',
    items: [
      { label: 'Akademik & Kurikulum', href: '/dashboard/academic', icon: <GraduationCap className="w-4 h-4" />, module: 'academic' },
      { label: 'Ujian & Soal AI', href: '/dashboard/exams', icon: <FileText className="w-4 h-4" />, module: 'exams' },
      { label: 'RPP Digital', href: '/dashboard/lesson-plans', icon: <BookMarked className="w-4 h-4" />, module: 'lesson-plans' },
      { label: 'Penilaian (KKTP)', href: '/dashboard/assessments', icon: <Target className="w-4 h-4" />, module: 'assessments' },
      { label: 'Remedial', href: '/dashboard/remediation', icon: <Activity className="w-4 h-4" />, module: 'remediation' },
      { label: 'Learning Path', href: '/dashboard/learning-path', icon: <Map className="w-4 h-4" />, module: 'learning-path' },
      { label: 'Ekstrakurikuler', href: '/dashboard/extracurricular', icon: <Trophy className="w-4 h-4" />, module: 'extracurricular' },
    ],
  },
  {
    label: 'Komunikasi',
    items: [
      { label: 'Portal Orang Tua', href: '/dashboard/parent-portal', icon: <Users2 className="w-4 h-4" />, module: 'parent-portal' },
      { label: 'Notifikasi & WA', href: '/dashboard/notifications', icon: <Bell className="w-4 h-4" />, module: 'notifications' },
      { label: 'Forum & Komunitas', href: '/dashboard/forum', icon: <MessageSquare className="w-4 h-4" />, module: 'forum' },
      { label: 'Rapat & Video', href: '/dashboard/meetings', icon: <Mic className="w-4 h-4" />, module: 'meetings' },
      { label: 'Blog Sekolah', href: '/dashboard/school-blog', icon: <Radio className="w-4 h-4" />, module: 'school-blog' },
      { label: 'Pengumuman', href: '/dashboard/announcements', icon: <Bell className="w-4 h-4" />, roles: ['ADMIN_SEKOLAH', 'KEPALA_SEKOLAH', 'TATA_USAHA'] },
    ],
  },
  {
    label: 'Keuangan',
    items: [
      { label: 'Pembayaran SPP', href: '/dashboard/payments', icon: <CreditCard className="w-4 h-4" />, module: 'payments' },
      { label: 'Kantin Digital', href: '/dashboard/canteen', icon: <ShoppingBag className="w-4 h-4" />, module: 'canteen' },
      { label: 'Perpustakaan', href: '/dashboard/library', icon: <Library className="w-4 h-4" />, module: 'library' },
      { label: 'Manajemen Aset', href: '/dashboard/assets', icon: <Package className="w-4 h-4" />, module: 'assets' },
      { label: 'Beasiswa', href: '/dashboard/scholarships', icon: <Gift className="w-4 h-4" />, module: 'scholarships' },
      { label: 'Anggaran (RKAS)', href: '/dashboard/budget', icon: <Banknote className="w-4 h-4" />, module: 'budget' },
      { label: 'Dana Donasi', href: '/dashboard/donations', icon: <HandHeart className="w-4 h-4" />, module: 'donations' },
      {
        label: 'Koperasi', icon: <PiggyBank className="w-4 h-4" />, module: 'cooperative',
        children: [
          { label: 'Tabungan Siswa', href: '/dashboard/cooperative', icon: <PiggyBank className="w-3.5 h-3.5" />, module: 'cooperative' },
          { label: 'Produk', href: '/dashboard/cooperative/products', icon: <BookOpen className="w-3.5 h-3.5" />, module: 'cooperative' },
        ],
      },
    ],
  },
  {
    label: 'AI & Analitik',
    items: [
      { label: 'Early Warning', href: '/dashboard/ews', icon: <AlertTriangle className="w-4 h-4" />, module: 'ews' },
      { label: 'AI Tutor', href: '/dashboard/ai-tutor', icon: <Bot className="w-4 h-4" />, module: 'ai-tutor' },
      { label: 'Analitik', href: '/dashboard/analytics', icon: <BarChart3 className="w-4 h-4" />, module: 'analytics' },
      { label: 'Analisis Sentimen', href: '/dashboard/sentiment', icon: <Activity className="w-4 h-4" />, module: 'sentiment' },
      { label: 'Data Warehouse', href: '/dashboard/warehouse', icon: <Database className="w-4 h-4" />, module: 'warehouse' },
    ],
  },
  {
    label: 'Kesehatan & Wellbeing',
    items: [
      { label: 'Kesehatan (UKS)', href: '/dashboard/health', icon: <Heart className="w-4 h-4" />, module: 'health' },
      { label: 'Konseling BK', href: '/dashboard/counseling', icon: <MessageSquare className="w-4 h-4" />, module: 'counseling' },
      { label: 'Anti-Bullying', href: '/dashboard/anti-bullying', icon: <Shield className="w-4 h-4" />, module: 'anti-bullying' },
      { label: 'Pemantauan Gizi', href: '/dashboard/nutrition', icon: <Leaf className="w-4 h-4" />, module: 'nutrition' },
      { label: 'Manajemen ABK', href: '/dashboard/special-needs', icon: <Stethoscope className="w-4 h-4" />, module: 'special-needs' },
    ],
  },
  {
    label: 'Operasional',
    items: [
      { label: 'SDM & Presensi Guru', href: '/dashboard/hr', icon: <UserCheck className="w-4 h-4" />, module: 'hr' },
      { label: 'Smart Gate', href: '/dashboard/smart-gate', icon: <Lock className="w-4 h-4" />, module: 'smart-gate' },
      { label: 'Booking Ruang', href: '/dashboard/room-booking', icon: <Building2 className="w-4 h-4" />, module: 'room-booking' },
      { label: 'Transportasi', href: '/dashboard/transport', icon: <Bus className="w-4 h-4" />, module: 'transport' },
      { label: 'Helpdesk', href: '/dashboard/helpdesk', icon: <Wrench className="w-4 h-4" />, module: 'helpdesk' },
      { label: 'Seragam', href: '/dashboard/uniforms', icon: <Shirt className="w-4 h-4" />, module: 'uniforms' },
      { label: 'Event & Kepanitiaan', href: '/dashboard/events', icon: <Calendar className="w-4 h-4" />, module: 'events' },
      { label: 'IoT & Smart Campus', href: '/dashboard/iot', icon: <Wifi className="w-4 h-4" />, module: 'iot' },
      { label: 'Volunteer', href: '/dashboard/volunteer', icon: <HandHeart className="w-4 h-4" />, module: 'volunteer' },
      { label: 'Komite Sekolah', href: '/dashboard/committee', icon: <Users2 className="w-4 h-4" />, module: 'committee' },
    ],
  },
  {
    label: 'Karir & Alumni',
    items: [
      { label: 'Alumni', href: '/dashboard/alumni', icon: <GraduationCap className="w-4 h-4" />, module: 'alumni' },
      { label: 'Prakerin/Magang', href: '/dashboard/internship', icon: <Briefcase className="w-4 h-4" />, module: 'internship' },
      { label: 'Pelatihan Guru', href: '/dashboard/teacher-training', icon: <BookOpen className="w-4 h-4" />, module: 'teacher-training' },
    ],
  },
  {
    label: 'Gamifikasi & Marketplace',
    items: [
      { label: 'Gamifikasi', href: '/dashboard/gamification', icon: <Trophy className="w-4 h-4" />, module: 'gamification' },
      { label: 'Marketplace Konten', href: '/dashboard/edu-marketplace', icon: <ShoppingBag className="w-4 h-4" />, module: 'edu-marketplace' },
      { label: 'Guru Les / Bimbel', href: '/dashboard/tutoring', icon: <BookOpen className="w-4 h-4" />, module: 'tutoring' },
    ],
  },
  {
    label: 'Integrasi',
    items: [
      { label: 'Dapodik/EMIS', href: '/dashboard/dapodik', icon: <Archive className="w-4 h-4" />, module: 'dapodik' },
      { label: 'LMS Integration', href: '/dashboard/lms', icon: <GraduationCap className="w-4 h-4" />, module: 'lms' },
      { label: 'Website Sekolah', href: '/dashboard/website', icon: <Globe className="w-4 h-4" />, module: 'website' },
      { label: 'Stakeholder', href: '/dashboard/stakeholders', icon: <Users className="w-4 h-4" />, module: 'stakeholders' },
    ],
  },
  {
    label: 'SaaS Platform',
    items: [
      { label: 'Tenant & Langganan', href: '/dashboard/tenant', icon: <School className="w-4 h-4" />, roles: ['EDS_SUPERADMIN', 'EDS_SALES', 'EDS_SUPPORT'] },
      { label: 'Listing Vendor', href: '/dashboard/listing', icon: <Map className="w-4 h-4" />, roles: ['EDS_SUPERADMIN', 'EDS_SALES'] },
      { label: 'Dashboard Yayasan', href: '/dashboard/foundation', icon: <Building2 className="w-4 h-4" />, roles: ['YAYASAN'] },
    ],
  },
  {
    label: 'Infrastruktur',
    items: [
      { label: 'Audit Log', href: '/dashboard/audit-log', icon: <FileText className="w-4 h-4" />, roles: ['ADMIN_SEKOLAH', 'EDS_SUPERADMIN'] },
      { label: 'Security Center', href: '/dashboard/security', icon: <Shield className="w-4 h-4" />, roles: ['ADMIN_SEKOLAH', 'EDS_SUPERADMIN'] },
      { label: 'Backup & Recovery', href: '/dashboard/backup', icon: <HardDrive className="w-4 h-4" />, roles: ['EDS_SUPERADMIN', 'ADMIN_SEKOLAH'] },
    ],
  },
];

interface SidebarProps {
  collapsed?: boolean;
}

export function Sidebar({ collapsed = false }: SidebarProps) {
  const pathname = usePathname();
  const { user } = useAuth();
  const [openMenus, setOpenMenus] = useState<Set<string>>(new Set());

  const toggleMenu = (label: string) => {
    setOpenMenus((prev) => {
      const next = new Set(prev);
      next.has(label) ? next.delete(label) : next.add(label);
      return next;
    });
  };

  const isVisible = (item: NavItem): boolean => {
    if (!user) return false;
    if (item.roles) return item.roles.includes(user.role);
    if (item.module) return canAccess(user, item.module);
    return true;
  };

  const isActive = (href: string): boolean => {
    if (href === '/dashboard') return pathname === '/dashboard';
    return pathname.startsWith(href);
  };

  const visibleGroups = NAV_GROUPS.map((group) => ({
    ...group,
    items: group.items.filter(isVisible),
  })).filter((group) => group.items.length > 0);

  return (
    <aside
      className={cn(
        'flex flex-col bg-gray-900 text-white transition-all duration-300 overflow-hidden',
        collapsed ? 'w-16' : 'w-64'
      )}
    >
      {/* Logo */}
      <div className="flex items-center gap-3 px-4 py-5 border-b border-gray-800 flex-shrink-0">
        <div className="flex-shrink-0 w-8 h-8 bg-primary-500 rounded-lg flex items-center justify-center">
          <School className="w-5 h-5 text-white" />
        </div>
        {!collapsed && (
          <div className="min-w-0">
            <p className="text-sm font-bold text-white truncate">
              {user?.schoolName || 'EDS Platform'}
            </p>
            <p className="text-xs text-gray-400 truncate">
              {user?.schoolSubdomain || 'ekosistem digital'}
            </p>
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-3 px-2 space-y-4">
        {visibleGroups.map((group) => (
          <div key={group.label}>
            {group.label && !collapsed && (
              <p className="px-3 mb-1 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                {group.label}
              </p>
            )}
            <div className="space-y-0.5">
              {group.items.map((item) => (
                <NavEntry
                  key={item.label}
                  item={item}
                  collapsed={collapsed}
                  isActive={isActive}
                  isOpen={openMenus.has(item.label)}
                  onToggle={() => toggleMenu(item.label)}
                  isVisible={isVisible}
                />
              ))}
            </div>
          </div>
        ))}
      </nav>

      {/* User info at bottom */}
      {!collapsed && user && (
        <div className="px-4 py-4 border-t border-gray-800 flex-shrink-0">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary-500 flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div className="min-w-0">
              <p className="text-sm font-medium text-white truncate">{user.name}</p>
              <p className="text-xs text-gray-400 truncate">{user.role}</p>
            </div>
          </div>
        </div>
      )}
    </aside>
  );
}

interface NavEntryProps {
  item: NavItem;
  collapsed: boolean;
  isActive: (href: string) => boolean;
  isOpen: boolean;
  onToggle: () => void;
  isVisible: (item: NavItem) => boolean;
}

function NavEntry({ item, collapsed, isActive, isOpen, onToggle, isVisible }: NavEntryProps) {
  if (item.children) {
    const visibleChildren = item.children.filter(isVisible);
    if (visibleChildren.length === 0) return null;

    return (
      <div>
        <button
          onClick={onToggle}
          className={cn(
            'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm',
            'text-gray-300 hover:bg-gray-800 hover:text-white transition-colors',
          )}
        >
          <span className="flex-shrink-0">{item.icon}</span>
          {!collapsed && (
            <>
              <span className="flex-1 text-left">{item.label}</span>
              {isOpen ? <ChevronDown className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
            </>
          )}
        </button>
        {isOpen && !collapsed && (
          <div className="ml-4 border-l border-gray-700 pl-2 mt-0.5">
            {visibleChildren.map((child) => (
              <Link
                key={child.label}
                href={child.href!}
                className={cn(
                  'flex items-center gap-3 px-3 py-2 rounded-lg text-sm mb-0.5',
                  'transition-colors',
                  isActive(child.href!)
                    ? 'bg-primary-600 text-white'
                    : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                )}
              >
                <span className="flex-shrink-0">{child.icon}</span>
                <span>{child.label}</span>
              </Link>
            ))}
          </div>
        )}
      </div>
    );
  }

  if (!item.href) return null;

  return (
    <Link
      href={item.href}
      title={collapsed ? item.label : undefined}
      className={cn(
        'flex items-center gap-3 px-3 py-2 rounded-lg text-sm',
        'transition-colors',
        isActive(item.href)
          ? 'bg-primary-600 text-white'
          : 'text-gray-300 hover:bg-gray-800 hover:text-white'
      )}
    >
      <span className="flex-shrink-0">{item.icon}</span>
      {!collapsed && <span>{item.label}</span>}
    </Link>
  );
}
