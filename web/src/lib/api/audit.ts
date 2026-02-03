import { api } from "./client";

export interface AuditLog {
  id: string;
  user_id: string;
  username?: string; // Optional if joined in backend
  action: string;
  entity: string;
  entity_id: string;
  old_values: any;
  new_values: any;
  ip_address: string;
  user_agent: string;
  created_at: number;
}

export interface AuditLogListResponse {
  data: AuditLog[];
  paging: {
    total: number;
    page: number;
    size: number;
  };
}

export const auditApi = {
  search: (filter: any) => {
    return api.post<AuditLogListResponse>("/audit-logs/search", filter);
  },

  export: (fromDate?: string, toDate?: string) => {
    const params = new URLSearchParams();
    if (fromDate) params.append("from_date", fromDate);
    if (toDate) params.append("to_date", toDate);
    
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
    return `${baseUrl}/audit-logs/export?${params.toString()}`; 
  }
};
