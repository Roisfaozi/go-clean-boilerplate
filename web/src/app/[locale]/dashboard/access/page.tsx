"use client";

import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "~/components/ui/accordion";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Card, CardContent } from "~/components/ui/card";
import { Skeleton } from "~/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { accessApi, AccessRight } from "~/lib/api/access";
import { Role, rolesApi } from "~/lib/api/roles";

export default function AccessPage() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<string[][]>([]);
  const [accessRights, setAccessRights] = useState<AccessRight[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchAll = async () => {
    try {
      const [rolesResp, permsResp, accessResp] = await Promise.all([
        rolesApi.getAll(),
        accessApi.getAllPermissions(),
        accessApi.getAllAccessRights(),
      ]);

      if (rolesResp.data) setRoles(rolesResp.data);
      if (permsResp.data) setPermissions(permsResp.data);
      if (accessResp.data.data) setAccessRights(accessResp.data.data);
    } catch (error) {
      console.error("Failed to fetch access data", error);
      toast.error("Failed to load permissions");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchAll();
  }, []);

  // Indices for sub, dom, obj, act based on casbin_model.conf
  // p = sub, dom, obj, act
  const SUB_IDX = 0;
  // const DOM_IDX = 1;
  const OBJ_IDX = 2;
  const ACT_IDX = 3;

  const hasPermission = (role: string, path: string, method: string) => {
    return permissions.some(
      (p) => p[SUB_IDX] === role && p[OBJ_IDX] === path && p[ACT_IDX] === method
    );
  };

  const togglePermission = async (
    role: string,
    path: string,
    method: string
  ) => {
    const exists = hasPermission(role, path, method);
    try {
      if (exists) {
        await accessApi.revokePermission(role, path, method);
        setPermissions((prev) =>
          prev.filter(
            (p) =>
              !(
                p[SUB_IDX] === role &&
                p[OBJ_IDX] === path &&
                p[ACT_IDX] === method
              )
          )
        );
        toast.success(`Revoked ${method} on ${path} for ${role}`);
      } else {
        await accessApi.grantPermission(role, path, method);
        // Optimistically add to local state
        setPermissions((prev) => [...prev, [role, "global", path, method]]);
        toast.success(`Granted ${method} on ${path} to ${role}`);
      }
    } catch (error) {
      toast.error("Failed to update permission");
    }
  };

  // Find endpoints that are NOT in any access right but are in permissions
  const ungroupedEndpoints = useMemo(() => {
    const allEndpointsInRights = new Set(
      accessRights.flatMap((ar) =>
        ar.endpoints.map((e) => `${e.method}:${e.path}`)
      )
    );

    const ungrouped = new Map<string, { method: string; path: string }>();

    permissions.forEach((p) => {
      // Ignore g rules or invalid p rules
      if (p.length < 4) return;

      const path = p[OBJ_IDX];
      const method = p[ACT_IDX];

      if (!path || !method) return;

      const key = `${method}:${path}`;
      if (!allEndpointsInRights.has(key)) {
        ungrouped.set(key, { method, path });
      }
    });

    return Array.from(ungrouped.values());
  }, [accessRights, permissions]);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <Skeleton className="h-[400px] w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Access Control</h2>
          <p className="text-muted-foreground">
            Manage granular permissions by grouping endpoints into Access
            Rights.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Icon name="Plus" className="mr-2 h-4 w-4" />
            Add Access Right
          </Button>
          <Button size="sm">
            <Icon name="Plus" className="mr-2 h-4 w-4" />
            New Endpoint
          </Button>
        </div>
      </div>

      <Accordion
        type="multiple"
        defaultValue={accessRights.map((ar) => ar.id)}
        className="space-y-4"
      >
        {accessRights.map((group) => (
          <AccordionItem
            key={group.id}
            value={group.id}
            className="bg-card rounded-lg border px-4"
          >
            <AccordionTrigger className="py-4 hover:no-underline">
              <div className="flex items-center gap-3">
                <div className="bg-primary/10 rounded-md p-2">
                  <Icon name="Shield" className="text-primary h-5 w-5" />
                </div>
                <div className="text-foreground text-left">
                  <div className="mb-1 text-lg leading-none font-bold">
                    {group.name}
                  </div>
                  <div className="text-muted-foreground text-xs font-normal">
                    {group.endpoints?.length || 0} endpoints •{" "}
                    {group.description || "No description"}
                  </div>
                </div>
              </div>
            </AccordionTrigger>
            <AccordionContent className="pt-2 pb-6">
              <div className="bg-background overflow-hidden rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-muted/50">
                      <TableHead className="w-[300px]">Endpoint</TableHead>
                      {roles.map((role) => (
                        <TableHead
                          key={role.id}
                          className="text-center font-bold"
                        >
                          {role.name}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {group.endpoints?.map((ep) => (
                      <TableRow key={ep.id}>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Badge
                              variant="outline"
                              className="bg-muted/30 font-mono text-[10px] uppercase"
                            >
                              {ep.method}
                            </Badge>
                            <span
                              className="max-w-[200px] truncate font-mono text-xs"
                              title={ep.path}
                            >
                              {ep.path}
                            </span>
                          </div>
                        </TableCell>
                        {roles.map((role) => {
                          const active = hasPermission(
                            role.name,
                            ep.path,
                            ep.method
                          );
                          return (
                            <TableCell
                              key={`${role.id}-${ep.id}`}
                              className="text-center"
                            >
                              <div
                                onClick={() =>
                                  togglePermission(
                                    role.name,
                                    ep.path,
                                    ep.method
                                  )
                                }
                                className={`mx-auto flex h-6 w-6 cursor-pointer items-center justify-center rounded-md transition-all ${
                                  active
                                    ? "bg-primary text-primary-foreground shadow-sm hover:scale-110"
                                    : "bg-muted text-muted-foreground/30 hover:bg-muted/80 hover:text-muted-foreground"
                                } `}
                              >
                                {active ? (
                                  <Icon name="Check" className="h-4 w-4" />
                                ) : (
                                  <Icon name="Lock" className="h-3 w-3" />
                                )}
                              </div>
                            </TableCell>
                          );
                        })}
                      </TableRow>
                    ))}
                    {!group.endpoints ||
                      (group.endpoints.length === 0 && (
                        <TableRow>
                          <TableCell
                            colSpan={roles.length + 1}
                            className="text-muted-foreground h-24 text-center italic"
                          >
                            No endpoints linked to this Access Right.
                          </TableCell>
                        </TableRow>
                      ))}
                  </TableBody>
                </Table>
              </div>
            </AccordionContent>
          </AccordionItem>
        ))}

        {ungroupedEndpoints.length > 0 && (
          <AccordionItem
            value="ungrouped"
            className="bg-muted/20 rounded-lg border px-4"
          >
            <AccordionTrigger className="py-4 hover:no-underline">
              <div className="flex items-center gap-3">
                <div className="rounded-md bg-yellow-500/10 p-2">
                  <Icon
                    name="TriangleAlert"
                    className="h-5 w-5 text-yellow-600"
                  />
                </div>
                <div className="text-foreground text-left">
                  <div className="mb-1 text-lg leading-none font-bold text-yellow-700">
                    Ungrouped Policies
                  </div>
                  <div className="text-muted-foreground text-xs font-normal">
                    {ungroupedEndpoints.length} active Casbin rules not mapped
                    to any Access Right.
                  </div>
                </div>
              </div>
            </AccordionTrigger>
            <AccordionContent className="pt-2 pb-6">
              <div className="bg-background overflow-hidden rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-muted/50">
                      <TableHead className="w-[300px]">Endpoint</TableHead>
                      {roles.map((role) => (
                        <TableHead
                          key={role.id}
                          className="text-center font-bold"
                        >
                          {role.name}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {ungroupedEndpoints.map((ep, idx) => (
                      <TableRow key={`ungrouped-${idx}`}>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Badge
                              variant="outline"
                              className="font-mono text-[10px] uppercase"
                            >
                              {ep.method}
                            </Badge>
                            <span className="font-mono text-xs">{ep.path}</span>
                          </div>
                        </TableCell>
                        {roles.map((role) => {
                          const active = hasPermission(
                            role.name,
                            ep.path,
                            ep.method
                          );
                          return (
                            <TableCell
                              key={`${role.id}-ungrouped-${idx}`}
                              className="text-center"
                            >
                              <div
                                onClick={() =>
                                  togglePermission(
                                    role.name,
                                    ep.path,
                                    ep.method
                                  )
                                }
                                className={`mx-auto flex h-6 w-6 cursor-pointer items-center justify-center rounded-md transition-all ${
                                  active
                                    ? "bg-yellow-500 text-white shadow-sm hover:scale-110"
                                    : "bg-muted text-muted-foreground/30 hover:bg-muted/80"
                                } `}
                              >
                                {active ? (
                                  <Icon name="Check" className="h-4 w-4" />
                                ) : (
                                  <Icon name="Lock" className="h-3 w-3" />
                                )}
                              </div>
                            </TableCell>
                          );
                        })}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </AccordionContent>
          </AccordionItem>
        )}
      </Accordion>

      {accessRights.length === 0 && !isLoading && (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <div className="bg-muted mb-4 rounded-full p-4">
              <Icon
                name="ShieldOff"
                className="text-muted-foreground h-8 w-8"
              />
            </div>
            <h3 className="text-lg font-semibold">No Access Rights Defined</h3>
            <p className="text-muted-foreground mb-6 max-w-xs text-center text-sm">
              Start by creating an Access Right group and linking API endpoints
              to it.
            </p>
            <Button>
              <Icon name="Plus" className="mr-2 h-4 w-4" />
              Create your first Access Right
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
