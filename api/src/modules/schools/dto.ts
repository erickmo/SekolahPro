import { z } from 'zod';

export const CreateSchoolDto = z.object({
  name: z.string().min(3),
  npsn: z.string().length(8),
  subdomain: z.string().min(3).regex(/^[a-z0-9-]+$/),
  address: z.string().min(5),
  logoUrl: z.string().url().optional(),
  config: z.record(z.unknown()).optional(),
  tenantPlan: z.enum(['BASIC', 'PRO', 'ENTERPRISE']).optional(),
});

export const UpdateSchoolDto = CreateSchoolDto.partial().omit({ npsn: true, subdomain: true });

export type CreateSchoolInput = z.infer<typeof CreateSchoolDto>;
export type UpdateSchoolInput = z.infer<typeof UpdateSchoolDto>;
