"use client";

import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import {
  OrganizationSettings,
  organizationsApi,
  Organization,
} from "~/lib/api/organizations";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

interface SettingsContextType {
  name: string;
  setName: (name: string) => void;
  settings: OrganizationSettings;
  updateSetting: (key: keyof OrganizationSettings, value: any) => void;
  isLoading: boolean;
  isDeleting: boolean;
  hasChanges: boolean;
  handleUpdate: () => Promise<void>;
  handleDelete: () => Promise<void>;
  currentOrganization: Organization | null;
}

const SettingsContext = createContext<SettingsContextType | undefined>(
  undefined
);

export function SettingsProvider({ children }: { children: ReactNode }) {
  const { currentOrganization, setCurrentOrganization } =
    useOrganizationStore();
  const [name, setName] = useState(currentOrganization?.name || "");
  const [settings, setSettings] = useState<OrganizationSettings>(
    currentOrganization?.settings || {}
  );
  const [isLoading, setIsLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const router = useRouter();

  useEffect(() => {
    if (currentOrganization) {
      setName(currentOrganization.name);
      setSettings(currentOrganization.settings || {});
    }
  }, [currentOrganization]);

  const updateSetting = (key: keyof OrganizationSettings, value: any) => {
    setSettings((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const handleUpdate = async () => {
    if (!currentOrganization) return;
    setIsLoading(true);
    try {
      const resp = await organizationsApi.update(currentOrganization.id, {
        name,
        settings,
      });
      if (resp.data) {
        setCurrentOrganization(resp.data);
        toast.success("Organization updated successfully");
      }
    } catch (error: any) {
      toast.error(error.message || "Failed to update organization");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!currentOrganization) return;
    if (
      !confirm(
        `Are you sure you want to permanently delete ${currentOrganization.name}? This action cannot be undone.`
      )
    )
      return;

    setIsDeleting(true);
    try {
      await organizationsApi.delete(currentOrganization.id);
      toast.success("Organization deleted successfully");
      setCurrentOrganization(null);
      router.push("/dashboard");
    } catch (error: any) {
      toast.error(error.message || "Failed to delete organization");
    } finally {
      setIsDeleting(false);
    }
  };

  const hasChanges =
    name !== currentOrganization?.name ||
    JSON.stringify(settings) !==
      JSON.stringify(currentOrganization?.settings || {});

  return (
    <SettingsContext.Provider
      value={{
        name,
        setName,
        settings,
        updateSetting,
        isLoading,
        isDeleting,
        hasChanges,
        handleUpdate,
        handleDelete,
        currentOrganization,
      }}
    >
      {children}
    </SettingsContext.Provider>
  );
}

export function useSettings() {
  const context = useContext(SettingsContext);
  if (context === undefined) {
    throw new Error("useSettings must be used within a SettingsProvider");
  }
  return context;
}
