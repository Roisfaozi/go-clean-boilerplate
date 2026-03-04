import { api } from "./client";

export interface Permission {
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

export interface ResourceCRUD {
  create: boolean;
  read: boolean;
  update: boolean;
  delete: boolean;
}

export interface ResourcePermission {
  name: string;
  base_path: string;
  role_permissions: Record<string, ResourceCRUD>;
  endpoint_count: number;
}

export interface ResourceAggregationResponse {
  resources: ResourcePermission[];
}

export interface RoleNode {
  id: string;
  name: string;
  description?: string;
  parent_id?: string | null;
  children?: RoleNode[];
  own_permissions: string[][];
  inherited_permissions: string[][];
  effective_permissions: string[][];
}

export interface InheritanceTreeResponse {
  roles: RoleNode[];
}

export interface AccessRightListResponse {
  data: {
    data: AccessRight[];
    meta: {
      total: number;
    };
  };
}

export interface RoleAccessRightStatus {
  id: string;
  name: string;
  endpoints: string[];
  is_assigned: boolean;
  is_partial: boolean;
}

export const accessApi = {
  getAllPermissions: () => {
    return api.get<{ data: string[][] }>("/permissions");
  },

  updatePermission: (oldPermission: string[], newPermission: string[]) => {
    return api.put("/permissions", {
      old_permission: oldPermission,
      new_permission: newPermission,
    });
  },

  assignRole: (userId: string, role: string, domain?: string) => {
    return api.post("/permissions/assign-role", {
      user_id: userId,
      role,
      domain,
    });
  },

  revokeRole: (userId: string, role: string, domain?: string) => {
    return api.post("/permissions/revoke-role", {
      user_id: userId,
      role,
      domain,
    });
  },

  grantPermission: (
    role: string,
    path: string,
    method: string,
    domain?: string
  ) => {
    return api.post("/permissions/grant", { role, path, method, domain });
  },

  revokePermission: (
    role: string,
    path: string,
    method: string,
    domain?: string
  ) => {
    return api.post("/permissions/revoke", { role, path, method, domain });
  },

  checkBatch: (items: { resource: string; action: string }[]) => {
    return api.post<{ data: { results: Record<string, boolean> } }>(
      "/permissions/check-batch",
      { items }
    );
  },

  getPermissionsForRole: (role: string) => {
    return api.get<{ data: string[][] }>(`/permissions/${role}`);
  },

  getUsersForRole: (role: string) => {
    return api.get<{ data: string[] }>(`/permissions/roles/${role}/users`);
  },

  getParentRoles: (role: string) => {
    return api.get<{ data: string[] }>(`/permissions/parents/${role}`);
  },

  addInheritance: (childRole: string, parentRole: string, domain?: string) => {
    return api.post("/permissions/inheritance", {
      child_role: childRole,
      parent_role: parentRole,
      domain,
    });
  },

  removeInheritance: (
    childRole: string,
    parentRole: string,
    domain?: string
  ) => {
    return api.delete("/permissions/inheritance", {
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        child_role: childRole,
        parent_role: parentRole,
        domain,
      }),
    } as any);
  },

  getResourceAggregation: () => {
    return api.get<{ data: ResourceAggregationResponse }>(
      "/permissions/resources"
    );
  },

  getInheritanceTree: () => {
    return api.get<{ data: InheritanceTreeResponse }>(
      "/permissions/inheritance-tree"
    );
  },

  getAllAccessRights: () => {
    return api.get<AccessRightListResponse>("/access-rights");
  },

  createAccessRight: (name: string, description: string) => {
    return api.post<{ data: AccessRight }>("/access-rights", {
      name,
      description,
    });
  },

  deleteAccessRight: (id: string) => {
    return api.delete(`/access-rights/${id}`);
  },

  linkEndpoint: (accessRightId: string, endpointId: string) => {
    return api.post("/access-rights/link", {
      access_right_id: accessRightId,
      endpoint_id: endpointId,
    });
  },
  unlinkEndpoint: (accessRightId: string, endpointId: string) => {
    return api.post("/access-rights/unlink", {
      access_right_id: accessRightId,
      endpoint_id: endpointId,
    });
  },

  createEndpoint: (method: string, path: string) => {
    return api.post<{ data: Endpoint }>("/endpoints", { method, path });
  },

  deleteEndpoint: (id: string) => {
    return api.delete(`/endpoints/${id}`);
  },

  searchEndpoints: (filter: any) => {
    return api.post<{ data: Endpoint[] }>("/endpoints/search", filter);
  },

  getRoleAccessRights: (role: string, domain = "global") => {
    return api.get<{ data: RoleAccessRightStatus[] }>(
      `/permissions/roles/${encodeURIComponent(role)}/access-rights?domain=${domain}`
    );
  },

  assignAccessRight: (
    role: string,
    accessRightId: string,
    domain = "global"
  ) => {
    return api.post("/permissions/assign-access-right", {
      role,
      access_right_id: accessRightId,
      domain,
    });
  },

  revokeAccessRight: (
    role: string,
    accessRightId: string,
    domain = "global"
  ) => {
    return api.delete("/permissions/revoke-access-right", {
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        role,
        access_right_id: accessRightId,
        domain,
      }),
    } as any);
  },
};
