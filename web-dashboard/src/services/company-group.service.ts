import { createEntityService } from './createEntityService'
import type { CompanyGroup, ApiCompanyGroup } from '@/types/entity.types'
import { mapApiCompanyGroup } from '@/types/entity.types'

export const companyGroupService = createEntityService<CompanyGroup, ApiCompanyGroup>(
  '/api/v1/company-groups',
  mapApiCompanyGroup,
)
