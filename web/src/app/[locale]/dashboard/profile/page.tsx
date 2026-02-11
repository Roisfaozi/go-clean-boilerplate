import { type Metadata } from "next";
import { getCurrentSession } from "~/lib/server/auth/session";
import { ProfileForm } from "~/components/dashboard/profile-form";

export const metadata: Metadata = {
  title: "Profile",
  description: "Manage your profile information",
};

export default async function ProfilePage() {
  const { user } = await getCurrentSession();

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-medium">Profile</h3>
        <p className="text-muted-foreground text-sm">
          Manage your personal information and avatar.
        </p>
      </div>
      <div className="max-w-2xl">
        <ProfileForm user={user} />
      </div>
    </div>
  );
}
