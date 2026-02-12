"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Button } from "~/components/ui/button";

export function PreferencesSettings() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>App Preferences</CardTitle>
        <CardDescription>
          Customize your interface and density settings.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex max-w-2xl flex-col gap-6 py-4">
          <div className="flex items-center justify-between border-b pb-4">
            <div className="space-y-0.5">
              <div className="text-sm font-medium">Interface Density</div>
              <div className="text-muted-foreground text-xs">
                Choose between Comfort and Compact modes.
              </div>
            </div>
            <div className="bg-muted flex rounded-md p-1">
              <Button
                variant="ghost"
                size="sm"
                className="bg-background h-7 px-3 text-xs shadow-sm"
              >
                Comfort
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 px-3 text-xs"
              >
                Compact
              </Button>
            </div>
          </div>

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <div className="text-sm font-medium">
                Email Notifications
              </div>
              <div className="text-muted-foreground text-xs">
                Receive security alerts via email.
              </div>
            </div>
            <div className="bg-primary relative h-6 w-11 rounded-full">
              <div className="absolute top-1 right-1 h-4 w-4 rounded-full bg-white shadow-sm" />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
