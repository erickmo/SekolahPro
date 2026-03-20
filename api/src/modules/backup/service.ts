import { prisma } from '../../lib/prisma';
import { NotFoundError } from '../../shared/errors';

interface BackupRecord {
  id: string;
  schoolId: string;
  createdAt: string;
  size: string;
  status: 'COMPLETED' | 'IN_PROGRESS' | 'FAILED';
  type: 'FULL' | 'INCREMENTAL';
}

// In-memory store for simulated backup metadata (per school)
const backupStore: Map<string, BackupRecord[]> = new Map();

function getSchoolBackups(schoolId: string): BackupRecord[] {
  return backupStore.get(schoolId) ?? [];
}

function estimateSize(): string {
  const mb = Math.floor(Math.random() * 450 + 50);
  return `~${mb} MB`;
}

export async function createBackup(schoolId: string): Promise<BackupRecord> {
  // Verify school exists
  const school = await prisma.school.findUnique({ where: { id: schoolId } });
  if (!school) throw new NotFoundError('Sekolah');

  const backup: BackupRecord = {
    id: crypto.randomUUID(),
    schoolId,
    createdAt: new Date().toISOString(),
    size: estimateSize(),
    status: 'COMPLETED',
    type: 'FULL',
  };

  const existing = getSchoolBackups(schoolId);
  backupStore.set(schoolId, [backup, ...existing]);

  return backup;
}

export async function listBackups(schoolId: string): Promise<BackupRecord[]> {
  return getSchoolBackups(schoolId);
}

export async function restoreBackup(backupId: string) {
  // Search across all schools for the backup ID
  for (const backups of backupStore.values()) {
    const found = backups.find((b) => b.id === backupId);
    if (found) {
      return {
        backupId,
        status: 'INITIATED',
        message: 'Restore dimulai, estimasi 5-10 menit',
        backup: found,
        initiatedAt: new Date().toISOString(),
      };
    }
  }
  throw new NotFoundError('Backup');
}

export async function getBackupStatus(schoolId: string) {
  const backups = getSchoolBackups(schoolId);

  const lastBackup = backups.length > 0 ? backups[0] : null;

  // Schedule next backup 24h after last, or tomorrow if none
  const nextScheduledDate = new Date();
  if (lastBackup) {
    nextScheduledDate.setTime(new Date(lastBackup.createdAt).getTime() + 24 * 60 * 60 * 1000);
  } else {
    nextScheduledDate.setDate(nextScheduledDate.getDate() + 1);
    nextScheduledDate.setHours(2, 0, 0, 0);
  }

  const totalSizeMb = backups.reduce((sum, b) => {
    const match = b.size.match(/(\d+)/);
    return sum + (match ? parseInt(match[1], 10) : 0);
  }, 0);

  return {
    lastBackup: lastBackup ? lastBackup.createdAt : null,
    nextScheduled: nextScheduledDate.toISOString(),
    totalBackups: backups.length,
    storageUsed: `~${totalSizeMb} MB`,
  };
}
