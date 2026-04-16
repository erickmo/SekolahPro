import { createVernonService } from './vernon.service'
import type { ClassRoom } from '@/types/class-room.types'

function transformClassRoom(raw: Record<string, unknown>): ClassRoom {
  const ay = raw.academic_year as Record<string, unknown> | undefined
  const ht = raw.homeroom_teacher as Record<string, unknown> | undefined

  return {
    id: raw.id as string,
    name: (raw.name ?? '') as string,
    gradeLevel: (raw.grade_level ?? '') as string,
    parallelId: (raw.parallel_id ?? '') as string,
    capacity: (raw.capacity ?? 0) as number,
    currentCount: (raw.current_count ?? 0) as number,
    major: (raw.major ?? 'umum') as ClassRoom['major'],
    isActive: (raw.is_active ?? true) as boolean,
    academicYearId: (raw.academic_year_id ?? '') as string,
    homeroomTeacherId: (raw.homeroom_teacher_id ?? '') as string,
    academicYear: ay
      ? {
          name: (ay.name ?? '') as string,
          code: (ay.code ?? '') as string,
          isActive: (ay.is_active ?? false) as boolean,
        }
      : undefined,
    homeroomTeacher: ht
      ? {
          fullName: (ht.full_name ?? '') as string,
          nip: (ht.nip ?? '') as string,
          nuptk: (ht.nuptk ?? '') as string,
        }
      : undefined,
    createdAt: raw.created_at as string,
    updatedAt: raw.updated_at as string,
  }
}

export const classRoomService = createVernonService<ClassRoom, Record<string, unknown>>(
  '/class_rooms',
  transformClassRoom,
)
