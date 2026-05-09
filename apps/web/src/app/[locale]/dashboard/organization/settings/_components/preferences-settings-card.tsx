"use client";

import { useSettings } from "./settings-context";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Label } from "~/components/ui/label";
import { Switch } from "~/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import { Separator } from "~/components/ui/separator";

export function PreferencesSettingsCard() {
  const { settings, updateSetting } = useSettings();

  return (
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
              Force all members to enable Multi-Factor Authentication to access
              this organization.
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
  );
}
