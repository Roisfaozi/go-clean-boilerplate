import { useState, useMemo } from "react";
import { z } from "zod";
import { PageHeader } from "@/components/layout/page-header";
import {
  NexusButton,
  NexusBadge,
  NexusCard,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Switch,
  Skeleton,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@casbin/ui";
import {
  CrudTable,
  CrudFormDialog,
  DeleteDialog,
  type CrudColumnDef,
  type FieldDef,
} from "@/features/shared";
import { Shield, Box, RefreshCw, GitBranch, Key } from "lucide-react";
import {
  usePermissions,
  useCreatePermission,
  useUpdatePermission,
  useDeletePermission,
  useResourceAggregation,
  useRoleAccessRights,
  useToggleAccessRight,
  useInheritanceTree,
  useAddInheritance,
  useRemoveInheritance,
} from "./permissionHooks";
import { useRoles } from "../roles/roleHooks";
import { RoleInheritanceTree } from "./role-inheritance-tree";
import type { Permission, Role, Resource } from "@/lib/api/types";

// --- Matrix Components ---

function MatrixCell({
  role,
  resourceId,
  domain,
}: {
  role: string;
  resourceId: string;
  domain?: string;
}) {
  const { data: roleData, isLoading } = useRoleAccessRights(role, domain);
  const toggle = useToggleAccessRight();

  const status = useMemo(() => {
    if (!roleData?.data) return { assigned: false, partial: false };
    const item = roleData.data.find((r: any) => r.id === resourceId);
    return { assigned: item?.is_assigned ?? false, partial: item?.is_partial ?? false };
  }, [roleData, resourceId]);

  if (isLoading) return <Skeleton className="mx-auto h-5 w-10" />;

  return (
    <div className="flex flex-col items-center gap-1">
      <Switch
        checked={status.assigned}
        onCheckedChange={(checked) =>
          toggle.mutate({ role, access_right_id: resourceId, granted: checked, domain })
        }
        disabled={toggle.isPending}
      />
      {status.partial && !status.assigned && (
        <span className="text-[10px] font-medium text-amber-500">Partial</span>
      )}
    </div>
  );
}

function PermissionMatrix({ roles, resources }: { roles: Role[]; resources: any[] }) {
  return (
    <NexusCard className="shadow-premium overflow-hidden border-none bg-white/50 backdrop-blur-sm">
      <div className="overflow-x-auto">
        <Table>
          <TableHeader className="bg-muted/50">
            <TableRow>
              <TableHead className="text-foreground w-[250px] py-6 pl-8 font-bold">
                <div className="flex items-center gap-2">
                  <Box className="text-primary h-4 w-4" />
                  Resources / Modules
                </div>
              </TableHead>
              {roles.map((role) => (
                <TableHead
                  key={role.id}
                  className="text-foreground min-w-[120px] text-center font-bold"
                >
                  <div className="flex flex-col items-center gap-1">
                    <Shield className="text-primary/70 h-4 w-4" />
                    {role.name}
                  </div>
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {resources.map((res) => (
              <TableRow key={res.id} className="hover:bg-primary/5 group transition-colors">
                <TableCell className="py-4 pl-8 font-medium">
                  <div className="flex flex-col">
                    <span className="text-foreground group-hover:text-primary transition-colors">
                      {res.name}
                    </span>
                    <span className="text-muted-foreground text-xs font-normal">
                      {res.endpoint_count} Endpoints
                    </span>
                  </div>
                </TableCell>
                {roles.map((role) => (
                  <TableCell key={`${res.id}-${role.id}`} className="text-center">
                    <MatrixCell role={role.name} resourceId={res.id} />
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </NexusCard>
  );
}

// --- Main Page ---

const columns: CrudColumnDef<Permission>[] = [
  {
    id: "role_name",
    header: "Role",
    accessorKey: "role_name",
    sortable: true,
    filterable: true,
  },
  {
    id: "access_right_name",
    header: "Access Right",
    accessorKey: "access_right_name",
    sortable: true,
  },
  {
    id: "granted",
    header: "Status",
    accessorKey: "granted",
    cell: (row) => (
      <NexusBadge variant={row.granted ? "success" : "danger"}>
        {row.granted ? "Granted" : "Denied"}
      </NexusBadge>
    ),
  },
];

export default function PermissionsPage() {
  const [activeTab, setActiveTab] = useState("matrix");
  const [createOpen, setCreateOpen] = useState(false);
  const [editItem, setEditItem] = useState<Permission | null>(null);
  const [deleteItem, setDeleteItem] = useState<Permission | null>(null);

  // Queries
  const {
    data: response,
    isLoading: permissionsLoading,
    refetch: refetchPermissions,
  } = usePermissions();
  const { data: rolesResponse, isLoading: rolesLoading } = useRoles();
  const { data: resourcesResponse, isLoading: resourcesLoading } = useResourceAggregation();
  const { data: inheritanceResponse, isLoading: inheritanceLoading } = useInheritanceTree();

  // Mutations
  const createPermission = useCreatePermission();
  const updatePermission = useUpdatePermission();
  const deletePermission = useDeletePermission();

  const permissions: Permission[] = useMemo(() => {
    if (response?.data) return response.data as Permission[];
    return [];
  }, [response]);

  const roles = useMemo(() => (rolesResponse?.data || []) as Role[], [rolesResponse]);
  const resources = useMemo(
    () => (resourcesResponse?.data?.resources || []) as any[],
    [resourcesResponse],
  );
  const inheritanceTree = useMemo(
    () => (inheritanceResponse?.data?.roles || []) as any[],
    [inheritanceResponse],
  );

  const isLoading = permissionsLoading || rolesLoading || resourcesLoading || inheritanceLoading;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Permissions Management"
        description="Configure access rights and security policies for your roles."
        actions={
          <div className="flex gap-2">
            <NexusButton variant="outline" size="sm" onClick={() => refetchPermissions()}>
              <RefreshCw className="mr-2 h-4 w-4" /> Refresh
            </NexusButton>
            <NexusButton size="sm" onClick={() => setCreateOpen(true)}>
              <Shield className="mr-2 h-4 w-4" /> Add Mapping
            </NexusButton>
          </div>
        }
      />

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <div className="mb-4 flex items-center justify-between">
          <TabsList className="bg-muted/50 p-1">
            <TabsTrigger
              value="matrix"
              className="data-[state=active]:bg-white data-[state=active]:shadow-sm"
            >
              <Key className="text-primary mr-2 h-4 w-4" /> Matrix View
            </TabsTrigger>
            <TabsTrigger
              value="inheritance"
              className="data-[state=active]:bg-white data-[state=active]:shadow-sm"
            >
              <GitBranch className="text-primary mr-2 h-4 w-4" /> Role Inheritance
            </TabsTrigger>
            <TabsTrigger
              value="list"
              className="data-[state=active]:bg-white data-[state=active]:shadow-sm"
            >
              <Shield className="text-primary mr-2 h-4 w-4" /> List View
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="matrix" className="mt-0 outline-none focus-visible:ring-0">
          {isLoading ? (
            <div className="grid grid-cols-1 gap-4">
              <Skeleton className="h-[400px] w-full rounded-xl" />
            </div>
          ) : (
            <PermissionMatrix roles={roles} resources={resources} />
          )}
        </TabsContent>

        <TabsContent value="inheritance" className="mt-0 outline-none focus-visible:ring-0">
          <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
            <div className="md:col-span-2">
              {isLoading ? (
                <Skeleton className="h-[400px] w-full rounded-xl" />
              ) : (
                <RoleInheritanceTree tree={inheritanceTree} />
              )}
            </div>
            <div className="space-y-4">
              <NexusCard className="bg-primary/5 border-primary/10 p-6">
                <h4 className="text-primary mb-2 flex items-center gap-2 font-bold">
                  <Shield className="h-4 w-4" /> About Inheritance
                </h4>
                <p className="text-muted-foreground text-xs leading-relaxed">
                  Roles can inherit permissions from other roles. For example, an <b>Editor</b> role
                  might inherit from <b>Viewer</b>, gaining all its permissions automatically.
                  <br />
                  <br />
                  Effective permissions shown in the tree include both direct and inherited rights.
                </p>
              </NexusCard>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="list" className="mt-0 outline-none focus-visible:ring-0">
          <CrudTable
            columns={columns}
            data={permissions}
            loading={isLoading}
            onEdit={setEditItem}
            onDelete={setDeleteItem}
          />
        </TabsContent>
      </Tabs>

      {/* CRUD Dialogs (keep for manual specific overrides if needed) */}
      <CrudFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        title="Manual Permission Mapping"
        description="Directly assign a specific access right to a role."
        fields={[
          {
            name: "role_id",
            label: "Role",
            type: "select",
            required: true,
            options: roles.map((r) => ({ label: r.name, value: r.id })),
          },
          {
            name: "access_right_id",
            label: "Access Right",
            type: "select",
            required: true,
            options: resources.map((r) => ({ label: r.name, value: r.id })),
          },
          { name: "granted", label: "Granted", type: "switch" },
        ]}
        schema={z.object({
          role_id: z.string().min(1),
          access_right_id: z.string().min(1),
          granted: z.boolean(),
        })}
        loading={createPermission.isPending}
        onSubmit={async (v) => {
          await createPermission.mutateAsync(v as any);
          setCreateOpen(false);
        }}
        submitLabel="Assign"
      />

      <DeleteDialog
        open={!!deleteItem}
        onOpenChange={(o) => !o && setDeleteItem(null)}
        resourceName="Permission Mapping"
        itemName={`${deleteItem?.role_name} → ${deleteItem?.access_right_name}`}
        loading={deletePermission.isPending}
        onConfirm={async () => {
          if (deleteItem) {
            await deletePermission.mutateAsync(String(deleteItem.id));
            setDeleteItem(null);
          }
        }}
      />
    </div>
  );
}
