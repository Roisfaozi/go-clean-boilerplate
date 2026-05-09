"use client";

import { useState } from "react";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { authApi } from "~/lib/api/auth";
import { toast } from "sonner";
import { Alert, AlertDescription, AlertTitle } from "~/components/ui/alert";

interface EmailVerificationBannerProps {
  isVerified: boolean;
  email: string;
}

export function EmailVerificationBanner({
  isVerified,
  email,
}: EmailVerificationBannerProps) {
  const [isLoading, setIsLoading] = useState(false);

  if (isVerified) {
    return (
      <div className="flex items-center gap-2 rounded-md border border-emerald-100 bg-emerald-50 p-3 text-sm text-emerald-600 dark:border-emerald-900/30 dark:bg-emerald-950/20">
        <Icon name="CircleCheck" className="h-4 w-4" />
        <span>
          Your email <strong>{email}</strong> is verified.
        </span>
      </div>
    );
  }

  const handleResend = async () => {
    setIsLoading(true);
    try {
      await authApi.resendVerification();
      toast.success("Verification email sent", {
        description: "Please check your inbox for the link.",
      });
    } catch (error: any) {
      toast.error(error.message || "Failed to resend verification email");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Alert
      variant="destructive"
      className="bg-destructive/5 border-destructive/20 text-destructive"
    >
      <Icon name="TriangleAlert" className="h-4 w-4" />
      <AlertTitle>Email not verified</AlertTitle>
      <AlertDescription className="mt-2 flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <span>
          Please verify your email <strong>{email}</strong> to access all
          features.
        </span>
        <Button
          variant="outline"
          size="sm"
          className="border-destructive/30 hover:bg-destructive/10 hover:text-destructive h-8"
          onClick={handleResend}
          disabled={isLoading}
        >
          {isLoading ? (
            <Icon name="Loader" className="mr-2 h-3 w-3 animate-spin" />
          ) : (
            <Icon name="Mail" className="mr-2 h-3 w-3" />
          )}
          Resend Link
        </Button>
      </AlertDescription>
    </Alert>
  );
}
