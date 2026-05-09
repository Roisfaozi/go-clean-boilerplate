"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import { Organization } from "~/lib/api/organizations";

interface OrganizationState {
  currentOrganization: Organization | null;
  setCurrentOrganization: (org: Organization | null) => void;
  clearOrganization: () => void;
}

export const useOrganizationStore = create<OrganizationState>()(
  persist(
    (set) => ({
      currentOrganization: null,
      setCurrentOrganization: (org) => set({ currentOrganization: org }),
      clearOrganization: () => set({ currentOrganization: null }),
    }),
    {
      name: "nexus-organization-storage",
    }
  )
);
