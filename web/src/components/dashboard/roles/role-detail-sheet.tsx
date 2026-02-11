"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import { ScrollArea } from "~/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "~/components/ui/sheet";
import { Skeleton } from "~/components/ui/skeleton";
import { Switch } from "~/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { accessApi, AccessRight, Endpoint } from "~/lib/api/access";
import { Role } from "~/lib/api/roles";
import { User, usersApi } from "~/lib/api/users";

interface RoleDetailSheetProps {
  role?: Role;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RoleDetailSheet({
  role,
  open,
  onOpenChange,
}: RoleDetailSheetProps) {
  const [members, setMembers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isAdding, setIsAdding] = useState(false);
  const [isProcessing, setIsProcessing] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const fetchMembers = useCallback(async () => {
    if (!role) return;
    setIsLoading(true);
    try {
      const resp = await accessApi.getUsersForRole(role.name);
      const userIds = resp.data || [];

      if (userIds.length === 0) {
        setMembers([]);
        return;
      }

      const userPromises = userIds
        .slice(0, 50)
        .map((id) => usersApi.getById(id));
      const userResps = await Promise.all(userPromises);
      setMembers(userResps.map((r) => r.data).filter(Boolean) as User[]);
    } catch (error) {
      console.error("Failed to fetch role members", error);
      toast.error("Failed to load members");
    } finally {
      setIsLoading(false);
    }
  }, [role]);

  useEffect(() => {
    if (open && role) {
      fetchMembers();
    }
  }, [open, role, fetchMembers]);

  const handleSearch = async (query: string) => {
    setSearchQuery(query);
    if (query.length < 2) {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    try {
      const resp = await usersApi.getAll(1, 10, query);
      if (resp.data) {
        const memberIds = new Set(members.map((m) => m.id));
        setSearchResults(resp.data.filter((u) => !memberIds.has(u.id)));
      }
    } catch (error) {
      console.error("Search failed", error);
    } finally {
      setIsSearching(false);
    }
  };

  const addMember = async (user: User) => {
    if (!role) return;
    setIsAdding(true);
    try {
      await accessApi.assignRole(user.id, role.name);
      toast.success(`${user.username} added to ${role.name}`);
      setMembers((prev) => [...prev, user]);
      setSearchQuery("");
      setSearchResults([]);
    } catch (error) {
      toast.error("Failed to add member");
    } finally {
      setIsAdding(false);
    }
  };

  const removeMember = async (userId: string, username: string) => {
    if (!role) return;
    setIsProcessing(userId);
    try {
      await accessApi.revokeRole(userId, role.name);
      toast.success(`${username} removed from ${role.name}`);
      setMembers((prev) => prev.filter((m) => m.id !== userId));
    } catch (error) {
      toast.error("Failed to remove member");
    } finally {
      setIsProcessing(null);
    }
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex h-full flex-col sm:max-w-md">
        <SheetHeader>
          <div className="mb-2 flex items-center gap-2">
            <div className="bg-primary/10 text-primary rounded-md p-2">
              <Icon name="Shield" className="h-5 w-5" />
            </div>
            <SheetTitle className="text-xl">{role?.name}</SheetTitle>
          </div>
          <SheetDescription>
            {role?.description ||
              "Manage members and permissions for this role."}
          </SheetDescription>
        </SheetHeader>

        <Tabs
          defaultValue="members"
          className="mt-6 flex flex-1 flex-col overflow-hidden"
        >
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="members">Members</TabsTrigger>
            <TabsTrigger value="permissions">Permissions</TabsTrigger>
          </TabsList>

          <TabsContent
            value="members"
            className="mt-4 flex min-h-0 flex-1 flex-col"
          >
            <div className="mb-4 flex items-center justify-between">
              <h3 className="flex items-center gap-2 text-sm font-semibold">
                <Icon name="Users" className="text-muted-foreground h-4 w-4" />
                Members ({members.length})
              </h3>

              <Popover>
                <PopoverTrigger asChild>
                  <Button size="sm" variant="outline" className="h-8">
                    <Icon name="UserPlus" className="mr-2 h-4 w-4" />
                    Add Member
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-[300px] p-0" align="end">
                  <Command shouldFilter={false}>
                    <CommandInput
                      placeholder="Search users..."
                      value={searchQuery}
                      onValueChange={handleSearch}
                    />
                    <CommandList>
                      {isSearching && (
                        <div className="p-4 text-center">
                          <Icon
                            name="Loader"
                            className="mx-auto mb-2 h-4 w-4 animate-spin"
                          />
                          <span className="text-muted-foreground text-xs">
                            Searching...
                          </span>
                        </div>
                      )}
                      {!isSearching &&
                        searchResults.length === 0 &&
                        searchQuery.length >= 2 && (
                          <CommandEmpty>No users found.</CommandEmpty>
                        )}
                      {!isSearching && searchQuery.length < 2 && (
                        <div className="text-muted-foreground p-4 text-center text-xs">
                          Type at least 2 characters to search...
                        </div>
                      )}
                      <CommandGroup>
                        {searchResults.map((user) => (
                          <CommandItem
                            key={user.id}
                            onSelect={() => addMember(user)}
                            className="flex cursor-pointer items-center gap-2"
                          >
                            <Avatar className="h-6 w-6">
                              <AvatarImage src={user.avatar_url} />
                              <AvatarFallback>
                                {user.username[0].toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <div className="flex flex-col">
                              <span className="text-sm font-medium">
                                {user.username}
                              </span>
                              <span className="text-muted-foreground text-[10px]">
                                {user.email}
                              </span>
                            </div>
                            <Icon
                              name="Plus"
                              className="text-muted-foreground ml-auto h-3 w-3"
                            />
                          </CommandItem>
                        ))}
                      </CommandGroup>
                    </CommandList>
                  </Command>
                </PopoverContent>
              </Popover>
            </div>

            <ScrollArea className="-mx-2 flex-1 px-2">
              <div className="space-y-2">
                {isLoading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <div key={i} className="flex items-center gap-3 p-2">
                      <Skeleton className="h-9 w-9 rounded-full" />
                      <div className="flex-1 space-y-1">
                        <Skeleton className="h-4 w-24" />
                        <Skeleton className="h-3 w-32" />
                      </div>
                    </div>
                  ))
                ) : members.length === 0 ? (
                  <div className="rounded-lg border-2 border-dashed py-12 text-center">
                    <Icon
                      name="Users"
                      className="text-muted-foreground/30 mx-auto mb-2 h-8 w-8"
                    />
                    <p className="text-muted-foreground text-sm">
                      No members assigned yet.
                    </p>
                  </div>
                ) : (
                  members.map((member) => (
                    <div
                      key={member.id}
                      className="hover:bg-muted/50 group flex items-center gap-3 rounded-md p-2 transition-colors"
                    >
                      <Avatar className="h-9 w-9 border">
                        <AvatarImage src={member.avatar_url} />
                        <AvatarFallback>
                          {member.username[0].toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <div className="min-w-0 flex-1">
                        <div className="mb-1 text-sm leading-none font-medium">
                          {member.username}
                        </div>
                        <div className="text-muted-foreground truncate text-xs">
                          {member.email}
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="text-destructive hover:bg-destructive/10 h-8 w-8 opacity-0 group-hover:opacity-100"
                        disabled={isProcessing === member.id}
                        onClick={() => removeMember(member.id, member.username)}
                      >
                        {isProcessing === member.id ? (
                          <Icon
                            name="Loader"
                            className="h-4 w-4 animate-spin"
                          />
                        ) : (
                          <Icon name="UserMinus" className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  ))
                )}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent
            value="permissions"
            className="mt-4 flex min-h-0 flex-1 flex-col"
          >
            <RolePermissionsTab role={role} />
          </TabsContent>
        </Tabs>

        <div className="mt-auto border-t pt-6">
          <Button
            variant="outline"
            className="w-full"
            onClick={() => onOpenChange(false)}
          >
            Close
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}

function RolePermissionsTab({ role }: { role?: Role }) {
  const [accessRights, setAccessRights] = useState<AccessRight[]>([]);
  const [rolePerms, setRolePerms] = useState<string[][]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isProcessing, setIsProcessing] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    if (!role) return;
    setIsLoading(true);
    try {
      const [accessResp, permsResp] = await Promise.all([
        accessApi.getAllAccessRights(),
        accessApi.getPermissionsForRole(role.name),
      ]);

      if (accessResp.data.data) {
        setAccessRights(accessResp.data.data);
      }

      if (permsResp.data) {
        setRolePerms(permsResp.data);
      }
    } catch (error) {
      console.error("Failed to fetch permissions data", error);
      toast.error("Failed to load permissions");
    } finally {
      setIsLoading(false);
    }
  }, [role]);

  useEffect(() => {
    if (role) {
      fetchData();
    }
  }, [role, fetchData]);

  const hasPermission = (path: string, method: string) => {
    return rolePerms.some(
      (p) => p[1] === path && p[2] === method.toUpperCase()
    );
  };

  const isGroupActive = (right: AccessRight) => {
    if (!right.endpoints || right.endpoints.length === 0) return false;
    return right.endpoints.every((e) => hasPermission(e.path, e.method));
  };

  const handleToggleGroup = async (right: AccessRight, active: boolean) => {
    if (!role || !right.endpoints || right.endpoints.length === 0) return;

    setIsProcessing(right.id);
    try {
      const promises = right.endpoints.map((e) => {
        if (active) {
          return accessApi.grantPermission(role.name, e.path, e.method);
        } else {
          return accessApi.revokePermission(role.name, e.path, e.method);
        }
      });

      await Promise.all(promises);
      toast.success(`${active ? "Granted" : "Revoked"} ${right.name}`);
      fetchData(); // Refresh permissions
    } catch (error) {
      console.error("Permission update failed", error);
      toast.error("Failed to update permissions");
    } finally {
      setIsProcessing(null);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-4 py-4">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    );
  }

  return (
    <ScrollArea className="-mx-2 flex-1 px-2">
      <div className="space-y-4 py-2">
        {accessRights.length === 0 ? (
          <div className="rounded-lg border-2 border-dashed py-12 text-center">
            <Icon
              name="Shield"
              className="text-muted-foreground/30 mx-auto mb-2 h-8 w-8"
            />
            <p className="text-muted-foreground text-sm">
              No access rights defined.
            </p>
          </div>
        ) : (
          accessRights.map((right: AccessRight) => (
            <div
              key={right.id}
              className="group hover:bg-muted/30 rounded-lg border p-4 transition-colors"
            >
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{right.name}</span>
                    <Badge variant="outline" className="text-[10px] uppercase">
                      {right.endpoints?.length || 0} ENDPOINTS
                    </Badge>
                  </div>
                  <p className="text-muted-foreground text-xs">
                    {right.description}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  {isProcessing === right.id && (
                    <Icon
                      name="Loader"
                      className="text-muted-foreground h-3 w-3 animate-spin"
                    />
                  )}
                  <Switch
                    checked={isGroupActive(right)}
                    onCheckedChange={(checked: boolean) =>
                      handleToggleGroup(right, checked)
                    }
                    disabled={
                      !!isProcessing || role?.name === "role:superadmin"
                    }
                  />
                </div>
              </div>
              {right.endpoints && right.endpoints.length > 0 && (
                <div className="mt-3 grid grid-cols-1 gap-1 border-t pt-3">
                  {right.endpoints.map((e: Endpoint) => (
                    <div
                      key={e.id}
                      className="flex items-center justify-between text-[11px]"
                    >
                      <code className="text-muted-foreground">{e.path}</code>
                      <Badge
                        variant={
                          hasPermission(e.path, e.method)
                            ? "default"
                            : "secondary"
                        }
                        className="h-4 px-1 text-[9px]"
                      >
                        {e.method}
                      </Badge>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </ScrollArea>
  );
}
