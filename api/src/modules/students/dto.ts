import { z } from 'zod';

export const CreateStudentDto = z.object({
  nisn: z.string().length(10),
  nik: z.string().length(16).optional(),
  name: z.string().min(2),
  birthDate: z.string(),
  gender: z.enum(['MALE', 'FEMALE']),
  address: z.string().optional(),
  religion: z.string().optional(),
  photoUrl: z.string().url().optional(),
  guardians: z.array(z.object({
    name: z.string().min(2),
    relationship: z.string(),
    phone: z.string(),
    email: z.string().email().optional(),
    occupation: z.string().optional(),
    address: z.string().optional(),
  })).optional(),
});

export const UpdateStudentDto = CreateStudentDto.partial();

export const StudentQueryDto = z.object({
  page: z.string().optional(),
  limit: z.string().optional(),
  search: z.string().optional(),
  classId: z.string().optional(),
  gender: z.enum(['MALE', 'FEMALE']).optional(),
});

export const ImportStudentsDto = z.object({
  students: z.array(CreateStudentDto),
});

export type CreateStudentInput = z.infer<typeof CreateStudentDto>;
export type UpdateStudentInput = z.infer<typeof UpdateStudentDto>;
