import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { PageHeader } from "@/components/layout/page-header";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@casbin/ui";
import { PermissionMatrix } from "@/features/permissions/permission-matrix";
import { RoleInheritanceTree } from "@/features/permissions/role-inheritance-tree";
import { AccessRightManager } from "@/features/permissions/access-right-manager";
import { permissionsApi } from "@/lib/api/permissions";
import { rolesApi } from "@/lib/api/roles";

export default function RolesPermissionsPage() {
	const [activeTab, setActiveTab] = useState("matrix");
	const queryClient = useQueryClient();

	const { data: rolesRes } = useQuery({
		queryKey: ["roles"],
		queryFn: () => rolesApi.list({ limit: 100 }),
	});

	const { data: matrixRes } = useQuery({
		queryKey: ["permissions-matrix"],
		queryFn: () => permissionsApi.getResourceAggregation(),
	});

	const { data: inheritanceRes } = useQuery({
		queryKey: ["inheritance-tree"],
		queryFn: () => permissionsApi.getInheritanceTree(),
	});

	const grantMutation = useMutation({
		mutationFn: permissionsApi.grant,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["permissions-matrix"] });
			queryClient.invalidateQueries({ queryKey: ["inheritance-tree"] });
		},
	});

	const revokeMutation = useMutation({
		mutationFn: permissionsApi.revoke,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["permissions-matrix"] });
			queryClient.invalidateQueries({ queryKey: ["inheritance-tree"] });
		},
	});

	const roles = rolesRes?.data?.map((r) => r.name) || [];
	const resources = matrixRes?.resources || [];
	const actions = ["create", "read", "update", "delete"];

	// Map backend aggregation to frontend matrix format
	const initialPermissions: Record<string, Record<string, string[]>> = {};
	resources.forEach((res) => {
		Object.entries(res.role_permissions).forEach(([roleName, crud]) => {
			if (!initialPermissions[roleName]) initialPermissions[roleName] = {};
			const grantedActions: string[] = [];
			if (crud.create) grantedActions.push("create");
			if (crud.read) grantedActions.push("read");
			if (crud.update) grantedActions.push("update");
			if (crud.delete) grantedActions.push("delete");
			initialPermissions[roleName][res.name] = grantedActions;
		});
	});

	const handlePermissionChange = async (
		role: string,
		resourceName: string,
		action: string,
		granted: boolean,
	) => {
		const resource = resources.find((r) => r.name === resourceName);
		if (!resource) return;

		const payload = {
			role,
			path: resource.base_path,
			method:
				action.toUpperCase() === "READ"
					? "GET"
					: action.toUpperCase() === "CREATE"
						? "POST"
						: action.toUpperCase() === "UPDATE"
							? "PUT"
							: "DELETE",
		};

		if (granted) {
			await grantMutation.mutateAsync(payload);
		} else {
			await revokeMutation.mutateAsync(payload);
		}
	};

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
						resources={resources.map((r) => r.name)}
						actions={actions}
						initialPermissions={initialPermissions}
						onPermissionChange={handlePermissionChange}
					/>
				</TabsContent>

				<TabsContent value="inheritance" className="mt-4">
					<div className="max-w-lg">
						<RoleInheritanceTree tree={inheritanceRes?.roles || []} />
					</div>
				</TabsContent>

				<TabsContent value="access-rights" className="mt-4">
					<AccessRightManager />
				</TabsContent>
			</Tabs>
		</div>
	);
}
