import { createEntityService } from './createEntityService'
import type { TenantOwner, ApiTenantOwner } from '@/types/entity.types'
import { mapApiTenantOwner } from '@/types/entity.types'

export const tenantOwnerService = createEntityService<TenantOwner, ApiTenantOwner>(
  '/api/v1/superuser/tenant-owners',
  mapApiTenantOwner,
)
