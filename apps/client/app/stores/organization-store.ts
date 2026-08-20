import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { Organization, OrgMember } from "@/lib/api/schemas";

interface OrganizationState {
	activeOrganization: Organization | null;
	members: OrgMember[];
	setActiveOrganization: (org: Organization | null) => void;
	setMembers: (members: OrgMember[]) => void;
}

export const useOrganizationStore = create<OrganizationState>()(
	persist(
		(set) => ({
			activeOrganization: null,
			members: [],
			setActiveOrganization: (org) => set({ activeOrganization: org }),
			setMembers: (members) => set({ members }),
		}),
		{ name: "nexus-org" },
	),
);
