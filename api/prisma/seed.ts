import { PrismaClient, Role } from '@prisma/client';
import bcrypt from 'bcryptjs';

const prisma = new PrismaClient();

async function main() {
  console.log('🌱 Seeding database...');

  // Create demo school
  const school = await prisma.school.upsert({
    where: { npsn: '12345678' },
    update: {},
    create: {
      name: 'SMA Demo EDS',
      npsn: '12345678',
      subdomain: 'demo',
      address: 'Jl. Pendidikan No. 1, Jakarta',
      config: { timezone: 'Asia/Jakarta', language: 'id', currency: 'IDR' },
      tenantPlan: 'PRO',
    },
  });
  console.log('✅ School created:', school.name);

  const hashedPassword = await bcrypt.hash('password123', 10);

  // Create users with different roles
  const users = [
    { email: 'superadmin@eds.id', name: 'Super Admin', role: Role.SUPERADMIN, schoolId: null },
    { email: 'admin@demo.eds.id', name: 'Admin Sekolah', role: Role.ADMIN_SEKOLAH, schoolId: school.id },
    { email: 'kepala@demo.eds.id', name: 'Kepala Sekolah', role: Role.KEPALA_SEKOLAH, schoolId: school.id },
    { email: 'kurikulum@demo.eds.id', name: 'Kepala Kurikulum', role: Role.KEPALA_KURIKULUM, schoolId: school.id },
    { email: 'bendahara@demo.eds.id', name: 'Bendahara', role: Role.BENDAHARA, schoolId: school.id },
    { email: 'guru@demo.eds.id', name: 'Guru Demo', role: Role.GURU, schoolId: school.id },
    { email: 'waliskelas@demo.eds.id', name: 'Wali Kelas', role: Role.WALI_KELAS, schoolId: school.id },
    { email: 'bk@demo.eds.id', name: 'Guru BK', role: Role.GURU_BK, schoolId: school.id },
    { email: 'operator@demo.eds.id', name: 'Operator SIMS', role: Role.OPERATOR_SIMS, schoolId: school.id },
    { email: 'kasir@demo.eds.id', name: 'Kasir Koperasi', role: Role.KASIR_KOPERASI, schoolId: school.id },
    { email: 'pustakawan@demo.eds.id', name: 'Pustakawan', role: Role.PUSTAKAWAN, schoolId: school.id },
    { email: 'uks@demo.eds.id', name: 'Petugas UKS', role: Role.PETUGAS_UKS, schoolId: school.id },
    { email: 'tu@demo.eds.id', name: 'Tata Usaha', role: Role.TATA_USAHA, schoolId: school.id },
  ];

  for (const userData of users) {
    await prisma.user.upsert({
      where: { email: userData.email },
      update: {},
      create: { ...userData, password: hashedPassword, isActive: true },
    });
  }
  console.log(`✅ ${users.length} users created`);

  // Create academic year
  const academicYear = await prisma.academicYear.upsert({
    where: { schoolId_name: { schoolId: school.id, name: '2025/2026' } },
    update: {},
    create: { schoolId: school.id, name: '2025/2026', startDate: new Date('2025-07-14'), endDate: new Date('2026-06-30'), isCurrent: true },
  });

  // Create semesters
  const sem1 = await prisma.semester.upsert({
    where: { academicYearId_name: { academicYearId: academicYear.id, name: 'Semester 1' } },
    update: {},
    create: { academicYearId: academicYear.id, name: 'Semester 1', startDate: new Date('2025-07-14'), endDate: new Date('2025-12-20'), isCurrent: true },
  });
  console.log('✅ Academic year & semester created');

  // Create subjects
  const subjects = ['Matematika', 'Bahasa Indonesia', 'Bahasa Inggris', 'Fisika', 'Kimia', 'Biologi', 'Sejarah', 'PPKn', 'Olahraga', 'Seni Budaya'];
  for (const [i, name] of subjects.entries()) {
    await prisma.subject.upsert({
      where: { code_schoolId: { code: `SUBJ-${i + 1}`, schoolId: school.id } },
      update: {},
      create: { schoolId: school.id, name, code: `SUBJ-${i + 1}`, gradeLevel: '10', curriculum: 'MERDEKA', weeklyHours: 3, isActive: true },
    });
  }
  console.log('✅ Subjects created');

  // Create demo teacher
  const teacher = await prisma.teacher.upsert({
    where: { email_schoolId: { email: 'guru@demo.eds.id', schoolId: school.id } },
    update: {},
    create: { schoolId: school.id, name: 'Guru Demo', email: 'guru@demo.eds.id', phone: '081234567890', nip: '197001012000011001', subjects: ['Matematika'], isActive: true },
  });

  // Create demo class
  const schoolClass = await prisma.schoolClass.upsert({
    where: { schoolId_name_academicYearId: { schoolId: school.id, name: 'X-A', academicYearId: academicYear.id } },
    update: {},
    create: { schoolId: school.id, name: 'X-A', gradeLevel: '10', academicYearId: academicYear.id, homeroomTeacherId: teacher.id, capacity: 36 },
  });
  console.log('✅ Teacher & class created');

  // Create demo students
  const studentsData = [
    { nisn: '0012345601', name: 'Ahmad Fauzi', gender: 'MALE' as const },
    { nisn: '0012345602', name: 'Siti Rahmah', gender: 'FEMALE' as const },
    { nisn: '0012345603', name: 'Budi Santoso', gender: 'MALE' as const },
    { nisn: '0012345604', name: 'Dewi Lestari', gender: 'FEMALE' as const },
    { nisn: '0012345605', name: 'Eko Prasetyo', gender: 'MALE' as const },
  ];

  for (const [i, s] of studentsData.entries()) {
    const student = await prisma.student.upsert({
      where: { nisn: s.nisn },
      update: {},
      create: { ...s, schoolId: school.id, birthDate: new Date(`2008-0${i + 1}-15`), riskScore: 0 },
    });

    // Enroll student
    await prisma.enrollment.upsert({
      where: { studentId_academicYearId: { studentId: student.id, academicYearId: academicYear.id } },
      update: {},
      create: { studentId: student.id, classId: schoolClass.id, academicYearId: academicYear.id, schoolId: school.id, status: 'ACTIVE' },
    });

    // Create savings account
    await prisma.savingsAccount.upsert({
      where: { studentId: student.id },
      update: {},
      create: { studentId: student.id, schoolId: school.id, balance: 50000 + i * 10000, accountNumber: `SAV-${Date.now()}-${student.id.slice(-4)}` },
    });
  }
  console.log('✅ 5 demo students created with enrollments & savings accounts');

  // Create notification template
  await prisma.notificationTemplate.upsert({
    where: { schoolId_name: { schoolId: school.id, name: 'attendance-absence' } },
    update: {},
    create: { schoolId: school.id, name: 'attendance-absence', type: 'ATTENDANCE', channel: 'WHATSAPP', titleTemplate: 'Pemberitahuan Ketidakhadiran', bodyTemplate: 'Yth. Orang Tua/Wali {{nama_siswa}}, siswa tidak hadir pada {{tanggal}}. Mohon konfirmasi.' },
  });
  console.log('✅ Notification template created');

  console.log('🎉 Seed completed! Demo credentials:');
  console.log('   Admin: admin@demo.eds.id / password123');
  console.log('   Kepala Sekolah: kepala@demo.eds.id / password123');
  console.log('   Guru: guru@demo.eds.id / password123');
}

main()
  .catch((e) => { console.error(e); process.exit(1); })
  .finally(async () => { await prisma.$disconnect(); });
