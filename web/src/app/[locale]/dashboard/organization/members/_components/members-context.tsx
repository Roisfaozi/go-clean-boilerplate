"use client";

import {
  createContext,
  useContext,
  useState,
  useCallback,
  ReactNode,
  useEffect,
} from "react";
import { Member, organizationsApi } from "~/lib/api/organizations";
import { Role, rolesApi } from "~/lib/api/roles";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";

interface MembersContextType {
  members: Member[];
  roles: Role[];
  isLoading: boolean;
  isInviting: boolean;
  fetchData: () => Promise<void>;
  inviteMember: (email: string, roleId: string) => Promise<void>;
  updateMemberRole: (userId: string, roleId: string) => Promise<void>;
  removeMember: (userId: string, name: string) => Promise<void>;
}

const MembersContext = createContext<MembersContextType | undefined>(undefined);

export function MembersProvider({ children }: { children: ReactNode }) {
  const { currentOrganization } = useOrganizationStore();
  const [members, setMembers] = useState<Member[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isInviting, setIsInviting] = useState(false);

  const fetchData = useCallback(async () => {
    if (!currentOrganization) return;
    setIsLoading(true);
    try {
      const [membersResp, rolesResp] = await Promise.all([
        organizationsApi.getMembers(currentOrganization.id),
        rolesApi.getAll(),
      ]);
      if (membersResp.data) setMembers(membersResp.data);
      if (rolesResp.data) setRoles(rolesResp.data);
    } catch (error) {
      console.error("Failed to fetch data", error);
      toast.error("Failed to load members");
    } finally {
      setIsLoading(false);
    }
  }, [currentOrganization]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const inviteMember = useCallback(
    async (email: string, roleId: string) => {
      if (!currentOrganization) return;
      setIsInviting(true);
      try {
        await organizationsApi.inviteMember(currentOrganization.id, {
          email,
          role_id: roleId,
        });
        toast.success("Invitation sent successfully");
        await fetchData();
      } catch (error: any) {
        toast.error(error.message || "Failed to send invitation");
        throw error;
      } finally {
        setIsInviting(false);
      }
    },
    [currentOrganization, fetchData]
  );

  const updateMemberRole = useCallback(
    async (userId: string, roleId: string) => {
      if (!currentOrganization) return;
      try {
        await organizationsApi.updateMemberRole(
          currentOrganization.id,
          userId,
          {
            role_id: roleId,
          }
        );
        toast.success("Member role updated");
        await fetchData();
      } catch (error) {
        toast.error("Failed to update role");
      }
    },
    [currentOrganization, fetchData]
  );

  const removeMember = useCallback(
    async (userId: string, name: string) => {
      if (!currentOrganization) return;
      if (
        !confirm(
          `Are you sure you want to remove ${name} from the organization?`
        )
      )
        return;

      try {
        await organizationsApi.removeMember(currentOrganization.id, userId);
        toast.success("Member removed successfully");
        await fetchData();
      } catch (error) {
        toast.error("Failed to remove member");
      }
    },
    [currentOrganization, fetchData]
  );

  return (
    <MembersContext.Provider
      value={{
        members,
        roles,
        isLoading,
        isInviting,
        fetchData,
        inviteMember,
        updateMemberRole,
        removeMember,
      }}
    >
      {children}
    </MembersContext.Provider>
  );
}

export function useMembers() {
  const context = useContext(MembersContext);
  if (context === undefined) {
    throw new Error("useMembers must be used within a MembersProvider");
  }
  return context;
}
