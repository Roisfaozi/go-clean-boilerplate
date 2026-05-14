import { type Metadata } from "next";
import { getCurrentSession } from "~/lib/server/auth/session";
import { ProfileProvider } from "./_components/profile-context";
import { ProfileHeader } from "./_components/profile-header";
import { ProfileContent } from "./_components/profile-content";

export const metadata: Metadata = {
  title: "Profile",
  description: "Manage your profile information",
};

export default async function ProfilePage() {
  const { user } = await getCurrentSession();

  return (
    <ProfileProvider user={user}>
      <div className="space-y-6">
        <ProfileHeader />
        <ProfileContent />
      </div>
    </ProfileProvider>
  );
}
