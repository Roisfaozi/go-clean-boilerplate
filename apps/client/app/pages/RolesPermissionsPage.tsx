import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import {  Tabs, TabsContent, TabsList, TabsTrigger  } from "@casbin/ui";
import { PermissionMatrix } from "@/features/permissions/permission-matrix";
import { RoleInheritanceTree, RoleNode } from "@/features/permissions/role-inheritance-tree";
import { AccessRightManager } from "@/features/permissions/access-right-manager";

const roles = ["Super Admin", "Admin", "Editor", "Viewer"];
const resources = ["Users", "Roles", "Organizations", "Projects", "Permissions", "Audit Logs", "Settings"];
const actions = ["Create", "Read", "Update", "Delete"];

const initialPermissions: Record<string, Record<string, string[]>> = {
  "Super Admin": {
    Users: ["Create", "Read", "Update", "Delete"],
    Roles: ["Create", "Read", "Update", "Delete"],
    Organizations: ["Create", "Read", "Update", "Delete"],
    Projects: ["Create", "Read", "Update", "Delete"],
    Permissions: ["Create", "Read", "Update", "Delete"],
    "Audit Logs": ["Read"],
    Settings: ["Create", "Read", "Update", "Delete"],
  },
  Admin: {
    Users: ["Create", "Read", "Update"],
    Roles: ["Read"],
    Organizations: ["Read", "Update"],
    Projects: ["Create", "Read", "Update", "Delete"],
    Permissions: ["Read"],
    "Audit Logs": ["Read"],
    Settings: ["Read", "Update"],
  },
  Editor: {
    Users: ["Read"],
    Projects: ["Create", "Read", "Update"],
    Organizations: ["Read"],
  },
  Viewer: {
    Users: ["Read"],
    Projects: ["Read"],
    Organizations: ["Read"],
  },
};

const roleTree: RoleNode[] = [
  {
    id: "super-admin",
    name: "Super Admin",
    permissions: ["all"],
    children: [
      {
        id: "admin",
        name: "Admin",
        permissions: ["users:crud", "projects:crud", "settings:rw"],
        children: [
          {
            id: "editor",
            name: "Editor",
            permissions: ["projects:crud", "users:read"],
            children: [
              {
                id: "viewer",
                name: "Viewer",
                permissions: ["read-only"],
              },
            ],
          },
        ],
      },
    ],
  },
];

export default function RolesPermissionsPage() {
  const [activeTab, setActiveTab] = useState("matrix");

  return (
    <div className="space-y-6">
      <PageHeader
        title="Permission Management"
        description="Assign permissions to roles, view inheritance, and manage access rights"
      />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="matrix">Permission Matrix</TabsTrigger>
          <TabsTrigger value="inheritance">Role Inheritance</TabsTrigger>
          <TabsTrigger value="access-rights">Access Rights</TabsTrigger>
        </TabsList>

        <TabsContent value="matrix" className="mt-4">
          <PermissionMatrix
            roles={roles}
            resources={resources}
            actions={actions}
            initialPermissions={initialPermissions}
          />
        </TabsContent>

        <TabsContent value="inheritance" className="mt-4">
          <div className="max-w-lg">
            <RoleInheritanceTree tree={roleTree} />
          </div>
        </TabsContent>

        <TabsContent value="access-rights" className="mt-4">
          <AccessRightManager />
        </TabsContent>
      </Tabs>
    </div>
  );
}
