import { AuthGuard } from "@/components/auth/auth-guard";
import { AppLayout } from "@/components/layout/app-layout";

export default function AppLayoutRoute() {
  return (
    <AuthGuard>
      <AppLayout />
    </AuthGuard>
  );
}
