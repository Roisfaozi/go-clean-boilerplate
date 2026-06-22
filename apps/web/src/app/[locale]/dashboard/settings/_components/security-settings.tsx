"use client";

import { useUserSettings } from "./user-settings-context";
import { SecurityForm } from "~/components/dashboard/security-form";
import { EmailVerificationBanner } from "~/components/dashboard/email-verification-banner";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";

export function SecuritySettings() {
  const { user } = useUserSettings();

  return (
    <div className="space-y-4">
      <EmailVerificationBanner
        isVerified={!!user?.emailVerifiedAt}
        email={user?.email || ""}
      />
      <Card>
        <CardHeader>
          <CardTitle>Security Settings</CardTitle>
          <CardDescription>
            Change your password and manage security preferences.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="max-w-2xl">
            <SecurityForm />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
