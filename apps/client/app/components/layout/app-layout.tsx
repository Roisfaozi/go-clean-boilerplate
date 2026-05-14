import { Outlet } from "react-router";
import { AppSidebar } from "./app-sidebar";
import { AppNavbar } from "./app-navbar";
import { AppBreadcrumb } from "./app-breadcrumb";
import { AppCommandPalette } from "@/components/navigation/app-command-palette";
import { useRealtimeInit } from "@/hooks/use-realtime";
import { UploadManager } from "@/components/upload/upload-manager";
import { UploadDuplicateDialog } from "@/components/upload/upload-duplicate-dialog";
import { useUploadSideEffects } from "@/lib/upload/use-upload-side-effects";

export function AppLayout() {
  useRealtimeInit();
  useUploadSideEffects();

  return (
    <div className="bg-background flex min-h-screen w-full">
      <AppSidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <AppNavbar />
        <main className="animate-fade-in flex-1 overflow-auto p-6">
          <AppBreadcrumb />
          <Outlet />
        </main>
      </div>
      <AppCommandPalette />
      <UploadManager />
      <UploadDuplicateDialog />
    </div>
  );
}
