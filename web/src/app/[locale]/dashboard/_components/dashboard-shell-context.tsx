"use client";

import {
  createContext,
  useContext,
  useCallback,
  ReactNode,
} from "react";
import { Organization, organizationsApi } from "~/lib/api/organizations";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";
import useSWR from "swr";

interface DashboardShellContextType {
  organizations: Organization[];
  currentOrganization: Organization | null;
  isLoading: boolean;
  setOrganization: (org: Organization) => void;
  refreshOrganizations: () => Promise<void>;
}

const DashboardShellContext = createContext<
  DashboardShellContextType | undefined
>(undefined);

export function DashboardShellProvider({ 
  children,
  initialData
}: { 
  children: ReactNode;
  initialData?: Organization[];
}) {
  const { currentOrganization, setCurrentOrganization } =
    useOrganizationStore();

  const {
    data: organizations = [],
    isLoading,
    mutate,
  } = useSWR("/api/v1/organizations/me", () => 
    organizationsApi.getMyOrganizations().then(res => res.data?.organizations || []),
    {
      fallbackData: initialData,
      keepPreviousData: true,
      onSuccess: (data) => {
        // Auto-select first org if none selected and we have data
        if (!currentOrganization && data.length > 0) {
          setCurrentOrganization(data[0]);
        }
      }
    }
  );

  const fetchOrgs = useCallback(async () => {
    await mutate();
  }, [mutate]);

  const setOrganization = useCallback(
    (org: Organization) => {
      setCurrentOrganization(org);
      toast.success(`Switched to ${org.name}`);
    },
    [setCurrentOrganization]
  );

  return (
    <DashboardShellContext.Provider
      value={{
        organizations,
        currentOrganization,
        isLoading,
        setOrganization,
        refreshOrganizations: fetchOrgs,
      }}
    >
      {children}
    </DashboardShellContext.Provider>
  );
}

export function useDashboardShell() {
  const context = useContext(DashboardShellContext);
  if (context === undefined) {
    throw new Error(
      "useDashboardShell must be used within a DashboardShellProvider"
    );
  }
  return context;
}
