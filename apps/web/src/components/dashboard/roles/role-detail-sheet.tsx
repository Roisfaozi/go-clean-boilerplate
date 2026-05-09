"use client";

import React, { useCallback, useEffect, useState } from "react";
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
import { accessApi, RoleAccessRightStatus } from "~/lib/api/access";
import { Role } from "~/lib/api/roles";
import { User, usersApi } from "~/lib/api/users";
import { useOrganizationStore } from "~/stores/use-organization-store";

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
  const { currentOrganization } = useOrganizationStore();
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
      const domain = currentOrganization?.slug || "global";
      const resp = await accessApi.getUsersForRole(role.name);
      // Note: Backend getUsersForRole might need domain too, checking accessApi
      const userIds = resp.data || [];

      if (userIds.length === 0) {
        setMembers([]);
        return;
      }

      const userPromises = userIds.slice(0, 50).map((id) =>
        usersApi.getById(id).catch((err) => {
          console.warn(`Failed to fetch user ${id}`, err);
          return null;
        })
      );
      const userResps = await Promise.all(userPromises);
      setMembers(
        userResps
          .filter((r): r is NonNullable<typeof r> => r !== null)
          .map((r) => r.data)
          .filter(Boolean) as User[]
      );
    } catch (error) {
      console.error("Failed to fetch role members", error);
      toast.error("Failed to load members");
    } finally {
      setIsLoading(false);
    }
  }, [role, currentOrganization]);

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
      const domain = currentOrganization?.slug || "global";
      await accessApi.assignRole(user.id, role.name, domain);
      toast.success(`${user.username} added to ${role.name} (${domain})`);
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
      const domain = currentOrganization?.slug || "global";
      await accessApi.revokeRole(userId, role.name, domain);
      toast.success(`${username} removed from ${role.name} (${domain})`);
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
                          <SearchUserItem
                            key={user.id}
                            user={user}
                            onSelect={addMember}
                          />
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
                    <RoleMemberItem
                      key={member.id}
                      member={member}
                      isProcessing={isProcessing === member.id}
                      onRemove={removeMember}
                    />
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
  const { currentOrganization } = useOrganizationStore();
  const [accessRights, setAccessRights] = useState<RoleAccessRightStatus[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isProcessing, setIsProcessing] = useState<string | null>(null);
  const [selectedDomain, setSelectedDomain] = useState<string>(
    currentOrganization?.slug || "global"
  );

  const fetchData = useCallback(async () => {
    if (!role) return;
    setIsLoading(true);
    try {
      const resp = await accessApi.getRoleAccessRights(
        role.name,
        selectedDomain
      );
      if (resp.data) {
        setAccessRights(resp.data);
      }
    } catch (error) {
      console.error("Failed to fetch permissions data", error);
      toast.error("Failed to load permissions");
    } finally {
      setIsLoading(false);
    }
  }, [role, selectedDomain]);

  useEffect(() => {
    if (role) {
      fetchData();
    }
  }, [role, fetchData]);

  const handleToggleGroup = async (
    right: RoleAccessRightStatus,
    active: boolean
  ) => {
    if (!role) return;

    setIsProcessing(right.id);
    try {
      if (active) {
        await accessApi.assignAccessRight(role.name, right.id, selectedDomain);
      } else {
        await accessApi.revokeAccessRight(role.name, right.id, selectedDomain);
      }

      toast.success(
        `${active ? "Assigned" : "Revoked"} ${right.name} in ${selectedDomain}`
      );
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
      <div className="mb-4 flex flex-col gap-2 rounded-md border p-3">
        <label className="text-muted-foreground text-xs font-semibold uppercase">
          Domain Context
        </label>
        <div className="flex items-center gap-2">
          <select
            className="border-input ring-offset-background placeholder:text-muted-foreground focus:ring-ring flex h-9 w-full items-center justify-between rounded-md border bg-transparent px-3 py-2 text-sm whitespace-nowrap shadow-sm focus:ring-1 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
            value={selectedDomain}
            onChange={(e) => setSelectedDomain(e.target.value)}
            disabled={isLoading || !!isProcessing}
          >
            <option value="global">
              Global Domain (Superadmin / System Wide)
            </option>
            {currentOrganization && (
              <option value={currentOrganization.slug}>
                {currentOrganization.name} ({currentOrganization.slug})
              </option>
            )}
          </select>
        </div>
      </div>

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
          accessRights.map((right: RoleAccessRightStatus) => (
            <div
              key={right.id}
              className="group hover:bg-muted/30 rounded-lg border p-4 transition-colors"
            >
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{right.name}</span>
                    <Badge
                      variant={
                        right.is_assigned
                          ? "default"
                          : right.is_partial
                            ? "secondary"
                            : "outline"
                      }
                      className="text-[10px] uppercase"
                    >
                      {right.is_assigned
                        ? "Assigned"
                        : right.is_partial
                          ? "Partial"
                          : `${right.endpoints?.length || 0} ENDPOINTS`}
                    </Badge>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {isProcessing === right.id && (
                    <Icon
                      name="Loader"
                      className="text-muted-foreground h-3 w-3 animate-spin"
                    />
                  )}
                  <Switch
                    checked={right.is_assigned}
                    onCheckedChange={(checked: boolean) =>
                      handleToggleGroup(right, checked)
                    }
                    disabled={!!isProcessing || role?.name.startsWith("role:")}
                  />
                </div>
              </div>
              {right.endpoints && right.endpoints.length > 0 && (
                <div className="mt-3 grid grid-cols-1 gap-1 border-t pt-3">
                  {right.endpoints.map((eStr, idx) => {
                    // format eStr: "GET /api/v1/users"
                    const [method, ...pathParts] = eStr.split(" ");
                    const path = pathParts.join(" ");
                    return (
                      <div
                        key={`${right.id}-ep-${idx}`}
                        className="flex items-center justify-between text-[11px]"
                      >
                        <code className="text-muted-foreground">{path}</code>
                        <Badge
                          variant={
                            right.is_assigned || right.is_partial
                              ? "default"
                              : "secondary"
                          }
                          className="h-4 px-1 text-[9px]"
                        >
                          {method}
                        </Badge>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </ScrollArea>
  );
}

const SearchUserItem = React.memo(function SearchUserItem({
  user,
  onSelect,
}: {
  user: User;
  onSelect: (user: User) => void;
}) {
  return (
    <CommandItem
      onSelect={() => onSelect(user)}
      className="flex cursor-pointer items-center gap-2"
    >
      <Avatar className="h-6 w-6">
        <AvatarImage src={user.avatar_url} />
        <AvatarFallback>{user.username[0].toUpperCase()}</AvatarFallback>
      </Avatar>
      <div className="flex flex-col">
        <span className="text-sm font-medium">{user.username}</span>
        <span className="text-muted-foreground text-[10px]">{user.email}</span>
      </div>
      <Icon name="Plus" className="text-muted-foreground ml-auto h-3 w-3" />
    </CommandItem>
  );
});

const RoleMemberItem = React.memo(function RoleMemberItem({
  member,
  isProcessing,
  onRemove,
}: {
  member: User;
  isProcessing: boolean;
  onRemove: (id: string, username: string) => void;
}) {
  return (
    <div className="hover:bg-muted/50 group flex items-center gap-3 rounded-md p-2 transition-colors">
      <Avatar className="h-9 w-9 border">
        <AvatarImage src={member.avatar_url} />
        <AvatarFallback>{member.username[0].toUpperCase()}</AvatarFallback>
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
        disabled={isProcessing}
        onClick={() => onRemove(member.id, member.username)}
      >
        {isProcessing ? (
          <Icon name="Loader" className="h-4 w-4 animate-spin" />
        ) : (
          <Icon name="UserMinus" className="h-4 w-4" />
        )}
      </Button>
    </div>
  );
});
