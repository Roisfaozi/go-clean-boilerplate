"use client";

import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  ReactNode,
} from "react";
import { auditApi, AuditLog } from "~/lib/api/audit";
import { toast } from "sonner";
import { useAuditStream } from "~/hooks/use-audit-stream";

interface AuditContextType {
  logs: AuditLog[];
  isLoading: boolean;
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  page: number;
  setPage: (page: number) => void;
  totalItems: number;
  pageSize: number;
  selectedLog: AuditLog | null;
  setSelectedLog: (log: AuditLog | null) => void;
  isDetailOpen: boolean;
  setIsDetailOpen: (open: boolean) => void;
  fetchLogs: () => Promise<void>;
  handleRowClick: (log: AuditLog) => void;
}

const AuditContext = createContext<AuditContextType | undefined>(undefined);

export function AuditProvider({ children }: { children: ReactNode }) {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [page, setPage] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const pageSize = 15;

  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [isDetailOpen, setIsDetailOpen] = useState(false);

  // Real-time listener
  const newLog = useAuditStream();

  useEffect(() => {
    if (newLog && page === 1) {
      setLogs((prev) => [newLog, ...prev.slice(0, pageSize - 1)]);
      setTotalItems((prev) => prev + 1);
    }
  }, [newLog, page]);

  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    try {
      const filter: any = {
        page: page,
        page_size: pageSize,
        sort: [{ colId: "created_at", sort: "desc" }],
      };

      if (searchTerm) {
        filter.filter = {
          action: { type: "contains", filter: searchTerm },
        };
      }

      const response = await auditApi.search(filter);
      if (response && response.data) {
        setLogs(response.data);
        setTotalItems(response.paging?.total || 0);
      } else {
        setLogs([]);
        setTotalItems(0);
      }
    } catch (error) {
      console.error("Failed to fetch audit logs:", error);
      toast.error("Failed to fetch audit logs");
    } finally {
      setIsLoading(false);
    }
  }, [page, searchTerm]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const handleRowClick = useCallback((log: AuditLog) => {
    setSelectedLog(log);
    setIsDetailOpen(true);
  }, []);

  return (
    <AuditContext.Provider
      value={{
        logs,
        isLoading,
        searchTerm,
        setSearchTerm,
        page,
        setPage,
        totalItems,
        pageSize,
        selectedLog,
        setSelectedLog,
        isDetailOpen,
        setIsDetailOpen,
        fetchLogs,
        handleRowClick,
      }}
    >
      {children}
    </AuditContext.Provider>
  );
}

export function useAudit() {
  const context = useContext(AuditContext);
  if (context === undefined) {
    throw new Error("useAudit must be used within an AuditProvider");
  }
  return context;
}
