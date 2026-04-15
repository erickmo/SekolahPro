import { apiClient } from './api.client'

export interface DashboardSummary {
  adminName: string
  userRole: string
  totalUsers: number
  activeEmployees: number
  transactionsToday: number
  revenueToday: number
  revenueGrowthPercent: number
  pendingApprovals: number
  systemHealth: string
  recentActivities: DashboardActivity[]
}

export interface DashboardActivity {
  id: string
  action: string
  actorName: string
  module: string
  timestamp: string
}

interface ApiDashboardSummary {
  admin_name: string
  user_role: string
  total_users: number
  active_employees: number
  transactions_today: number
  revenue_today: number
  revenue_growth_percent: number
  pending_approvals: number
  system_health: string
  recent_activities: Array<{
    id: string
    action: string
    actor_name: string
    module: string
    timestamp: string
  }>
}

function mapSummary(raw: ApiDashboardSummary): DashboardSummary {
  return {
    adminName: raw.admin_name,
    userRole: raw.user_role,
    totalUsers: raw.total_users,
    activeEmployees: raw.active_employees,
    transactionsToday: raw.transactions_today,
    revenueToday: raw.revenue_today,
    revenueGrowthPercent: raw.revenue_growth_percent,
    pendingApprovals: raw.pending_approvals,
    systemHealth: raw.system_health,
    recentActivities: (raw.recent_activities ?? []).map((a) => ({
      id: a.id,
      action: a.action,
      actorName: a.actor_name,
      module: a.module,
      timestamp: a.timestamp,
    })),
  }
}

export const dashboardService = {
  getSummary: async (): Promise<DashboardSummary> => {
    const resp = await apiClient.get<{ data: ApiDashboardSummary }>('/api/v1/dashboard/summary')
    return mapSummary(resp.data)
  },
}
