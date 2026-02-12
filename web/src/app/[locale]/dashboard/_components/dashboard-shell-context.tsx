"use client";

import { createContext, useContext, useState, useCallback, useEffect, ReactNode } from "react";
import { Organization, organizationsApi } from "~/lib/api/organizations";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";

interface DashboardShellContextType {
  organizations: Organization[];
  currentOrganization: Organization | null;
  isLoading: boolean;
  setOrganization: (org: Organization) => void;
  refreshOrganizations: () => Promise<void>;
}

const DashboardShellContext = createContext<DashboardShellContextType | undefined>(undefined);

export function DashboardShellProvider({ children }: { children: ReactNode }) {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const { currentOrganization, setCurrentOrganization } = useOrganizationStore();
  const [isLoading, setIsLoading] = useState(true);

  const fetchOrgs = useCallback(async () => {
    setIsLoading(true);
    try {
      const resp = await organizationsApi.getMyOrganizations();
      if (resp.data?.organizations) {
        setOrganizations(resp.data.organizations);
        // Sync with store if needed or auto-select
        if (!currentOrganization && resp.data.organizations.length > 0) {
          setCurrentOrganization(resp.data.organizations[0]);
        }
      }
    } catch (error) {
      console.error("Failed to fetch organizations", error);
      toast.error("Failed to load workspaces");
    } finally {
      setIsLoading(false);
    }
  }, [currentOrganization, setCurrentOrganization]);

  useEffect(() => {
    fetchOrgs();
  }, [fetchOrgs]);

  const setOrganization = useCallback((org: Organization) => {
    setCurrentOrganization(org);
    toast.success(`Switched to ${org.name}`);
  }, [setCurrentOrganization]);

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
    throw new Error("useDashboardShell must be used within a DashboardShellProvider");
  }
  return context;
}
