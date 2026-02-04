import { type Metadata } from "next";
import { getCurrentSession } from "~/lib/server/auth/session";
import { ProfileForm } from "~/components/dashboard/profile-form";
import { SecurityForm } from "~/components/dashboard/security-form";
import { EmailVerificationBanner } from "~/components/dashboard/email-verification-banner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";

export const metadata: Metadata = {
  title: "Settings",
  description: "Manage your account settings and preferences.",
};

export default async function SettingsPage() {
  const { user } = await getCurrentSession();

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground">
          Manage your account settings and preferences.
        </p>
      </div>

      <Tabs defaultValue="profile" className="space-y-4">
        <TabsList className="bg-muted/50 p-1">
          <TabsTrigger value="profile" className="gap-2">
            <Icon name="User" className="h-4 w-4" />
            Profile
          </TabsTrigger>
          <TabsTrigger value="security" className="gap-2">
            <Icon name="Lock" className="h-4 w-4" />
            Security
          </TabsTrigger>
          <TabsTrigger value="preferences" className="gap-2">
            <Icon name="Settings2" className="h-4 w-4" />
            Preferences
          </TabsTrigger>
        </TabsList>

        <TabsContent value="profile">
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
        </TabsContent>

        <TabsContent value="security" className="space-y-4">
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
        </TabsContent>

        <TabsContent value="preferences">
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
        </TabsContent>
      </Tabs>
    </div>
  );
}
