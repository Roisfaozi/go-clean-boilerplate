import { useQuery } from "@tanstack/react-query";
import { auditService } from "./auditService";

const KEY = ["audit-logs"];

export function useAuditLogs(params: {
  page?: number;
  limit?: number;
  sort?: string;
  search?: string;
  filters?: any[];
}) {
  return useQuery({
    queryKey: [...KEY, params],
    queryFn: () => auditService.search(params),
  });
}

export function useExportAuditLogs() {
  return async (params: {
    from_date?: string;
    to_date?: string;
    format: "csv" | "excel";
  }) => {
    const response = await auditService.export(params);
    const url = window.URL.createObjectURL(new Blob([response as any]));
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute(
      "download",
      `audit-logs-${new Date().toISOString()}.${params.format}`,
    );
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  };
}
