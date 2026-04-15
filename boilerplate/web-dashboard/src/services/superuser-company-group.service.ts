import { createEntityService } from './createEntityService'
import type { CompanyGroup, ApiCompanyGroup } from '@/types/entity.types'
import { mapApiCompanyGroup } from '@/types/entity.types'

export const superuserCompanyGroupService = createEntityService<CompanyGroup, ApiCompanyGroup>(
  '/api/v1/superuser/company-groups',
  mapApiCompanyGroup,
)
