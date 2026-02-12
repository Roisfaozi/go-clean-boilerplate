"use client";

import { createContext, useContext, useState, useCallback, useEffect, ReactNode } from "react";
import { toast } from "sonner";
import { accessApi, AccessRight, Endpoint } from "~/lib/api/access";

interface AccessRightsContextType {
  accessRights: AccessRight[];
  endpoints: Endpoint[];
  isLoading: boolean;
  isProcessing: string | null;
  isCreating: boolean;
  fetchData: () => Promise<void>;
  createAccessRight: (name: string, description: string) => Promise<void>;
  createEndpoint: (method: string, path: string) => Promise<void>;
  deleteAccessRight: (id: string) => Promise<void>;
  deleteEndpoint: (id: string) => Promise<void>;
  toggleLink: (accessRightId: string, endpointId: string, isLinked: boolean) => Promise<void>;
}

const AccessRightsContext = createContext<AccessRightsContextType | undefined>(undefined);

export function AccessRightsProvider({ children }: { children: ReactNode }) {
  const [accessRights, setAccessRights] = useState<AccessRight[]>([]);
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isProcessing, setIsProcessing] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [arResp, epResp] = await Promise.all([
        accessApi.getAllAccessRights(),
        accessApi.searchEndpoints({ page: 1, page_size: 1000 }),
      ]);
      if (arResp && arResp.data) setAccessRights(arResp.data.data);
      if (epResp && epResp.data) setEndpoints(epResp.data);
    } catch (error) {
      toast.error("Failed to fetch data");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const createAccessRight = useCallback(async (name: string, description: string) => {
    setIsCreating(true);
    try {
      await accessApi.createAccessRight(name, description);
      toast.success("Access Right created");
      await fetchData();
    } catch (error) {
      toast.error("Failed to create Access Right");
      throw error;
    } finally {
      setIsCreating(false);
    }
  }, [fetchData]);

  const createEndpoint = useCallback(async (method: string, path: string) => {
    setIsCreating(true);
    try {
      await accessApi.createEndpoint(method, path);
      toast.success("Endpoint registered");
      await fetchData();
    } catch (error) {
      toast.error("Failed to register Endpoint");
      throw error;
    } finally {
      setIsCreating(false);
    }
  }, [fetchData]);

  const deleteAccessRight = useCallback(async (id: string) => {
    setIsProcessing(id);
    try {
      await accessApi.deleteAccessRight(id);
      toast.success("Access Right deleted");
      await fetchData();
    } catch (error) {
      toast.error("Failed to delete Access Right");
    } finally {
      setIsProcessing(null);
    }
  }, [fetchData]);

  const deleteEndpoint = useCallback(async (id: string) => {
    setIsProcessing(id);
    try {
      await accessApi.deleteEndpoint(id);
      toast.success("Endpoint deleted");
      await fetchData();
    } catch (error) {
      toast.error("Failed to delete Endpoint");
    } finally {
      setIsProcessing(null);
    }
  }, [fetchData]);

  const toggleLink = useCallback(async (
    accessRightId: string,
    endpointId: string,
    isLinked: boolean
  ) => {
    const processId = `${accessRightId}-${endpointId}`;
    setIsProcessing(processId);
    try {
      if (isLinked) {
        await accessApi.unlinkEndpoint(accessRightId, endpointId);
        toast.success("Endpoint unlinked");
      } else {
        await accessApi.linkEndpoint(accessRightId, endpointId);
        toast.success("Endpoint linked");
      }
      await fetchData();
    } catch (error) {
      toast.error("Failed to update access right link");
    } finally {
      setIsProcessing(null);
    }
  }, [fetchData]);

  return (
    <AccessRightsContext.Provider
      value={{
        accessRights,
        endpoints,
        isLoading,
        isProcessing,
        isCreating,
        fetchData,
        createAccessRight,
        createEndpoint,
        deleteAccessRight,
        deleteEndpoint,
        toggleLink,
      }}
    >
      {children}
    </AccessRightsContext.Provider>
  );
}

export function useAccessRights() {
  const context = useContext(AccessRightsContext);
  if (context === undefined) {
    throw new Error("useAccessRights must be used within an AccessRightsProvider");
  }
  return context;
}
