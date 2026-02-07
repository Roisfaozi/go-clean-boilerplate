import { api } from "./client";

export interface DashboardSummary {
  total_users: number;
  total_roles: number;
  total_audit_logs: number;
  total_org_members: number;
}

export interface ActivityPoint {
  date: string;
  audits: number;
  logins: number;
}

export interface DashboardActivity {
  points: ActivityPoint[];
}

export interface SystemInsights {
  avg_latency_ms: number;
  error_rate: number;
  uptime: string;
  most_active_role: string;
}

export const statsApi = {
  getSummary: () => 
    api.get<{ data: DashboardSummary }>("/stats/summary").then(res => res.data),
    
  getActivity: (days: number = 7) => 
    api.get<{ data: DashboardActivity }>(`/stats/activity?days=${days}`).then(res => res.data),
    
  getInsights: () => 
    api.get<{ data: SystemInsights }>("/stats/insights").then(res => res.data),
};