"use client";

import { useUserSettings } from "./user-settings-context";
import { ProfileForm } from "~/components/dashboard/profile-form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";

export function ProfileSettings() {
  const { user } = useUserSettings();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Profile Information</CardTitle>
        <CardDescription>
          Update your account profile details and avatar.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="max-w-2xl">
          <ProfileForm user={user} />
        </div>
      </CardContent>
    </Card>
  );
}
