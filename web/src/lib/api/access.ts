import { api } from "./client";

export interface Permission {
  // Structure depends on backend response, usually array of strings
  // [sub, obj, act]
  [key: string]: any;
}

export interface AccessRight {
  id: string;
  name: string;
  description: string;
  endpoints: Endpoint[];
  created_at: number;
  updated_at: number;
}

export interface Endpoint {
  id: string;
  method: string;
  path: string;
  created_at: number;
}

export interface AccessRightListResponse {
  data: AccessRight[];
  meta: {
    total: number;
  };
}

export const accessApi = {
  // --- Permissions (Casbin) ---
  
  // Get all permissions
  getAllPermissions: () => {
    return api.get<{ data: string[][] }>("/permissions");
  },

  // Update permission (policy)
  updatePermission: (oldPermission: string[], newPermission: string[]) => {
    return api.put("/permissions", { old_permission: oldPermission, new_permission: newPermission });
  },

  // Assign role to user
  assignRole: (userId: string, role: string) => {
    return api.post("/permissions/assign-role", { user_id: userId, role });
  },

  // Revoke role from user
  revokeRole: (userId: string, role: string) => {
    return api.post("/permissions/revoke-role", { user_id: userId, role });
  },

  // Grant permission to role
  grantPermission: (role: string, path: string, method: string) => {
    return api.post("/permissions/grant", { role, path, method });
  },

  // Revoke permission from role
  revokePermission: (role: string, path: string, method: string) => {
    return api.post("/permissions/revoke", { role, path, method });
  },

  // Check batch permissions
  checkBatch: (items: { resource: string; action: string }[]) => {
    return api.post<{ data: { results: Record<string, boolean> } }>("/permissions/check-batch", { items });
  },

  // Get permissions for role
  getPermissionsForRole: (role: string) => {
    return api.get<{ data: string[][] }>(`/permissions/roles/${role}`);
  },

  // Get users for role
  getUsersForRole: (role: string) => {
    return api.get<{ data: string[] }>(`/permissions/roles/${role}/users`);
  },

  // Get parent roles
  getParentRoles: (role: string) => {
    return api.get<{ data: string[] }>(`/permissions/parents/${role}`);
  },

  // Add inheritance
  addInheritance: (childRole: string, parentRole: string) => {
    return api.post("/permissions/inheritance", { child_role: childRole, parent_role: parentRole });
  },

  // Remove inheritance
  removeInheritance: (childRole: string, parentRole: string) => {
    return api.delete("/permissions/inheritance", { 
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ child_role: childRole, parent_role: parentRole }) 
    } as any);
  },

  // --- Access Rights (Resource Groups) ---

  getAllAccessRights: () => {
    return api.get<AccessRightListResponse>("/access-rights");
  },

  createAccessRight: (name: string, description: string) => {
    return api.post<{ data: AccessRight }>("/access-rights", { name, description });
  },

  deleteAccessRight: (id: string) => {
    return api.delete(`/access-rights/${id}`);
  },

  linkEndpoint: (accessRightId: string, endpointId: string) => {
    return api.post("/access-rights/link", { access_right_id: accessRightId, endpoint_id: endpointId });
  },

  // --- Endpoints ---

  createEndpoint: (method: string, path: string) => {
    return api.post<{ data: Endpoint }>("/endpoints", { method, path });
  },

  deleteEndpoint: (id: string) => {
    return api.delete(`/endpoints/${id}`);
  },
  
  searchEndpoints: (filter: any) => {
      return api.post<{ data: Endpoint[] }>("/endpoints/search", filter);
  }
};
