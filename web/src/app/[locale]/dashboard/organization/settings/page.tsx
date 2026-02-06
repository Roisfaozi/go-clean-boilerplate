"use client";

import { useState, useEffect } from "react";
import { useOrganizationStore } from "~/stores/use-organization-store";
import {
  organizationsApi,
  OrganizationSettings,
} from "~/lib/api/organizations";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { toast } from "sonner";
import { Separator } from "~/components/ui/separator";
import { useRouter } from "next/navigation";
import { Switch } from "~/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";

export default function OrganizationSettingsPage() {
  const { currentOrganization, setCurrentOrganization } =
    useOrganizationStore();
  const [name, setName] = useState(currentOrganization?.name || "");
  const [settings, setSettings] = useState<OrganizationSettings>(
    currentOrganization?.settings || {}
  );
  const [isLoading, setIsLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const router = useRouter();

  useEffect(() => {
    if (currentOrganization) {
      setName(currentOrganization.name);
      setSettings(currentOrganization.settings || {});
    }
  }, [currentOrganization]);

  const handleUpdate = async () => {
    if (!currentOrganization) return;
    setIsLoading(true);
    try {
      const resp = await organizationsApi.update(currentOrganization.id, {
        name,
        settings,
      });
      if (resp.data) {
        setCurrentOrganization(resp.data);
        toast.success("Organization updated successfully");
      }
    } catch (error: any) {
      toast.error(error.message || "Failed to update organization");
    } finally {
      setIsLoading(false);
    }
  };

  const updateSetting = (key: keyof OrganizationSettings, value: any) => {
    setSettings((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const handleDelete = async () => {
    if (!currentOrganization) return;
    if (
      !confirm(
        `Are you sure you want to permanently delete ${currentOrganization.name}? This action cannot be undone.`
      )
    )
      return;

    setIsDeleting(true);
    try {
      await organizationsApi.delete(currentOrganization.id);
      toast.success("Organization deleted successfully");
      setCurrentOrganization(null);
      router.push("/dashboard");
    } catch (error: any) {
      toast.error(error.message || "Failed to delete organization");
    } finally {
      setIsDeleting(false);
    }
  };

  if (!currentOrganization) {
    return (
      <div className="flex h-[400px] items-center justify-center rounded-lg border-2 border-dashed">
        <p className="text-muted-foreground">No organization selected.</p>
      </div>
    );
  }

  const hasChanges =
    name !== currentOrganization.name ||
    JSON.stringify(settings) !==
      JSON.stringify(currentOrganization.settings || {});

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">
            Organization Settings
          </h2>
          <p className="text-muted-foreground">
            Update your organization profile and general settings.
          </p>
        </div>
        <Button onClick={handleUpdate} disabled={isLoading || !hasChanges}>
          {isLoading && (
            <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
          )}
          Save All Changes
        </Button>
      </div>

      <div className="grid max-w-3xl gap-6">
        <Card>
          <CardHeader>
            <CardTitle>General Information</CardTitle>
            <CardDescription>
              The display name and identity of your organization.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Organization Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Acme Corp"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="slug">Organization Slug (URL)</Label>
              <Input
                id="slug"
                value={currentOrganization.slug}
                disabled
                className="bg-muted font-mono text-xs"
              />
              <p className="text-muted-foreground text-[10px] italic">
                Slug cannot be changed after creation.
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Preferences & Security</CardTitle>
            <CardDescription>
              Configure organization-wide behavior and security policies.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="flex items-center justify-between space-x-2">
              <div className="flex flex-col space-y-1">
                <Label htmlFor="theme">Default Theme</Label>
                <p className="text-muted-foreground text-xs">
                  The default visual theme for all members of this organization.
                </p>
              </div>
              <Select
                value={settings.theme || "system"}
                onValueChange={(v) => updateSetting("theme", v)}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Select theme" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">Light</SelectItem>
                  <SelectItem value="dark">Dark</SelectItem>
                  <SelectItem value="system">System</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <Separator />

            <div className="flex items-center justify-between space-x-2">
              <div className="flex flex-col space-y-1">
                <Label htmlFor="mfa">Require MFA</Label>
                <p className="text-muted-foreground text-xs">
                  Force all members to enable Multi-Factor Authentication to
                  access this organization.
                </p>
              </div>
              <Switch
                id="mfa"
                checked={settings.mfa_required || false}
                onCheckedChange={(checked) =>
                  updateSetting("mfa_required", checked)
                }
              />
            </div>
          </CardContent>
        </Card>

        <Card className="border-destructive/20 bg-destructive/5">
          <CardHeader>
            <CardTitle className="text-destructive">Danger Zone</CardTitle>
            <CardDescription>
              Permanently delete this organization and all associated data.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground text-sm">
              Once you delete an organization, there is no going back. Please be
              certain.
            </p>
          </CardContent>
          <CardFooter className="border-destructive/10 flex items-center justify-between border-t px-6 py-4">
            <div className="text-muted-foreground text-xs italic">
              Owned by{" "}
              {currentOrganization.owner_id === "me" ? "you" : "another admin"}
            </div>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? (
                <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Icon name="Trash2" className="mr-2 h-4 w-4" />
              )}
              Delete Organization
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
