"use client";

import { useProfile } from "./profile-context";
import { ProfileForm } from "~/components/dashboard/profile-form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";

export function ProfileContent() {
  const { user } = useProfile();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Personal Information</CardTitle>
        <CardDescription>
          Update your display name and email address.
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
