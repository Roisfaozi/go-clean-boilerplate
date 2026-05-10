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
import { Plus, LayoutGrid, List, Shield, Box, RefreshCw } from "lucide-react";
import {
  usePermissions,
  useCreatePermission,
  useUpdatePermission,
  useDeletePermission,
  useResourceAggregation,
  useRoleAccessRights,
  useToggleAccessRight,
} from "./permissionHooks";
import { useRoles } from "../roles/roleHooks";
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

  if (isLoading) return <Skeleton className="h-5 w-10 mx-auto" />;

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
        <span className="text-[10px] text-amber-500 font-medium">Partial</span>
      )}
    </div>
  );
}

function PermissionMatrix({ roles, resources }: { roles: Role[]; resources: any[] }) {
  return (
    <NexusCard className="overflow-hidden border-none shadow-premium bg-white/50 backdrop-blur-sm">
      <div className="overflow-x-auto">
        <Table>
          <TableHeader className="bg-muted/50">
            <TableRow>
              <TableHead className="w-[250px] font-bold text-foreground py-6 pl-8">
                <div className="flex items-center gap-2">
                  <Box className="h-4 w-4 text-primary" />
                  Resources / Modules
                </div>
              </TableHead>
              {roles.map((role) => (
                <TableHead key={role.id} className="text-center font-bold text-foreground min-w-[120px]">
                  <div className="flex flex-col items-center gap-1">
                    <Shield className="h-4 w-4 text-primary/70" />
                    {role.name}
                  </div>
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {resources.map((res) => (
              <TableRow key={res.id} className="hover:bg-primary/5 transition-colors group">
                <TableCell className="font-medium pl-8 py-4">
                  <div className="flex flex-col">
                    <span className="text-foreground group-hover:text-primary transition-colors">
                      {res.name}
                    </span>
                    <span className="text-xs text-muted-foreground font-normal">
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
  const { data: response, isLoading: permissionsLoading, refetch: refetchPermissions } = usePermissions();
  const { data: rolesResponse, isLoading: rolesLoading } = useRoles();
  const { data: resourcesResponse, isLoading: resourcesLoading } = useResourceAggregation();

  // Mutations
  const createPermission = useCreatePermission();
  const updatePermission = useUpdatePermission();
  const deletePermission = useDeletePermission();

  const permissions: Permission[] = useMemo(() => {
    if (response?.data) return response.data as Permission[];
    return [];
  }, [response]);

  const roles = useMemo(() => (rolesResponse?.data || []) as Role[], [rolesResponse]);
  const resources = useMemo(() => (resourcesResponse?.data?.resources || []) as any[], [resourcesResponse]);

  const isLoading = permissionsLoading || rolesLoading || resourcesLoading;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Permissions Management"
        description="Configure access rights and security policies for your roles."
        actions={
          <div className="flex gap-2">
            <NexusButton variant="outline" size="sm" onClick={() => refetchPermissions()}>
              <RefreshCw className="h-4 w-4 mr-2" /> Refresh
            </NexusButton>
            <NexusButton size="sm" onClick={() => setCreateOpen(true)}>
              <Plus className="h-4 w-4 mr-2" /> Add Mapping
            </NexusButton>
          </div>
        }
      />

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <div className="flex items-center justify-between mb-4">
          <TabsList className="bg-muted/50 p-1">
            <TabsTrigger value="matrix" className="data-[state=active]:bg-white data-[state=active]:shadow-sm">
              <LayoutGrid className="h-4 w-4 mr-2" /> Matrix View
            </TabsTrigger>
            <TabsTrigger value="list" className="data-[state=active]:bg-white data-[state=active]:shadow-sm">
              <List className="h-4 w-4 mr-2" /> List View
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="matrix" className="mt-0">
          {isLoading ? (
            <div className="grid grid-cols-1 gap-4">
              <Skeleton className="h-[400px] w-full rounded-xl" />
            </div>
          ) : (
            <PermissionMatrix roles={roles} resources={resources} />
          )}
        </TabsContent>

        <TabsContent value="list" className="mt-0">
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
            options: roles.map(r => ({ label: r.name, value: r.id })),
          },
          {
            name: "access_right_id",
            label: "Access Right",
            type: "select",
            required: true,
            options: resources.map(r => ({ label: r.name, value: r.id })),
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
