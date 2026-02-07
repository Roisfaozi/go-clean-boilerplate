"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Checkbox } from "~/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/components/ui/dialog";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { ScrollArea } from "~/components/ui/scroll-area";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { accessApi, AccessRight, Endpoint } from "~/lib/api/access";

export default function AccessRightsPage() {
  const [accessRights, setAccessRights] = useState<AccessRight[]>([]);
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Create Access Right state
  const [newArName, setNewArName] = useState("");
  const [newArDesc, setNewArDesc] = useState("");
  const [isArDialogOpen, setIsArDialogOpen] = useState(false);

  // Create Endpoint state
  const [newEpPath, setNewEpPath] = useState("");
  const [newEpMethod, setNewEpMethod] = useState("GET");
  const [isEpDialogOpen, setIsEpDialogOpen] = useState(false);

  // Link state
  const [selectedAr, setSelectedAr] = useState<AccessRight | null>(null);
  const [isLinkDialogOpen, setIsLinkDialogOpen] = useState(false);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [arResp, epResp] = await Promise.all([
        accessApi.getAllAccessRights(),
        accessApi.searchEndpoints({ page: 1, page_size: 1000 }),
      ]);
      if (arResp && arResp.data) setAccessRights(arResp.data);
      if (epResp && epResp.data) setEndpoints(epResp.data);
    } catch (error) {
      toast.error("Failed to fetch data");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreateAr = async () => {
    try {
      await accessApi.createAccessRight(newArName, newArDesc);
      toast.success("Access Right created");
      setIsArDialogOpen(false);
      setNewArName("");
      setNewArDesc("");
      fetchData();
    } catch (error) {
      toast.error("Failed to create Access Right");
    }
  };

  const handleCreateEp = async () => {
    try {
      await accessApi.createEndpoint(newEpMethod, newEpPath);
      toast.success("Endpoint created");
      setIsEpDialogOpen(false);
      setNewEpPath("");
      setNewEpMethod("GET");
      fetchData();
    } catch (error) {
      toast.error("Failed to create Endpoint");
    }
  };

  const handleDeleteAr = async (id: string) => {
    try {
      await accessApi.deleteAccessRight(id);
      toast.success("Access Right deleted");
      fetchData();
    } catch (error) {
      toast.error("Failed to delete Access Right");
    }
  };

  const handleDeleteEp = async (id: string) => {
    try {
      await accessApi.deleteEndpoint(id);
      toast.success("Endpoint deleted");
      fetchData();
    } catch (error) {
      toast.error("Failed to delete Endpoint");
    }
  };

  const handleLink = async (endpointId: string) => {
    if (!selectedAr) return;
    try {
      await accessApi.linkEndpoint(selectedAr.id, endpointId);
      toast.success("Endpoint linked");
      fetchData(); // Refresh to update mappings
    } catch (error) {
      toast.error("Failed to link endpoint");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">
            Access Rights & Endpoints
          </h2>
          <p className="text-muted-foreground">
            Define resource groups and register API endpoints.
          </p>
        </div>
      </div>

      <Tabs defaultValue="access-rights" className="w-full">
        <TabsList className="grid w-full max-w-[400px] grid-cols-2">
          <TabsTrigger value="access-rights">Access Rights</TabsTrigger>
          <TabsTrigger value="endpoints">All Endpoints</TabsTrigger>
        </TabsList>

        <TabsContent value="access-rights" className="mt-4">
          <div className="mb-4 flex justify-end">
            <Dialog open={isArDialogOpen} onOpenChange={setIsArDialogOpen}>
              <DialogTrigger asChild>
                <Button size="sm">
                  <Icon name="Plus" className="mr-2 h-4 w-4" />
                  New Access Right
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Create Access Right</DialogTitle>
                  <DialogDescription>
                    Grouping endpoints makes it easier to manage permissions.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Name</Label>
                    <Input
                      id="name"
                      value={newArName}
                      onChange={(e) => setNewArName(e.target.value)}
                      placeholder="e.g. User Management"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="desc">Description</Label>
                    <Input
                      id="desc"
                      value={newArDesc}
                      onChange={(e) => setNewArDesc(e.target.value)}
                      placeholder="Manage all user related operations"
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button
                    variant="outline"
                    onClick={() => setIsArDialogOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button onClick={handleCreateAr}>Create</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {accessRights && accessRights.length > 0 && accessRights.map((ar) => (
              <Card key={ar.id}>
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">{ar.name}</CardTitle>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDeleteAr(ar.id)}
                    >
                      <Icon
                        name="Trash2"
                        className="text-destructive h-4 w-4"
                      />
                    </Button>
                  </div>
                  <CardDescription>
                    {ar.description || "No description"}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="text-muted-foreground flex items-center justify-between text-xs font-semibold tracking-wider uppercase">
                      Linked Endpoints ({ar.endpoints?.length || 0})
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 text-[10px]"
                        onClick={() => {
                          setSelectedAr(ar);
                          setIsLinkDialogOpen(true);
                        }}
                      >
                        Manage
                      </Button>
                    </div>
                    <ScrollArea className="h-[120px] pr-4">
                      <div className="space-y-1">
                        {ar.endpoints?.map((ep) => (
                          <div
                            key={ep.id}
                            className="hover:bg-muted/50 flex items-center gap-2 rounded p-1 text-xs"
                          >
                            <Badge
                              variant="outline"
                              className="h-4 px-1 font-mono text-[8px]"
                            >
                              {ep.method}
                            </Badge>
                            <span className="flex-1 truncate font-mono">
                              {ep.path}
                            </span>
                          </div>
                        ))}
                        {(!ar.endpoints || ar.endpoints.length === 0) && (
                          <div className="text-muted-foreground py-4 text-center text-xs italic">
                            No endpoints linked.
                          </div>
                        )}
                      </div>
                    </ScrollArea>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="endpoints" className="mt-4">
          <div className="mb-4 flex justify-end">
            <Dialog open={isEpDialogOpen} onOpenChange={setIsEpDialogOpen}>
              <DialogTrigger asChild>
                <Button size="sm">
                  <Icon name="Plus" className="mr-2 h-4 w-4" />
                  Register Endpoint
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Register API Endpoint</DialogTitle>
                  <DialogDescription>
                    Add a new endpoint to the system catalog.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="path">Path</Label>
                    <Input
                      id="path"
                      value={newEpPath}
                      onChange={(e) => setNewEpPath(e.target.value)}
                      placeholder="/api/v1/users"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="method">Method</Label>
                    <select
                      className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-md border px-3 py-2 text-sm file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
                      value={newEpMethod}
                      onChange={(e) => setNewEpMethod(e.target.value)}
                    >
                      <option value="GET">GET</option>
                      <option value="POST">POST</option>
                      <option value="PUT">PUT</option>
                      <option value="PATCH">PATCH</option>
                      <option value="DELETE">DELETE</option>
                    </select>
                  </div>
                </div>
                <DialogFooter>
                  <Button
                    variant="outline"
                    onClick={() => setIsEpDialogOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button onClick={handleCreateEp}>Register</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Method</TableHead>
                  <TableHead>Path</TableHead>
                  <TableHead>Created At</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {endpoints.map((ep) => (
                  <TableRow key={ep.id}>
                    <TableCell>
                      <Badge variant="outline" className="font-mono">
                        {ep.method}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {ep.path}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-xs">
                      {new Date(ep.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDeleteEp(ep.id)}
                      >
                        <Icon
                          name="Trash2"
                          className="text-destructive h-4 w-4"
                        />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </TabsContent>
      </Tabs>

      {/* Link Endpoint Dialog */}
      <Dialog open={isLinkDialogOpen} onOpenChange={setIsLinkDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Manage Endpoints for {selectedAr?.name}</DialogTitle>
            <DialogDescription>
              Select endpoints to include in this Access Right group.
            </DialogDescription>
          </DialogHeader>
          <ScrollArea className="mt-4 h-[300px] rounded-md border p-4">
            <div className="space-y-4">
              {endpoints.map((ep) => {
                const isLinked = selectedAr?.endpoints?.some(
                  (e) => e.id === ep.id
                );
                return (
                  <div key={ep.id} className="flex items-center space-x-2">
                    <Checkbox
                      id={`ep-${ep.id}`}
                      checked={isLinked}
                      onCheckedChange={() => handleLink(ep.id)}
                    />
                    <label
                      htmlFor={`ep-${ep.id}`}
                      className="flex items-center gap-2 text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                    >
                      <Badge variant="outline" className="h-4 px-1 text-[10px]">
                        {ep.method}
                      </Badge>
                      <span className="font-mono">{ep.path}</span>
                    </label>
                  </div>
                );
              })}
            </div>
          </ScrollArea>
          <DialogFooter>
            <Button onClick={() => setIsLinkDialogOpen(false)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
