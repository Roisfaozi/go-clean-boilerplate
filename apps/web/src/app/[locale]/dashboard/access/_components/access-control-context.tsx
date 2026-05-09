"use client";

import { createContext, useContext, useState, useCallback, ReactNode } from "react";
import { Role } from "~/lib/api/roles";

interface AccessControlContextType {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  roleDialogOpen: boolean;
  setRoleDialogOpen: (open: boolean) => void;
  handleRoleClick: (role: Role) => void;
}

const AccessControlContext = createContext<AccessControlContextType | undefined>(undefined);

export function AccessControlProvider({ children }: { children: ReactNode }) {
  const [activeTab, setActiveTab] = useState("matrix");
  const [roleDialogOpen, setRoleDialogOpen] = useState(false);

  const handleRoleClick = useCallback((role: Role) => {
    // TODO: open role slide-over panel
    console.log("Role clicked:", role);
  }, []);

  return (
    <AccessControlContext.Provider
      value={{
        activeTab,
        setActiveTab,
        roleDialogOpen,
        setRoleDialogOpen,
        handleRoleClick,
      }}
    >
      {children}
    </AccessControlContext.Provider>
  );
}

export function useAccessControl() {
  const context = useContext(AccessControlContext);
  if (context === undefined) {
    throw new Error("useAccessControl must be used within an AccessControlProvider");
  }
  return context;
}
