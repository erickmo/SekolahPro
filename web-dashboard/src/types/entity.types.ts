/**
 * Shared base types for all entities.
 * Add your domain-specific entity types here as you build features.
 *
 * Example:
 *   export interface User extends BaseEntity {
 *     name: string
 *     email: string
 *   }
 */

export interface BaseEntity {
  id: string
  createdAt: string
  updatedAt: string
  createdBy?: string
}
