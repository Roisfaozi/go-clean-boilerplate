"use client";

import { LogOutIcon } from "lucide-react";
import { Button } from "../ui/button";
import { authApi } from "~/lib/api/auth";
import { toast } from "~/hooks/use-toast";

export default function LogoutButton({ className }: { className?: string }) {
  return (
    <div className={className}>
      <Button
        type="submit"
        onClick={async () => {
          try {
            await authApi.logout();
            window.location.href = "/login";
          } catch (error) {
            toast({
              title: "Logout failed",
              variant: "destructive",
            });
          }
        }}
        variant="destructive"
      >
        <LogOutIcon className="mr-2 h-4 w-4" />
        <span>Log out</span>
      </Button>
    </div>
  );
}
